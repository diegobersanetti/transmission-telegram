package bot

import (
	"context"
	"strings"
	"unicode/utf8"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// splitMessage splits a long text into chunks of at most maxChars runes,
// preferring to split on newline boundaries.
func splitMessage(text string, maxChars int) []string {
	if maxChars <= 0 {
		maxChars = 4096
	}

	var chunks []string
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

		chunks = append(chunks, text[:splitAt])

		// Move to the next chunk, skip the newline.
		text = text[splitAt:]
		if len(text) > 0 && text[0] == '\n' {
			text = text[1:]
		}
	}

	if len(text) > 0 {
		chunks = append(chunks, text)
	}

	return chunks
}

// Send sends a message to a chat, splitting long messages safely.
// Returns the message ID of the last sent message.
func (b *Bot) Send(ctx context.Context, text string, chatID int64, markdown bool) int {
	return b.SendWithKeyboard(ctx, text, chatID, markdown, nil)
}

// SendWithKeyboard sends a message with an optional inline keyboard markup.
func (b *Bot) SendWithKeyboard(ctx context.Context, text string, chatID int64, markdown bool, markup models.ReplyMarkup) int {
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

	chunks := splitMessage(text, 4096)
	var lastMsgID int

	for i, chunk := range chunks {
		params := &tgbot.SendMessageParams{
			ChatID:    chatID,
			Text:      chunk,
			ParseMode: parseMode,
			LinkPreviewOptions: &models.LinkPreviewOptions{
				IsDisabled: tgbot.True(),
			},
		}
		if i == len(chunks)-1 && markup != nil {
			params.ReplyMarkup = markup
		}

		resp, err := b.API.SendMessage(ctx, params)
		if err != nil {
			b.Logger.Printf("[ERROR] Send: %s", err)
			continue
		}
		lastMsgID = resp.ID
	}

	return lastMsgID
}
