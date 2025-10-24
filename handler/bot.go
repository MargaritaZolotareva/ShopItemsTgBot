package handler

import (
	"VkScraper/internal/file_formatter"
	"VkScraper/service"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var ctx = context.Background()

func StartBotWebhook(db *gorm.DB, bot *tgbotapi.BotAPI, redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var updates tgbotapi.Update
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			log.Println("Ошибка при декодировании webhook:", err)
			return
		}
		if !isAccessAllowed(updates) {
			return
		}

		if updates.CallbackQuery != nil {
			handleCallback(bot, updates, redisClient)
		}
		if updates.Message != nil {
			handleMessage(bot, updates, db, redisClient)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func getAllowedUsers() map[int]bool {
	usersEnv := os.Getenv("ALLOWED_USERS")
	users := strings.Split(usersEnv, ",")
	allowedUsers := make(map[int]bool)
	for _, user := range users {
		if userId, err := strconv.Atoi(user); err == nil {
			allowedUsers[userId] = true
		}
	}

	return allowedUsers
}

func getUserFromId(updates tgbotapi.Update) (int, bool) {
	if updates.CallbackQuery != nil && updates.CallbackQuery.From != nil {
		return updates.CallbackQuery.From.ID, true
	}

	if updates.Message != nil && updates.Message.From != nil {
		return updates.Message.From.ID, true
	}

	return 0, false
}

func isAccessAllowed(updates tgbotapi.Update) bool {
	allowedUsers := getAllowedUsers()

	userFromId, found := getUserFromId(updates)
	if !found {
		log.Println("Пользователь не найден.")
		return false
	}

	if _, ok := allowedUsers[userFromId]; !ok {
		log.Println("Доступ для пользователя с ID ", userFromId, " запрещён")
		return false
	}
	return true
}

func handleCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update, redisClient *redis.Client) {
	state, err := redisClient.Get(ctx, fmt.Sprintf("state:%d", update.CallbackQuery.Message.Chat.ID)).Result()
	if err != nil && err != redis.Nil {
		log.Println("Ошибка при получении состояния из Redis:", err)
		return
	}

	switch state {
	case "awaiting_xlsx_file_platform":
		handleFilePlatform(bot, update, redisClient)
	default:
		if update.CallbackQuery.Data == "download_file" {
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Введите платформу для загрузки товаров:")
			replyKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Яндекс", "yandex"),
					tgbotapi.NewInlineKeyboardButtonData("2ГИС", "2gis"),
				),
			)

			msg.ReplyMarkup = replyKeyboard

			bot.Send(msg)
			redisClient.Set(ctx, fmt.Sprintf("state:%d", update.CallbackQuery.Message.Chat.ID), "awaiting_xlsx_file_platform", 0).Err()
		}
		if update.CallbackQuery.Data == "update_prices" {
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Введите стоимость порции гелия (в рублях):")
			bot.Send(msg)
			redisClient.Set(ctx, fmt.Sprintf("state:%d", update.CallbackQuery.Message.Chat.ID), "awaiting_helium_price", 0).Err()
		}
	}
}

func handleMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *gorm.DB, redisClient *redis.Client) {
	state, err := redisClient.Get(ctx, fmt.Sprintf("state:%d", update.Message.Chat.ID)).Result()
	if err != nil && err != redis.Nil {
		log.Println("Ошибка при получении состояния из Redis:", err)
		return
	}

	switch state {
	case "awaiting_helium_price":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Обновление цен запущено")
		bot.Send(msg)
		go handleHeliumPriceInput(bot, update, db, redisClient)
	default:
		if update.Message.Command() == "start" {
			msg := createStartMenu(update.Message.Chat.ID)
			bot.Send(msg)
		}
	}
}

func createStartMenu(chatId int64) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatId, "Выберите действие:")
	replyKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Обновить цены товаров", "update_prices"),
			tgbotapi.NewInlineKeyboardButtonData("Скачать файл товаров для сторонней платформы", "download_file"),
		),
	)
	msg.ReplyMarkup = replyKeyboard
	return msg
}

func handleHeliumPriceInput(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *gorm.DB, redisClient *redis.Client) {
	heliumPrice, err := strconv.Atoi(update.Message.Text)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка ввода. Пожалуйста, введите число.")
		bot.Send(msg)
		return
	}

	balloonService := &service.BalloonService{DB: db}
	vkService := &service.VKMarketService{
		AccessToken: os.Getenv("VK_ACCESS_TOKEN"),
		GroupID:     os.Getenv("VK_GROUP_ID"),
	}

	balloons, err := balloonService.GetAllProducts()
	if err != nil {
		log.Fatal("Ошибка получения товаров:", err)
	}

	limitPerSecond := 3
	requestsDone := 0
	for _, balloon := range balloons {
		newPrice := balloonService.CalculateNewPrice(balloon, heliumPrice)
		if requestsDone == limitPerSecond {
			time.Sleep(time.Second)
			requestsDone = 0
		}
		err := vkService.UpdateProductPrice(balloon.Sku, newPrice)
		requestsDone++
		if err != nil {
			log.Println("Ошибка обновления цены товара", balloon.ID, ":", err)
		}
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Цены успешно обновлены!")
	bot.Send(msg)
	redisClient.Del(ctx, fmt.Sprintf("state:%d", update.Message.Chat.ID))

	msg = createStartMenu(update.Message.Chat.ID)
	bot.Send(msg)
}

func handleFilePlatform(bot *tgbotapi.BotAPI, update tgbotapi.Update, redisClient *redis.Client) {
	format := update.CallbackQuery.Data
	formatter, err := file_formatter.NewFileFormatter(format)
	if err != nil {
		log.Println(err, format)
	}
	products := GetAllProductsFromVk()

	fileName, fileBytes, err := formatter.Generate(products)
	doc := tgbotapi.NewDocumentUpload(update.CallbackQuery.Message.Chat.ID, tgbotapi.FileBytes{
		Name:  fileName,
		Bytes: fileBytes,
	})
	bot.Send(doc)
	redisClient.Del(ctx, fmt.Sprintf("state:%d", update.CallbackQuery.Message.Chat.ID))

	msg := createStartMenu(update.CallbackQuery.Message.Chat.ID)
	bot.Send(msg)
}
