package bot

import (
	"context"
	"time"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// liveUpdate repeatedly edits a sent message with fresh data.
// updateFn is called each iteration and should return the new message text ("" to skip).
// finalFn is called once after the loop for a "done" state display (nil to skip entirely).
func (b *Bot) liveUpdate(ctx context.Context, chatID int64, msgID int, updateFn func() string, finalFn func() string) {
	if b.Config.NoLive {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}

	for i := 0; i < b.Config.Duration; i++ {
		select {
		case <-ctx.Done():
			return
		case <-time.After(b.Config.Interval):
		}

		if text := updateFn(); text != "" {
			b.API.EditMessageText(ctx, &tgbot.EditMessageTextParams{
				ChatID:    chatID,
				MessageID: msgID,
				Text:      text,
				ParseMode: models.ParseModeMarkdownV1,
			})
		}
	}

	if finalFn != nil {
		select {
		case <-ctx.Done():
			return
		case <-time.After(b.Config.Interval):
		}

		if text := finalFn(); text != "" {
			b.API.EditMessageText(ctx, &tgbot.EditMessageTextParams{
				ChatID:    chatID,
				MessageID: msgID,
				Text:      text,
				ParseMode: models.ParseModeMarkdownV1,
			})
		}
	}
}
