package bot

import (
	"log"
	"regexp"
	"strings"
	"sync/atomic"

	"github.com/pyed/transmission"
	"github.com/pyed/transmission-telegram/internal/config"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

// Bot manages the Telegram bot and Transmission client interaction.
type Bot struct {
	API    *tgbotapi.BotAPI
	Client *transmission.TransmissionClient
	Config *config.Config
	Logger *log.Logger

	chatID       int64 // accessed atomically for thread safety
	mdReplacer   *strings.Replacer
	trackerRegex *regexp.Regexp
	commands     map[string]commandFunc
}

// commandFunc is the handler signature for bot commands.
type commandFunc func(ud tgbotapi.Update, args []string)

// New creates a new Bot instance.
func New(cfg *config.Config, client *transmission.TransmissionClient, api *tgbotapi.BotAPI, logger *log.Logger) *Bot {
	b := &Bot{
		API:    api,
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
	return b
}

// Run starts the main update processing loop.
func (b *Bot) Run(updates <-chan tgbotapi.Update) {
	for update := range updates {
		if update.Message == nil {
			continue
		}

		if !b.Config.Masters.Contains(update.Message.From.UserName) {
			b.Logger.Printf("[INFO] Ignored a message from: %s", update.Message.From.String())
			continue
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
			go handler(update, tokens[1:])
		} else if command == "" {
			go b.receiveTorrent(update)
		} else {
			go b.Send("No such command, try /help", update.Message.Chat.ID, false)
		}
	}
}

// ChatID returns the current chat ID atomically (used by the notify package).
func (b *Bot) ChatID() int64 {
	return atomic.LoadInt64(&b.chatID)
}
