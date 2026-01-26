package bot

import (
	"ShopItemsTgBot/internal/handler"
	gotgbot "github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"log"
	"os"
)

func StartBot(redisInst *redis.Client, db *gorm.DB) {
	token := os.Getenv("TG_BOT_TOKEN")
	bot, err := gotgbot.NewBot(token, nil)
	if err != nil {
		log.Fatal("Error creating bot: ", err)
	}
	log.Printf("Authorized on account %s", bot.Username)
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Println("an error occurred while handling update:", err.Error())
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})

	h := &handler.CustomHandler{RedisClient: redisInst, DB: db}
	dispatcher.AddHandler(handlers.NewMessage(message.Text, h.HandleMessage))
	dispatcher.AddHandler(handlers.NewCommand("start", handler.CreateStartMenu))

	updater := ext.NewUpdater(dispatcher, &ext.UpdaterOpts{})
	webhookOpts := ext.WebhookOpts{
		ListenAddr:  "0.0.0.0:8080",
		SecretToken: os.Getenv("TG_BOT_WEBHOOK_SECRET"),
	}
	err = updater.StartWebhook(bot, "/"+token[:16], webhookOpts)
	if err != nil {
		panic("failed to start webhook: " + err.Error())
	}
	err = updater.SetAllBotWebhooks(os.Getenv("TG_BOT_WEBHOOK_DOMAIN"), &gotgbot.SetWebhookOpts{
		MaxConnections:     10,
		DropPendingUpdates: true,
		SecretToken:        webhookOpts.SecretToken,
	})
	if err != nil {
		panic("failed to set webhook: " + err.Error())
	}

	log.Printf("Bot has been started... %s", bot.Username)

	updater.Idle()
}
