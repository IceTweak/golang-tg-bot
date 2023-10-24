package golangtgbot

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
							"*Category* _%s_\n"+
							"*Year* _%s_\n"+
							"*Owner* _%s_\n"+
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

		} else {
			// Logic for comand input
			fmt.Println("GGWP")
		}
	}
}
