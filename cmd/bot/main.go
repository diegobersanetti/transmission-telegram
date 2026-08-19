package main

import (
	"context"
	"fmt"
	"log/slog"
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

	// Setup structured logger
	var logOutput *os.File = os.Stdout
	if cfg.LogFile != "" {
		logf, err := os.OpenFile(cfg.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Open logfile: %s\n", err)
			os.Exit(1)
		}
		defer logf.Close()
		logOutput = logf
	}

	logger := slog.New(slog.NewTextHandler(logOutput, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("Starting transmission-telegram",
		"version", config.VERSION,
		"masters", cfg.Masters,
		"url", cfg.RPCURL,
		"user", cfg.Username,
	)

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

	// Start completion notifications
	if cfg.TransLogFile != "" {
		notify.StartTailer(cfg.TransLogFile, b.ChatID, func(text string, chatID int64, markdown bool) int {
			return b.Send(ctx, text, chatID, markdown)
		}, logger)
	} else {
		notify.StartWatcher(ctx, cfg.Interval, client.GetTorrents, b.ChatID, func(text string, chatID int64, markdown bool) int {
			return b.Send(ctx, text, chatID, markdown)
		}, logger)
	}

	// Run the bot (blocks until context is cancelled)
	b.Start(ctx)
}
