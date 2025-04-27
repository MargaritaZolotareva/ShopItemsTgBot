package handler

import (
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

		if updates.Message != nil {
			handleMessage(bot, updates, db, redisClient)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func handleMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *gorm.DB, redisClient *redis.Client) {
	// Получаем состояние пользователя из Redis
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
		if update.Message.Command() == "download_from_vk" {
			fileName, fileBytes := GenerateXLSXHandler()
			doc := tgbotapi.NewDocumentUpload(update.Message.Chat.ID, tgbotapi.FileBytes{
				Name:  fileName,
				Bytes: fileBytes,
			})
			bot.Send(doc)
		}
		if update.Message.Command() == "update_prices" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите стоимость порции гелия (в рублях):")
			bot.Send(msg)
			redisClient.Set(ctx, fmt.Sprintf("state:%d", update.Message.Chat.ID), "awaiting_helium_price", 0).Err()
		}
	}
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
}
