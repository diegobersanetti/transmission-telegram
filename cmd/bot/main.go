package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pyed/transmission"
	"github.com/pyed/transmission-telegram/internal/bot"
	"github.com/pyed/transmission-telegram/internal/config"
	"github.com/pyed/transmission-telegram/internal/notify"
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

	// Create bot instance
	b, err := bot.New(cfg, client, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Telegram: %s\n", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Start log tailer for completion notifications
	if cfg.TransLogFile != "" {
		notify.StartTailer(cfg.TransLogFile, b.ChatID, func(text string, chatID int64, markdown bool) int {
			return b.Send(ctx, text, chatID, markdown)
		}, logger)
	}

	// Run the bot (blocks until context is cancelled)
	b.Start(ctx)
}
