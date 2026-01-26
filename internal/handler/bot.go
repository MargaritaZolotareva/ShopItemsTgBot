package handler

import (
	"ShopItemsTgBot/internal/file_formatter"
	"ShopItemsTgBot/internal/service"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func StartBotWebhook(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
	if token != os.Getenv("TG_BOT_WEBHOOK_SECRET") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		log.Println("Invalid secret token")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func CheckUserMiddleware(b *gotgbot.Bot, ctx *ext.Context) error {
	if !isAccessAllowed(ctx) {
		b.SendMessage(ctx.EffectiveMessage.Chat.Id, "Вы не авторизованы для использования этого бота.", nil)
		return fmt.Errorf("access denied for user %d", ctx.EffectiveSender.User.Id)
	}

	return nil
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

func isAccessAllowed(ctx *ext.Context) bool {
	allowedUsers := getAllowedUsers()

	userFromId := int(ctx.EffectiveSender.User.Id)
	if _, ok := allowedUsers[userFromId]; !ok {
		log.Println("Доступ для пользователя с ID ", userFromId, " запрещён")
		return false
	}
	return true
}

func (h *CustomHandler) SendRequestPriceMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	rCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx.EffectiveMessage.Reply(b, "Введите стоимость порции гелия (в рублях):", nil)
	h.RedisClient.Set(rCtx, fmt.Sprintf("state:%d", ctx.EffectiveMessage.Chat.Id), "awaiting_helium_price", 0).Err()
	return nil
}

func (h *CustomHandler) ShowFileTypes(b *gotgbot.Bot, ctx *ext.Context) error {
	if err := CheckUserMiddleware(b, ctx); err != nil {
		return nil
	}
	rCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	ctx.EffectiveMessage.Reply(b, "Введите платформу для загрузки товаров:", &gotgbot.SendMessageOpts{
		ParseMode: "html",
		ReplyMarkup: &gotgbot.ReplyKeyboardMarkup{
			Keyboard: [][]gotgbot.KeyboardButton{
				{
					{Text: "yandex"},
					{Text: "2gis"},
				},
			},
			ResizeKeyboard:  true,
			OneTimeKeyboard: true,
		},
	})
	h.RedisClient.Set(rCtx, fmt.Sprintf("state:%d", ctx.EffectiveMessage.Chat.Id), "awaiting_xlsx_file_platform", 0).Err()
	return nil
}

func HandleHeliumPriceInput(b *gotgbot.Bot, ctx *ext.Context, redisClient *redis.Client, db *gorm.DB) {
	heliumPrice, err := strconv.Atoi(ctx.EffectiveMessage.Text)
	if err != nil {
		_, err := ctx.EffectiveMessage.Reply(b, "Ошибка ввода. Пожалуйста, введите число.", nil)
		if err != nil {
			log.Println("failed to send message: %w", err)
		}
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

	_, err = ctx.EffectiveMessage.Reply(b, "Цены успешно обновлены!", nil)
	if err != nil {
		log.Println("failed to send message: %w", err)
	}

	rCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	redisClient.Del(rCtx, fmt.Sprintf("state:%d", ctx.EffectiveMessage.Chat.Id))

	CreateStartMenu(b, ctx)
}

func isValidCallbackData(data string) bool {
	return data == "yandex" || data == "2gis"
}

func (h *CustomHandler) HandleFilePlatform(b *gotgbot.Bot, ctx *ext.Context) error {
	format := ctx.EffectiveMessage.Text

	if !isValidCallbackData(format) {
		log.Printf("Invalid callback data: %s", format)
		_, err := b.SendMessage(
			ctx.EffectiveMessage.Chat.Id,
			"Неправильный формат файла. Пожалуйста, выберите вариант из меню.",
			nil,
		)
		if err != nil {
			log.Println("Error sending message:", err)
		}
		return nil
	}

	formatter, err := file_formatter.NewFileFormatter(format)
	if err != nil {
		log.Println(err, format)
	}
	products := GetAllProductsFromVk()

	fileName, fileBytes, err := formatter.Generate(products)
	buf := bytes.NewBuffer(fileBytes)

	_, err = b.SendDocument(ctx.EffectiveChat.Id,
		gotgbot.InputFileByReader(fileName, buf),
		&gotgbot.SendDocumentOpts{
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId: ctx.EffectiveMessage.MessageId,
			},
		})
	if err != nil {
		log.Println("failed to send document: %w", err)
		return nil
	}

	rCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	h.RedisClient.Del(rCtx, fmt.Sprintf("state:%d", ctx.EffectiveMessage.Chat.Id))

	CreateStartMenu(b, ctx)

	return nil
}

func CreateStartMenu(b *gotgbot.Bot, ctx *ext.Context) error {
	if err := CheckUserMiddleware(b, ctx); err != nil {
		return nil
	}
	_, err := b.SendMessage(ctx.EffectiveMessage.Chat.Id, "Выберите действие:", &gotgbot.SendMessageOpts{
		ParseMode: "html",
		ReplyMarkup: &gotgbot.ReplyKeyboardMarkup{
			Keyboard: [][]gotgbot.KeyboardButton{{
				{Text: "Обновить цены товаров"},
				{Text: "Скачать файл товаров для загрузки"},
			}},
			ResizeKeyboard:  true,
			OneTimeKeyboard: true,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}
	return nil
}

type CustomHandler struct {
	RedisClient *redis.Client
	DB          *gorm.DB
}

func (h *CustomHandler) HandleMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	if err := CheckUserMiddleware(b, ctx); err != nil {
		return nil
	}
	rCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	state, err := h.RedisClient.Get(rCtx, fmt.Sprintf("state:%d", ctx.EffectiveMessage.Chat.Id)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Println("Ошибка при получении состояния из Redis:", err)
		return nil
	}

	switch state {
	case "awaiting_helium_price":
		_, err := ctx.EffectiveMessage.Reply(b, "Обновление цен запущено", nil)
		if err != nil {
			log.Println("failed to send message: %w", err)
		}
		go HandleHeliumPriceInput(b, ctx, h.RedisClient, h.DB)
	case "awaiting_xlsx_file_platform":
		h.HandleFilePlatform(b, ctx)
	default:
		if ctx.EffectiveMessage.Text == "Скачать файл товаров для загрузки" {
			h.ShowFileTypes(b, ctx)
		}
		if ctx.EffectiveMessage.Text == "Обновить цены товаров" {
			h.SendRequestPriceMessage(b, ctx)
		}
	}
	return nil
}
