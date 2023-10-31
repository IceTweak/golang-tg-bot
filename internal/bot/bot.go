package bot

import (
	"net/http"

	config "github.com/IceTweak/golang-tg-bot/internal/config"
	model "github.com/IceTweak/golang-tg-bot/internal/model"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	zap "go.uber.org/zap"
)

// Describes Bot structure
type Bot struct {
	API    *tgbotapi.BotAPI
	Config *config.Config
	Logger *zap.Logger
	// Flow   model.Flow
	// Service    *service.Service
	// Repository *repository.Repository
}

// Init Bot
func Init(config *config.Config, logger *zap.Logger) *Bot {
	return &Bot{
		Config: config,
		Logger: logger,
		// Flow:   flow,
		// Service:    service,
		// Repository: repository,
	}
}

// Run Bot
func (bot *Bot) Run() {
	botAPI, err := bot.NewBotAPI()
	if err != nil {
		bot.Logger.Fatal("failed create new bot api instance", zap.String("error", err.Error()))
	}
	bot.API = botAPI

	// !!! THERE IS NO WEBHOOK SERVER FOR NOW
	// if err := bot.SetWebhook(); err != nil {
	// 	bot.Logger.Fatal("failed set webhook", zap.String("error", err.Error()))
	// }
	// !!!

	if err := bot.SetBotCommands(); err != nil {
		bot.Logger.Fatal("failed set bot commands", zap.String("error", err.Error()))
	}

	// !!! THERE IS NO WEBHOOK SERVER FOR NOW
	// go bot.StartWebhookServer()
	// bot.Logger.Info("http webhook server started")

	// updates := b.API.ListenForWebhook("/" + b.API.Token)
	// !!!

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.API.GetUpdatesChan(u)
	for update := range updates {
		go bot.UpdateRouter(update)
	}
}

// Setup telegram bot API for Bot
func (bot *Bot) NewBotAPI() (*tgbotapi.BotAPI, error) {
	botAPI, err := tgbotapi.NewBotAPI(bot.Config.Bot.Token)
	if err != nil {
		return nil, err
	}
	bot.Logger.Info("authorized success, bot api instance created", zap.String("account", botAPI.Self.UserName))
	return botAPI, nil
}

// Setter for webhook in Config
func (bot *Bot) SetWebhook() error {
	webhook, err := tgbotapi.NewWebhook(bot.Config.Bot.WebhookLink + bot.API.Token)
	if err != nil {
		return err
	}
	_, err = bot.API.Request(webhook)
	if err != nil {
		return err
	}
	info, err := bot.API.GetWebhookInfo()
	if err != nil {
		return err
	}
	bot.Logger.Info("webhook info", zap.Any("webhook", info))
	if info.LastErrorDate != 0 {
		return err
	}
	return nil
}

// Configure the bot menu
func (bot *Bot) InitBotCommands() tgbotapi.SetMyCommandsConfig {
	commands := []model.CommandEntity{
		{
			Key:  model.PinCommand,
			Name: "pin",
		},
		/* implement your commands in the same way
		{
			Key:  model.<...>,
			Name: "...",
		},
		...
		*/
	}
	tgCommands := make([]tgbotapi.BotCommand, 0, len(commands))
	for _, cmd := range commands {
		tgCommands = append(tgCommands, tgbotapi.BotCommand{
			Command:     "/" + string(cmd.Key),
			Description: cmd.Name,
		})
	}
	commandsConfig := tgbotapi.NewSetMyCommands(tgCommands...)
	return commandsConfig
}

// Request to set Bot commands config
func (bot *Bot) SetBotCommands() error {
	commandsConfig := bot.InitBotCommands()
	_, err := bot.API.Request(commandsConfig)
	if err != nil {
		return err
	}
	return nil
}

// Start listen to WebHooks server
func (bot *Bot) StartWebhookServer() {
	if err := http.ListenAndServe(bot.Config.Server.Host+bot.Config.Server.Port, nil); err != nil {
		bot.Logger.Fatal("failed start http server", zap.String("error", err.Error()))
	}
}

// Send message or request
func (bot *Bot) SendMessage(msg tgbotapi.Chattable) {
	_, err := bot.API.Request(msg)
	if err != nil {
		bot.Logger.Error("failed send message to bot", zap.String("error", err.Error()))
	}
}
