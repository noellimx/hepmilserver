package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	TGBOT_TOKEN := os.Getenv("TGBOT_TOKEN")

	bot, err := tgbotapi.NewBotAPI(TGBOT_TOKEN)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true // Optional: shows all bot activity in logs

	log.Printf("Authorized on account %s", bot.Self.UserName)

	commands := []tgbotapi.BotCommand{
		{
			Command:     "start",
			Description: "Start interacting with the bot",
		},
		{
			Command:     "help",
			Description: "Get help and usage information",
		},
	}

	cfg := tgbotapi.NewSetMyCommands(commands...)
	_, err = bot.Request(cfg)
	if err != nil {
		log.Fatalf("failed to set bot commands: %v", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello, world!")
			bot.Send(msg)
		}
	}
}
