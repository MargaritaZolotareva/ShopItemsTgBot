package main

import (
	"ShopItemsTgBot/bot"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var ctx = context.Background()

func init() {
	if os.Getenv("GO_ENV") == "development" {
		err := godotenv.Load("/build/.env")
		if err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}
}

func InitDB() *gorm.DB {
	dsn := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
	)

	var err error
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Ошибка при соединении с БД: %v", err)
	}

	return db
}

func initRedis() *redis.Client {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Ошибка при подключении к Redis: %v", err)
	}
	return client
}

func main() {
	log.SetOutput(os.Stdout)
	db := InitDB()
	redisInst := initRedis()

	bot.StartBot(redisInst, db)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Graceful shutdown started")
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Closing redis")
	redisInst.Close()

	log.Println("Closing DB connection...")
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
	}

	log.Println("Shutdown complete")
}
