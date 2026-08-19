package bot

import (
	"strings"
	"unicode/utf8"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

// Send sends a message to a chat, splitting long messages safely.
// Returns the message ID of the last sent message.
func (b *Bot) Send(text string, chatID int64, markdown bool) int {
	// set typing action
	action := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	b.API.Send(action)

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

		msg := tgbotapi.NewMessage(chatID, text[:splitAt])
		msg.DisableWebPagePreview = true
		if markdown {
			msg.ParseMode = tgbotapi.ModeMarkdown
		}
		if _, err := b.API.Send(msg); err != nil {
			b.Logger.Printf("[ERROR] Send: %s", err)
		}

		// Move to the next chunk, skip the newline.
		text = text[splitAt:]
		if len(text) > 0 && text[0] == '\n' {
			text = text[1:]
		}
	}

	// Send the remaining (or only) chunk.
	msg := tgbotapi.NewMessage(chatID, text)
	msg.DisableWebPagePreview = true
	if markdown {
		msg.ParseMode = tgbotapi.ModeMarkdown
	}

	resp, err := b.API.Send(msg)
	if err != nil {
		b.Logger.Printf("[ERROR] Send: %s", err)
	}

	return resp.MessageID
}
