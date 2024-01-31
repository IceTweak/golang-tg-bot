package main

import (
	"flag"

	"github.com/IceTweak/golang-tg-bot/internal/bot"
	"github.com/IceTweak/golang-tg-bot/internal/config"
	"github.com/IceTweak/golang-tg-bot/internal/logger"

	"go.uber.org/zap"
)

// TODO: ability to add media (photo, video, etc.) to pinned message in /pin command

// TODO: need to answer only to direct messages to the bo bot

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

	bot := bot.Init(cfg, &logger)

	bot.Run()
}
