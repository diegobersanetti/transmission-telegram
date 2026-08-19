package bot

import (
	"context"
	"log"
	"regexp"
	"strings"
	"sync/atomic"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/pyed/transmission"
	"github.com/pyed/transmission-telegram/internal/config"
)

// Bot manages the Telegram bot and Transmission client interaction.
type Bot struct {
	API    *tgbot.Bot
	Client *transmission.TransmissionClient
	Config *config.Config
	Logger *log.Logger

	chatID       int64 // accessed atomically for thread safety
	mdReplacer   *strings.Replacer
	trackerRegex *regexp.Regexp
	commands     map[string]commandFunc
}

// commandFunc is the handler signature for bot commands.
type commandFunc func(ctx context.Context, ud *models.Update, args []string)

// New creates a new Bot instance.
func New(cfg *config.Config, client *transmission.TransmissionClient, logger *log.Logger) (*Bot, error) {
	b := &Bot{
		Client: client,
		Config: cfg,
		Logger: logger,
		mdReplacer: strings.NewReplacer(
			"*", "•",
			"[", "(",
			"]", ")",
			"_", "-",
			"`", "'",
		),
		trackerRegex: regexp.MustCompile(`[https?|udp]://([^:/]*)`),
	}
	b.registerCommands()

	opts := []tgbot.Option{
		tgbot.WithDefaultHandler(b.handleUpdate),
	}

	api, err := tgbot.New(cfg.BotToken, opts...)
	if err != nil {
		return nil, err
	}
	b.API = api

	me, err := api.GetMe(context.Background())
	if err == nil {
		logger.Printf("[INFO] Authorized: %s", me.Username)
	}

	return b, nil
}

// Start starts the bot polling loop (blocks until context is cancelled).
func (b *Bot) Start(ctx context.Context) {
	b.API.Start(ctx)
}

// handleUpdate processes incoming updates from Telegram.
func (b *Bot) handleUpdate(ctx context.Context, _ *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	username := ""
	if update.Message.From != nil {
		username = update.Message.From.Username
	}

	if !b.Config.Masters.Contains(username) {
		fromStr := username
		if fromStr == "" && update.Message.From != nil {
			fromStr = update.Message.From.FirstName
		}
		b.Logger.Printf("[INFO] Ignored a message from: %s", fromStr)
		return
	}

	if b.Config.TransLogFile != "" && atomic.LoadInt64(&b.chatID) != update.Message.Chat.ID {
		atomic.StoreInt64(&b.chatID, update.Message.Chat.ID)
	}

	tokens := strings.Split(update.Message.Text, " ")

	// Auto-detect magnet/http links and prepend "add" command
	if strings.HasPrefix(tokens[0], "magnet") || strings.HasPrefix(tokens[0], "http") {
		tokens = append([]string{"add"}, tokens...)
	}

	// Normalize command: lowercase and strip leading "/"
	command := strings.ToLower(tokens[0])
	command = strings.TrimPrefix(command, "/")

	if handler, ok := b.commands[command]; ok {
		go handler(ctx, update, tokens[1:])
	} else if command == "" {
		go b.receiveTorrent(ctx, update)
	} else {
		go b.Send(ctx, "No such command, try /help", update.Message.Chat.ID, false)
	}
}

// ChatID returns the current chat ID atomically (used by the notify package).
func (b *Bot) ChatID() int64 {
	return atomic.LoadInt64(&b.chatID)
}
