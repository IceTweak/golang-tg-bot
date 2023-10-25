package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	godotenv "github.com/joho/godotenv"
)

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
	companies := ParseCompFromXml("companies.xml")

	for update := range updates {
		chat := update.FromChat()
		// TODO - error occurs here
		if chat.IsChannel() || chat.IsGroup() {
			if botAsMember, err := bot.GetChatMember(tgbotapi.GetChatMemberConfig{
				ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
					ChatID: chat.ID,
					UserID: bot.Self.ID,
				},
			}); !botAsMember.IsAdministrator() {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
				msg.Text = "To continue working with me, you need to grant me administrator rights. Thanks!"
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
				continue
			} else if err != nil {
				log.Println(err)
			}
		}
		if update.Message == nil && update.InlineQuery != nil { // If we got a message
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

		} else if update.Message.IsCommand() {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			switch update.Message.Command() {
			case "pin":
				pinConfig := tgbotapi.PinChatMessageConfig{
					ChatID:              chat.ID,
					ChannelUsername:     update.Message.From.UserName,
					MessageID:           update.Message.MessageID,
					DisableNotification: false,
				}
				if _, err := bot.Request(pinConfig); err != nil {
					log.Println(err)
				}
				msg.Text = "Succsessfully pin your message!"
			case "sayhi":
				msg.Text = "Hi :)"
			case "status":
				msg.Text = "I'm ok."
			default:
				msg.Text = "I don't know that command"
			}

			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
		} else {
			fmt.Println("None of the message types handled!")
		}
	}
}
