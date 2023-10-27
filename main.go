package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	godotenv "github.com/joho/godotenv"
)

/**
* TODO: /pin command need to take 2 arguments:
* 1) Message text that's need to be pinned to chat(-s)
* 2) List of chats:
*	TODO: how to handle chats (ids, links, etc.)?
Then it's needed to print chats where bot is not an admin
*/

// TODO: ability to add media (photo, video, etc.) to pinned message in /pin command

// TODO: realize command palette (e.g. buttons with commands names)

// TODO: add description to commands

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
	botChats := make(map[int64]int64)

	for update := range updates {
		if update.Message == nil && update.InlineQuery != nil {
			HandleInlineMode(*bot, update)
		} else if update.Message != nil && update.Message.IsCommand() &&
			CheckAdminRole(*bot, update.FromChat().ID) {

			switch update.Message.Command() {
			case "pin":
				PinMessage(*bot, update, botChats)
			default:
				SendMessage(*bot, update.FromChat().ID, "I don't know that command")
			}

		} else if update.Message != nil && len(update.Message.NewChatMembers) > 0 {
			chat := update.FromChat()
			log.Println("Bot is added to group")
			botChats[chat.ID] = chat.ID
		} else {
			log.Println("None of the message types handled!")
		}
	}
}
