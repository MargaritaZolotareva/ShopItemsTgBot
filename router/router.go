package router

import (
	"VkScraper/handler"
	"github.com/go-redis/redis/v8"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gorm.io/gorm"
	"net/http"
)

func SetupRoutes(db *gorm.DB, bot *tgbotapi.BotAPI, redisClient *redis.Client) {
	http.HandleFunc("/", handler.StartBotWebhook(db, bot, redisClient))
}
