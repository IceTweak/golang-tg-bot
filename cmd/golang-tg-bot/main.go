package main

import (
	"flag"

	"github.com/IceTweak/golang-tg-bot/internal/bot"
	"github.com/IceTweak/golang-tg-bot/internal/config"
	"github.com/IceTweak/golang-tg-bot/internal/logger"

	"go.uber.org/zap"
)

// TODO: ability to add media (photo, video, etc.) to pinned message in /pin command

// TODO: realize command palette (e.g. buttons with commands names)

// TODO: add description to commands

// TODO: fix "not a command" message after inline response

// TODO: need to add -100 to channels and group IDs for bot

func main() {
	configPath := flag.String("c", "./cmd/golang-tg-bot/config.yaml", "path to go-telegram-bot-example config")
	flag.Parse()

	logger := logger.GetLogger()

	cfg := &config.Config{}

	err := config.GetConfiguration(*configPath, cfg)

	if err != nil {
		logger.Fatal("failed get configuration", zap.String("reason", err.Error()))
	}

	logger.Info("configured", zap.Any("config", cfg))

	// Init bot
	bot := bot.Init(cfg, &logger)

	bot.Run()
}
