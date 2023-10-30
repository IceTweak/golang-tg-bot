package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	lo "github.com/samber/lo"
)

// TODO: need to add -100 to channels and group IDs for bot

func SendMessage(bot tgbotapi.BotAPI, chatId int64, text string) {
	msg := tgbotapi.NewMessage(chatId, text)
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

// Answer inline requests
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

// Checks if the bot has an ability to pin messages
func CheckPinAbility(bot tgbotapi.BotAPI, chatId int64) bool {
	// Get tgbotapi.Chat from chatId
	chat, err := bot.GetChat(tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: chatId,
		},
	})

	if err != nil {
		return false
		// log.Println("Cannot parse this chat because not a member")
		// log.Println(err)
	}

	// Boolean filter, only for chats - not for direct messages to the Bot
	isPrivGroupOrChan := (chat.IsGroup() || chat.IsChannel()) && chat.IsPrivate()
	isPubGroupOrChan := chat.IsGroup() || chat.IsChannel()

	if isPrivGroupOrChan || isPubGroupOrChan || chat.IsSuperGroup() {
		if botAsMember, err := bot.GetChatMember(tgbotapi.GetChatMemberConfig{
			ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
				ChatID: chat.ChatConfig().ChatID,
				UserID: bot.Self.ID,
			},
		}); !botAsMember.CanPinMessages {
			return false
		} else if err != nil {
			log.Println(err)
		}
	}
	return true
}

// Process /pin command
func PinMessage(bot tgbotapi.BotAPI, update tgbotapi.Update) {
	currChat := update.FromChat()
	pinText, botChats := parsePinCommand(update.Message.CommandArguments())
	msg := tgbotapi.NewMessage(currChat.ChatConfig().ChatID, "")

	if len(pinText) == 0 {
		msg.Text = fmt.Sprintf(
			"%s\n\n%s",
			"To use the /pin command correctly, enter the text of the message, with a space before the text, for example:",
			"/pin The text to be pinned... [Chats] first-id, second-id",
		)
		if _, err := bot.Send(msg); err != nil {
			log.Println(err)
		}
		return
	}

	if len(botChats) > 0 {
		var chatsWithoutPinAbility []string

		for _, chatId := range botChats {

			if !CheckPinAbility(bot, chatId) {
				chatsWithoutPinAbility = append(chatsWithoutPinAbility, strconv.Itoa(int(chatId)))
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

		if len(chatsWithoutPinAbility) > 0 {
			msg.Text = fmt.Sprintf(
				"Some chats are not add bot to the chat or not provied ability to pin messages for me (below are the chat IDs):\n%s",
				strings.Join(chatsWithoutPinAbility, "\n"),
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

// Parse /pin command arguments and returns it
func parsePinCommand(command string) (string, []int64) {
	options := strings.Split(command, "[Chats] ")
	if len(options) != 2 {
		return "", []int64{}
	}
	strChatIds := strings.Split(options[1], ",")
	chatIds := lo.Map(strChatIds, func(item string, _ int) int64 {
		id, err := strconv.Atoi(item)
		if err != nil {
			log.Println(err)
		}
		return int64(id)
	})
	return options[0], chatIds
}
