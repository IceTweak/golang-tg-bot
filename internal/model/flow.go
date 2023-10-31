package model

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

/*
CommandFlow (starting a specific script (flow) for manipulating an object, is a command)
		  |
		  Usecase (actions that can be performed on an object)
				|
				Chain (algorithm, sequence of steps to implement an action and obtain some result)
					|
					Step (certain, specific step, action)


An example of how the described flow looks like with one use case and one step:
{
	"pin":{
		"pinMessage":{
			"0":{
				"handler": HandlerFunc0(),
				"message":"Write a message to pin",
			},
			"1":{
				"handler": HandlerFunc1(),
				"message":"Write a chats IDs where messages will be pinned",
			}

		}
	}
}
*/

type (
	Case string
	Step int

	Message struct {
		Text    string
		Buttons []Button
	}
	Button struct {
		Name         string
		CallbackData CallbackData
	}
)

// Assemble a bot message from the described local model
func (msg Message) BuildBotMessage(chatID int64) tgbotapi.MessageConfig {
	replyMessage := tgbotapi.NewMessage(chatID, msg.Text)
	var buttonRows [][]tgbotapi.InlineKeyboardButton
	for _, button := range msg.Buttons {
		buttonRows = append(buttonRows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(button.Name, button.CallbackData.Encode()),
		),
		)
	}
	markup := tgbotapi.NewInlineKeyboardMarkup(
		buttonRows...,
	)
	replyMessage.ReplyMarkup = markup
	replyMessage.ParseMode = tgbotapi.ModeHTML
	return replyMessage
}
