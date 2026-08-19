package bot

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
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
	Logger *slog.Logger

	chatID       int64 // accessed atomically for thread safety
	mdReplacer   *strings.Replacer
	trackerRegex *regexp.Regexp
	commands     map[string]commandFunc
}

// commandFunc is the handler signature for bot commands.
type commandFunc func(ctx context.Context, ud *models.Update, args []string)

// New creates a new Bot instance.
func New(cfg *config.Config, client *transmission.TransmissionClient, logger *slog.Logger) (*Bot, error) {
	if logger == nil {
		logger = slog.Default()
	}

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
		tgbot.WithErrorsHandler(func(err error) {
			logger.Error("Telegram API error", "error", err)
		}),
	}

	api, err := tgbot.New(cfg.BotToken, opts...)
	if err != nil {
		return nil, err
	}
	b.API = api

	// Remove any stale webhook so long polling receives updates reliably
	_, _ = api.DeleteWebhook(context.Background(), &tgbot.DeleteWebhookParams{
		DropPendingUpdates: false,
	})

	me, err := api.GetMe(context.Background())
	if err == nil {
		logger.Info("Telegram bot authorized", "username", me.Username)
	}

	// Register command list with Telegram menu button
	go func() {
		_, err := api.SetMyCommands(context.Background(), &tgbot.SetMyCommandsParams{
			Commands: defaultBotCommands(),
		})
		if err != nil {
			logger.Warn("SetMyCommands failed", "error", err)
		}
	}()

	return b, nil
}

// defaultBotCommands returns the list of primary commands to display in Telegram's menu.
func defaultBotCommands() []models.BotCommand {
	return []models.BotCommand{
		{Command: "list", Description: "List torrents (optional tracker filter)"},
		{Command: "head", Description: "List first n torrents with live speed"},
		{Command: "tail", Description: "List last n torrents with live speed"},
		{Command: "active", Description: "List active uploading/downloading torrents"},
		{Command: "downs", Description: "List downloading torrents"},
		{Command: "seeding", Description: "List seeding torrents"},
		{Command: "paused", Description: "List paused torrents"},
		{Command: "checking", Description: "List verifying torrents"},
		{Command: "errors", Description: "List torrents with errors"},
		{Command: "speed", Description: "Show current upload & download speeds"},
		{Command: "stats", Description: "Show Transmission cumulative stats"},
		{Command: "free", Description: "Show download directory free disk space"},
		{Command: "turtle", Description: "Toggle Turtle Mode (alternative speed limits)"},
		{Command: "count", Description: "Show torrent count by status"},
		{Command: "help", Description: "Show help message"},
		{Command: "version", Description: "Show Transmission & bot versions"},
	}
}

// Start starts the bot polling loop (blocks until context is cancelled).
func (b *Bot) Start(ctx context.Context) {
	b.API.Start(ctx)
}

// normalizeCommand processes the raw first token of a message:
// - detects magnet/http links
// - strips leading "/"
// - strips trailing "@botname" from group mentions
// - lowercases the command
func normalizeCommand(token string) (cmd string, isLink bool) {
	if strings.HasPrefix(token, "magnet") || strings.HasPrefix(token, "http") {
		return "add", true
	}
	token = strings.ToLower(token)
	token = strings.TrimPrefix(token, "/")
	token = strings.Split(token, "@")[0]
	return token, false
}

// handleUpdate processes incoming updates from Telegram.
func (b *Bot) handleUpdate(ctx context.Context, _ *tgbot.Bot, update *models.Update) {
	if update.CallbackQuery != nil {
		go b.handleCallbackQuery(ctx, update.CallbackQuery)
		return
	}

	if update.Message == nil {
		return
	}

	username := ""
	var userID int64
	if update.Message.From != nil {
		username = update.Message.From.Username
		userID = update.Message.From.ID
	}

	if !b.Config.Masters.Contains(username, userID) {
		fromStr := username
		if fromStr == "" && update.Message.From != nil {
			fromStr = fmt.Sprintf("%s (ID: %d)", update.Message.From.FirstName, userID)
		}
		b.Logger.Warn("Ignored message from unauthorized sender",
			"sender", fromStr,
			"username", username,
			"user_id", userID,
			"chat_id", update.Message.Chat.ID,
			"allowed_masters", b.Config.Masters,
		)
		return
	}

	// Always store active chat ID for notifications
	atomic.StoreInt64(&b.chatID, update.Message.Chat.ID)

	b.Logger.Info("Received message",
		"sender", username,
		"user_id", userID,
		"chat_id", update.Message.Chat.ID,
		"text", update.Message.Text,
	)

	fields := strings.Fields(update.Message.Text)
	if len(fields) == 0 {
		if update.Message.Document != nil {
			go b.receiveTorrent(ctx, update)
		}
		return
	}

	command, isLink := normalizeCommand(fields[0])
	var args []string
	if isLink {
		args = fields
	} else {
		args = fields[1:]
	}

	b.Logger.Info("Dispatching command", "command", command, "args", args, "sender", username)

	if handler, ok := b.commands[command]; ok {
		go handler(ctx, update, args)
	} else if command == "" {
		go b.receiveTorrent(ctx, update)
	} else {
		b.Logger.Warn("Unknown command", "command", command, "sender", username)
		go b.Send(ctx, "No such command, try /help", update.Message.Chat.ID, false)
	}
}

// handleCallbackQuery processes button clicks from inline keyboards.
func (b *Bot) handleCallbackQuery(ctx context.Context, cb *models.CallbackQuery) {
	username := cb.From.Username
	userID := cb.From.ID

	if !b.Config.Masters.Contains(username, userID) {
		b.API.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{
			CallbackQueryID: cb.ID,
			Text:            "Unauthorized",
			ShowAlert:       true,
		})
		return
	}

	// Expected callback format: cmd:<action>:<id>
	parts := strings.Split(cb.Data, ":")
	if len(parts) != 3 || parts[0] != "cmd" {
		b.API.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{
			CallbackQueryID: cb.ID,
		})
		return
	}

	action := parts[1]
	torrentID, err := strconv.Atoi(parts[2])
	if err != nil {
		b.API.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{
			CallbackQueryID: cb.ID,
			Text:            "Invalid torrent ID",
		})
		return
	}

	var answerText string
	switch action {
	case "stop":
		status, err := b.Client.StopTorrent(torrentID)
		if err != nil {
			answerText = "Error stopping: " + err.Error()
		} else {
			answerText = fmt.Sprintf("[%s] Torrent paused", status)
		}
	case "start":
		status, err := b.Client.StartTorrent(torrentID)
		if err != nil {
			answerText = "Error resuming: " + err.Error()
		} else {
			answerText = fmt.Sprintf("[%s] Torrent resumed", status)
		}
	case "del":
		name, err := b.Client.DeleteTorrent(torrentID, false)
		if err != nil {
			answerText = "Error deleting: " + err.Error()
		} else {
			answerText = "Deleted: " + name
		}
	}

	b.API.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{
		CallbackQueryID: cb.ID,
		Text:            answerText,
	})
}

// ChatID returns the current chat ID atomically (used by the notify package).
func (b *Bot) ChatID() int64 {
	return atomic.LoadInt64(&b.chatID)
}
