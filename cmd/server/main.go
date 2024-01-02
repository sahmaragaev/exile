package main

import (
	"exile-telegram-bot/pkg/bot"
	"exile-telegram-bot/pkg/config"
	"exile-telegram-bot/pkg/db"
	"log"
)

func main() {
	err := config.LoadConfig("config/config.json")
    if err != nil {
        log.Fatalf("Error loading config file: %v", err)
    }

	db.Connect(config.AppConfig.MongoURI)
	db.InitializeCollections()
    if err := bot.StartBot(config.AppConfig.TelegramToken); err != nil {
        log.Fatalf("Failed to start the Telegram bot: %v", err)
    }
}