package bot

import (
	"time"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

// liveUpdate repeatedly edits a sent message with fresh data.
// updateFn is called each iteration and should return the new message text ("" to skip).
// finalFn is called once after the loop for a "done" state display (nil to skip entirely).
func (b *Bot) liveUpdate(chatID int64, msgID int, updateFn func() string, finalFn func() string) {
	if b.Config.NoLive {
		return
	}
	for i := 0; i < b.Config.Duration; i++ {
		time.Sleep(b.Config.Interval)
		if text := updateFn(); text != "" {
			editConf := tgbotapi.NewEditMessageText(chatID, msgID, text)
			editConf.ParseMode = tgbotapi.ModeMarkdown
			b.API.Send(editConf)
		}
	}
	if finalFn != nil {
		time.Sleep(b.Config.Interval)
		if text := finalFn(); text != "" {
			editConf := tgbotapi.NewEditMessageText(chatID, msgID, text)
			editConf.ParseMode = tgbotapi.ModeMarkdown
			b.API.Send(editConf)
		}
	}
}
