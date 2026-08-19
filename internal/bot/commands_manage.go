package bot

import (
	"fmt"
	"strconv"

	"github.com/pyed/transmission"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

// add takes an URL to a .torrent file to add it to transmission
func (b *Bot) add(ud tgbotapi.Update, args []string) {
	if len(args) == 0 {
		b.Send("*add:* needs at least one URL", ud.Message.Chat.ID, false)
		return
	}

	// loop over the URL/s and add them
	for _, url := range args {
		cmd := transmission.NewAddCmdByURL(url)

		torrent, err := b.Client.ExecuteAddCommand(cmd)
		if err != nil {
			b.Send("*add:* "+err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		// check if torrent.Name is empty, then an error happened
		if torrent.Name == "" {
			b.Send("*add:* error adding "+url, ud.Message.Chat.ID, false)
			continue
		}
		b.Send(fmt.Sprintf("*Added:* <%d> %s", torrent.ID, torrent.Name), ud.Message.Chat.ID, false)
	}
}

// receiveTorrent gets an update that potentially has a .torrent file to add
func (b *Bot) receiveTorrent(ud tgbotapi.Update) {
	if ud.Message.Document == nil {
		return // has no document
	}

	// get the file ID and make the config
	fconfig := tgbotapi.FileConfig{
		FileID: ud.Message.Document.FileID,
	}
	file, err := b.API.GetFile(fconfig)
	if err != nil {
		b.Send("*receiver:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	// add by file URL
	b.add(ud, []string{file.Link(b.Config.BotToken)})
}

// del takes an id or more, and delete the corresponding torrent/s
func (b *Bot) del(ud tgbotapi.Update, args []string) {
	// make sure that we got an argument
	if len(args) == 0 {
		b.Send("*del:* needs an ID", ud.Message.Chat.ID, false)
		return
	}

	// loop over args to read each potential id
	for _, id := range args {
		num, err := strconv.Atoi(id)
		if err != nil {
			b.Send(fmt.Sprintf("*del:* %s is not an ID", id), ud.Message.Chat.ID, false)
			continue
		}

		name, err := b.Client.DeleteTorrent(num, false)
		if err != nil {
			b.Send("*del:* "+err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		b.Send("*Deleted:* "+name, ud.Message.Chat.ID, false)
	}
}

// deldata takes an id or more, and delete the corresponding torrent/s with their data
func (b *Bot) deldata(ud tgbotapi.Update, args []string) {
	// make sure that we got an argument
	if len(args) == 0 {
		b.Send("*deldata:* needs an ID", ud.Message.Chat.ID, false)
		return
	}
	// loop over args to read each potential id
	for _, id := range args {
		num, err := strconv.Atoi(id)
		if err != nil {
			b.Send(fmt.Sprintf("*deldata:* %s is not an ID", id), ud.Message.Chat.ID, false)
			continue
		}

		name, err := b.Client.DeleteTorrent(num, true)
		if err != nil {
			b.Send("*deldata:* "+err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		b.Send("Deleted with data: "+name, ud.Message.Chat.ID, false)
	}
}

// stop takes id[s] of torrent[s] or 'all' to stop them
func (b *Bot) stop(ud tgbotapi.Update, args []string) {
	// make sure that we got at least one argument
	if len(args) == 0 {
		b.Send("*stop:* needs an argument", ud.Message.Chat.ID, false)
		return
	}

	// if the first argument is 'all' then stop all torrents
	if args[0] == "all" {
		if err := b.Client.StopAll(); err != nil {
			b.Send("*stop:* error occurred while stopping some torrents", ud.Message.Chat.ID, false)
			return
		}
		b.Send("Stopped all torrents", ud.Message.Chat.ID, false)
		return
	}

	for _, id := range args {
		num, err := strconv.Atoi(id)
		if err != nil {
			b.Send(fmt.Sprintf("*stop:* %s is not a number", id), ud.Message.Chat.ID, false)
			continue
		}
		status, err := b.Client.StopTorrent(num)
		if err != nil {
			b.Send("*stop:* "+err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		torrent, err := b.Client.GetTorrent(num)
		if err != nil {
			b.Send(fmt.Sprintf("[fail] *stop:* No torrent with an ID of %d", num), ud.Message.Chat.ID, false)
			continue
		}
		b.Send(fmt.Sprintf("[%s] *stop:* %s", status, torrent.Name), ud.Message.Chat.ID, false)
	}
}

// start takes id[s] of torrent[s] or 'all' to start them
func (b *Bot) start(ud tgbotapi.Update, args []string) {
	// make sure that we got at least one argument
	if len(args) == 0 {
		b.Send("*start:* needs an argument", ud.Message.Chat.ID, false)
		return
	}

	// if the first argument is 'all' then start all torrents
	if args[0] == "all" {
		if err := b.Client.StartAll(); err != nil {
			b.Send("*start:* error occurred while starting some torrents", ud.Message.Chat.ID, false)
			return
		}
		b.Send("Started all torrents", ud.Message.Chat.ID, false)
		return
	}

	for _, id := range args {
		num, err := strconv.Atoi(id)
		if err != nil {
			b.Send(fmt.Sprintf("*start:* %s is not a number", id), ud.Message.Chat.ID, false)
			continue
		}
		status, err := b.Client.StartTorrent(num)
		if err != nil {
			b.Send("*start:* "+err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		torrent, err := b.Client.GetTorrent(num)
		if err != nil {
			b.Send(fmt.Sprintf("[fail] *start:* No torrent with an ID of %d", num), ud.Message.Chat.ID, false)
			continue
		}
		b.Send(fmt.Sprintf("[%s] *start:* %s", status, torrent.Name), ud.Message.Chat.ID, false)
	}
}

// check takes id[s] of torrent[s] or 'all' to verify them
func (b *Bot) check(ud tgbotapi.Update, args []string) {
	// make sure that we got at least one argument
	if len(args) == 0 {
		b.Send("*check:* needs an argument", ud.Message.Chat.ID, false)
		return
	}

	// if the first argument is 'all' then verify all torrents
	if args[0] == "all" {
		if err := b.Client.VerifyAll(); err != nil {
			b.Send("*check:* error occurred while verifying some torrents", ud.Message.Chat.ID, false)
			return
		}
		b.Send("Verifying all torrents", ud.Message.Chat.ID, false)
		return
	}

	for _, id := range args {
		num, err := strconv.Atoi(id)
		if err != nil {
			b.Send(fmt.Sprintf("*check:* %s is not a number", id), ud.Message.Chat.ID, false)
			continue
		}
		status, err := b.Client.VerifyTorrent(num)
		if err != nil {
			b.Send("*check:* "+err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		torrent, err := b.Client.GetTorrent(num)
		if err != nil {
			b.Send(fmt.Sprintf("[fail] *check:* No torrent with an ID of %d", num), ud.Message.Chat.ID, false)
			continue
		}
		b.Send(fmt.Sprintf("[%s] *check:* %s", status, torrent.Name), ud.Message.Chat.ID, false)
	}
}
