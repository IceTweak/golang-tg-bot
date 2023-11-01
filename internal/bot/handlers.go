package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	commonanswers "github.com/IceTweak/golang-tg-bot/internal/messages/common-answers"
	model "github.com/IceTweak/golang-tg-bot/internal/model"
	parser "github.com/IceTweak/golang-tg-bot/internal/parser"
	"github.com/samber/lo"
	zap "go.uber.org/zap"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (bot *Bot) UpdateRouter(update tgbotapi.Update) {
	if msg := update.Message; msg != nil {
		if msg.IsCommand() {
			bot.SendMessage(bot.CommandsHandler(update.Message.Command(), update))
		} else {
			bot.SendMessage(bot.MessageHandler(update))
		}
	}
	if inlineQuery := update.InlineQuery; inlineQuery != nil {
		bot.SendMessage(bot.InlineQueryHandler(*inlineQuery))
	}
	// if cq := update.CallbackQuery; cq != nil {
	// 	bot.SendMessage(bot.CallbacksHandler(updLocal))
	// }
}

func (bot *Bot) CommandsHandler(command string, update tgbotapi.Update) tgbotapi.Chattable {
	updLocal := model.DecodeToLocal(update)
	msg := update.Message
	/*
		your commands processing logic should be here
		return <message>
	*/
	switch command {
	case "pin":
		return bot.pinMessage(updLocal, msg)
	default:
		return commonanswers.UnknownCommand().BuildBotMessage(updLocal.TelegramChatID)
	}
}

func (bot *Bot) MessageHandler(update tgbotapi.Update) tgbotapi.Chattable {
	updLocal := model.DecodeToLocal(update)
	/*
		your message processing logic should be here
		return <message>
	*/
	return commonanswers.UnknownMessage().BuildBotMessage(updLocal.TelegramChatID)
}

// TODO:
// func (bot *Bot) CallbacksHandler(updLocal *model.UpdateLocal) tgbotapi.Chattable {
// 	cData := updLocal.CallbackData
// 	replyMessage, err := bot.Flow.Handle(&cData, updLocal)
// 	if err != nil {
// 		bot.Logger.Error("error", zap.String("reason", err.Error()))
// 		return commonanswers.UnknownError().BuildBotMessage(int64(updLocal.TelegramChatID))
// 	}
// 	return replyMessage
// }

// Answer inline requests
func (bot *Bot) InlineQueryHandler(inlineQuery tgbotapi.InlineQuery) tgbotapi.Chattable {
	// Parse Companies from file
	companies := parser.ParseCompFromXml("./internal/parser/companies.xml")

	// Filter companies along query
	filteredCompanies := parser.Filter(companies.Companies, func(comp parser.Company) bool {
		return strings.Index(strings.ToLower(comp.Title), strings.ToLower(inlineQuery.Query)) >= 0
	})

	var articles []interface{}
	if len(filteredCompanies) == 0 {
		msg := tgbotapi.NewInlineQueryResultArticleMarkdown(inlineQuery.ID, "No one companies matches", "No one companies matches")
		articles = append(articles, msg)
	} else {
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
				strings.Join(parser.MapLinks(comp.Links.Links), "\n"),
			)

			msg := tgbotapi.NewInlineQueryResultArticleMarkdown(comp.Title, comp.Title, text)
			msg.ThumbURL = "https://i.guim.co.uk/img/media/ef8492feb3715ed4de705727d9f513c168a8b196/37_0_1125_675/master/1125.jpg?width=1200&height=1200&quality=85&auto=format&fit=crop&s=d456a2af571d980d8b2985472c262b31"
			msg.ThumbHeight, msg.ThumbWidth = 600, 600
			articles = append(articles, msg)
		}
	}

	return tgbotapi.InlineConfig{
		InlineQueryID: inlineQuery.ID,
		IsPersonal:    true,
		CacheTime:     0,
		Results:       articles,
	}
}

// * Start local methods section

// Checks if the bot has an ability to pin messages
func (bot *Bot) checkPinAbility(chatId int64) bool {
	// Get tgbotapi.Chat from chatId
	chat, err := bot.API.GetChat(tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: chatId,
		},
	})

	if err != nil {
		bot.Logger.Error("failed to get chat from provided chat ID", zap.String("error", err.Error()))
	}

	// Boolean filter, only for chats - not for direct messages to the Bot
	isPrivGroupOrChan := (chat.IsGroup() || chat.IsChannel()) && chat.IsPrivate()
	isPubGroupOrChan := chat.IsGroup() || chat.IsChannel()

	if isPrivGroupOrChan || isPubGroupOrChan || chat.IsSuperGroup() {
		if botAsMember, err := bot.API.GetChatMember(tgbotapi.GetChatMemberConfig{
			ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
				ChatID: chatId,
				UserID: bot.API.Self.ID,
			},
		}); !botAsMember.CanPinMessages {
			return false
		} else if err != nil {
			bot.Logger.Error("failed to check bot role in chat", zap.String("error", err.Error()))
		}
	}
	return true
}

// Process /pin command
func (bot *Bot) pinMessage(update *model.UpdateLocal, message *tgbotapi.Message) tgbotapi.Chattable {
	pinText, botChats := parsePinCommand(message.CommandArguments())
	msg := tgbotapi.NewMessage(int64(update.TelegramChatID), "")

	if len(pinText) == 0 {
		msg.Text = fmt.Sprintf(
			"%s\n\n%s",
			"To use the /pin command correctly, enter the text of the message, with a space before the text, for example:",
			"/pin The text to be pinned... [Chats] first-id, second-id",
		)
		return msg
	}

	if len(botChats) > 0 {
		var chatsWithoutPinAbility []string

		for _, chatId := range botChats {

			if bot.checkPinAbility(chatId) {
				chatsWithoutPinAbility = append(chatsWithoutPinAbility, strconv.Itoa(int(chatId)))
			} else {
				pinMsg := tgbotapi.NewMessage(chatId, "")
				pinMsg.Text = pinText
				sendedPinMsg, err := bot.API.Send(pinMsg)

				if err != nil {
					errText := fmt.Sprintf("failed to pin message in chat: %d", chatId)
					bot.Logger.Error(errText, zap.String("error", err.Error()))
				}

				pinConfig := tgbotapi.PinChatMessageConfig{
					ChatID:              chatId,
					ChannelUsername:     bot.API.Self.UserName,
					MessageID:           sendedPinMsg.MessageID,
					DisableNotification: false,
				}

				bot.SendMessage(pinConfig)
			}
		}

		if len(chatsWithoutPinAbility) > 0 {
			msg.Text = fmt.Sprintf(
				"Some chats are not add bot to the chat or not provied ability to pin messages for me (below are the chat IDs):\n%s",
				strings.Join(chatsWithoutPinAbility, "\n"),
			)
			return msg
		}
	}

	msg.Text = "Successfully pin your message to chats!"
	return msg
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

// * End local methods section
