package model

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Local object with the necessary data to work on updates coming from Telegram
type UpdateLocal struct {
	TelegramUserID int64
	TelegramChatID int64
	CallbackData   CallbackData
}

// Decode the incoming update into the local model
func DecodeToLocal(upd tgbotapi.Update) *UpdateLocal {
	tgUser := upd.SentFrom()
	tgChat := upd.FromChat()
	var cData CallbackData
	if query := upd.CallbackQuery; query != nil {
		cDataBot := CallbackDataBot(upd.CallbackData())
		cData = *cDataBot.Decode()
	}
	return &UpdateLocal{
		TelegramUserID: tgUser.ID,
		TelegramChatID: tgChat.ID,
		CallbackData:   cData,
	}
}
