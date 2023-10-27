package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SendMessage(bot tgbotapi.BotAPI, chatId int64, text string) {
	msg := tgbotapi.NewMessage(chatId, text)
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func HandleInlineMode(bot tgbotapi.BotAPI, update tgbotapi.Update) {
	// Parse Companies from file
	companies := ParseCompFromXml("companies.xml")

	// Process query
	query := update.InlineQuery.Query

	// Filter companies along query
	filteredCompanies := Filter(companies.Companies, func(comp Company) bool {
		return strings.Index(strings.ToLower(comp.Title), strings.ToLower(query)) >= 0
	})

	var articles []interface{}
	if len(filteredCompanies) == 0 {
		msg := tgbotapi.NewInlineQueryResultArticleMarkdown(update.InlineQuery.ID, "No one companies matches", "No one companies matches")
		articles = append(articles, msg)
	} else {
		var i = 0
		for _, comp := range filteredCompanies {
			text := fmt.Sprintf(
				"*%s*\n"+
					"*Category:* _%s_\n"+
					"*Year:* _%s_\n"+
					"*Owner:* _%s_\n"+
					"*Social links:* \n"+
					"%s",
				comp.Title,
				comp.Category,
				comp.Year,
				comp.Owner,
				strings.Join(MapLinks(comp.Links.Links), "\n"),
			)

			msg := tgbotapi.NewInlineQueryResultArticleMarkdown(comp.Title, comp.Title, text)
			articles = append(articles, msg)
			if i >= 5 {
				break
			}
		}
	}

	inlineConfig := tgbotapi.InlineConfig{
		InlineQueryID: update.InlineQuery.ID,
		IsPersonal:    true,
		CacheTime:     0,
		Results:       articles,
	}

	if _, err := bot.Request(inlineConfig); err != nil {
		log.Println(err)
	}
}

// Checks if the bot has an administrator role
func CheckAdminRole(bot tgbotapi.BotAPI, chatId int64) bool {
	chat, err := bot.GetChat(tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: chatId,
		},
	})

	if err != nil {
		log.Println(err)
	}

	isPrivGroupOrChan := (chat.IsGroup() || chat.IsChannel()) && chat.IsPrivate()
	isPubGroupOrChan := chat.IsGroup() || chat.IsChannel()

	if isPrivGroupOrChan || isPubGroupOrChan || chat.IsSuperGroup() {
		if botAsMember, err := bot.GetChatMember(tgbotapi.GetChatMemberConfig{
			ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
				ChatID: chat.ChatConfig().ChatID,
				UserID: bot.Self.ID,
			},
		}); !botAsMember.IsAdministrator() {
			// msg := tgbotapi.NewMessage(chat.ChatConfig().ChatID, "")
			// msg.Text = fmt.Sprintf(
			// 	"%s\n%s",
			// 	"To continue working with me in group or channel,",
			// 	"please grant me administrator rights. Thanks!",
			// )

			// if _, err := bot.Send(msg); err != nil {
			// 	log.Println(err)
			// }
			return false
		} else if err != nil {
			log.Println(err)
		}
	}
	return true
}

// Handles /pin command
func PinMessage(bot tgbotapi.BotAPI, update tgbotapi.Update, botChats map[int64]int64) {
	currChat := update.FromChat()
	pinText := update.Message.CommandArguments()
	msg := tgbotapi.NewMessage(currChat.ChatConfig().ChatID, "")

	if len(pinText) == 0 {
		msg.Text = fmt.Sprintf(
			"%s\n\n%s",
			"To use the /pin command correctly, enter the text of the message, with a space before the text, for example:",
			"/pin <The text to be pinned>",
		)
		if _, err := bot.Send(msg); err != nil {
			log.Println(err)
		}
		return
	}

	if len(botChats) > 0 {
		var chatsWithoutAdminRole []string

		// Iterating through chats where bot is a member
		for _, chatId := range botChats {

			if !CheckAdminRole(bot, chatId) {
				chatsWithoutAdminRole = append(chatsWithoutAdminRole, strconv.Itoa(int(chatId)))
			} else {
				pinMsg := tgbotapi.NewMessage(chatId, "")
				pinMsg.Text = pinText
				sendedPinMsg, err := bot.Send(pinMsg)

				if err != nil {
					log.Println(err)
				}

				pinConfig := tgbotapi.PinChatMessageConfig{
					ChatID:              chatId,
					ChannelUsername:     bot.Self.UserName,
					MessageID:           sendedPinMsg.MessageID,
					DisableNotification: false,
				}

				if _, err := bot.Request(pinConfig); err != nil {
					log.Println(err)
				}
			}
		}

		if len(chatsWithoutAdminRole) > 0 {
			msg.Text = fmt.Sprintf(
				"Some chats are not provied Administrator role for me (below are the chat IDs):\n%s",
				strings.Join(chatsWithoutAdminRole, "\n"),
			)
			if _, err := bot.Send(msg); err != nil {
				log.Println(err)
			}
		}
	} else {
		msg.Text = "Bot is not added to any group for now -_-"
		if _, err := bot.Send(msg); err != nil {
			log.Println(err)
		}
	}
}
