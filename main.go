package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	godotenv "github.com/joho/godotenv"
)

// TODO: ability to add media (photo, video, etc.) to pinned message in /pin command

// TODO: realize command palette (e.g. buttons with commands names)

// TODO: add description to commands

// TODO: fix "not a command" message after inline response

func initEnvVars() {

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	initEnvVars()
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))

	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil && update.InlineQuery != nil {
			HandleInlineMode(*bot, update)
		} else if update.Message != nil && update.Message.IsCommand() {

			switch update.Message.Command() {
			case "pin":
				PinMessage(*bot, update)
			default:
				SendMessage(*bot, update.FromChat().ID, "I don't know that command")
			}

		} else {
			SendMessage(*bot, update.FromChat().ID, "I can't understand you. Type /help to see all available commands")
		}
	}
}
