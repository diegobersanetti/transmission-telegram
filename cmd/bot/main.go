package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pyed/transmission"
	"github.com/pyed/transmission-telegram/internal/bot"
	"github.com/pyed/transmission-telegram/internal/config"
	"github.com/pyed/transmission-telegram/internal/notify"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func main() {
	cfg := config.Parse()

	// Setup logger
	logger := log.New(os.Stdout, "", log.LstdFlags)
	if cfg.LogFile != "" {
		logf, err := os.OpenFile(cfg.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
		logger.SetOutput(logf)
	}

	logger.Printf("[INFO] Token=****\n\t\tMasters=%s\n\t\tURL=%s\n\t\tUSER=%s\n\t\tPASS=****",
		cfg.Masters, cfg.RPCURL, cfg.Username)

	// Initialize Transmission client
	client, err := transmission.New(cfg.RPCURL, cfg.Username, cfg.Password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Transmission: Make sure you have the right URL, Username and Password\n")
		os.Exit(1)
	}

	// Initialize Telegram bot
	api, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Telegram: %s\n", err)
		os.Exit(1)
	}
	logger.Printf("[INFO] Authorized: %s", api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := api.GetUpdatesChan(u)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Telegram: %s\n", err)
		os.Exit(1)
	}

	// Create bot instance
	b := bot.New(cfg, client, api, logger)

	// Start log tailer for completion notifications
	if cfg.TransLogFile != "" {
		notify.StartTailer(cfg.TransLogFile, b.ChatID, b.Send, logger)
	}

	// Run the bot
	b.Run(updates)
}
