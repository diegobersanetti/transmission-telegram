package bot

import (
	"context"
	"strings"
	"unicode/utf8"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Send sends a message to a chat, splitting long messages safely.
// Returns the message ID of the last sent message.
func (b *Bot) Send(ctx context.Context, text string, chatID int64, markdown bool) int {
	if ctx == nil {
		ctx = context.Background()
	}

	// set typing action
	b.API.SendChatAction(ctx, &tgbot.SendChatActionParams{
		ChatID: chatID,
		Action: models.ChatActionTyping,
	})

	var parseMode models.ParseMode
	if markdown {
		parseMode = models.ParseModeMarkdownV1
	}

	// Telegram is limited to 4096 chars per message.
	// Split long messages on newline boundaries.
	const maxChars = 4096

	for utf8.RuneCountInString(text) > maxChars {
		// Convert rune limit to a byte offset.
		byteLimit := 0
		for i := 0; i < maxChars && byteLimit < len(text); i++ {
			_, size := utf8.DecodeRuneInString(text[byteLimit:])
			byteLimit += size
		}

		// Search backward from byteLimit for a newline to split on.
		splitAt := strings.LastIndex(text[:byteLimit], "\n")
		if splitAt == -1 {
			// No newline found; force split at the rune boundary.
			splitAt = byteLimit
		}

		_, err := b.API.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:    chatID,
			Text:      text[:splitAt],
			ParseMode: parseMode,
			LinkPreviewOptions: &models.LinkPreviewOptions{
				IsDisabled: tgbot.True(),
			},
		})
		if err != nil {
			b.Logger.Printf("[ERROR] Send: %s", err)
		}

		// Move to the next chunk, skip the newline.
		text = text[splitAt:]
		if len(text) > 0 && text[0] == '\n' {
			text = text[1:]
		}
	}

	// Send the remaining (or only) chunk.
	resp, err := b.API.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: parseMode,
		LinkPreviewOptions: &models.LinkPreviewOptions{
			IsDisabled: tgbot.True(),
		},
	})
	if err != nil {
		b.Logger.Printf("[ERROR] Send: %s", err)
		return 0
	}

	return resp.ID
}
