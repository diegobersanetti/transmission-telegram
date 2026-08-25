package bot

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/pyed/transmission"
)

// add takes an URL to a .torrent file to add it to transmission
func (b *Bot) add(ctx context.Context, ud *models.Update, args []string) {
	if len(args) == 0 {
		b.Send(ctx, "*add:* needs at least one URL", ud.Message.Chat.ID, false)
		return
	}

	// loop over the URL/s and add them
	for _, url := range args {
		cmd := transmission.NewAddCmdByURL(url)

		torrent, err := b.Client.ExecuteAddCommand(cmd)
		if err != nil {
			b.Send(ctx, "*add:* "+err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		// check if torrent.Name is empty, then an error happened
		if torrent.Name == "" {
			b.Send(ctx, "*add:* error adding "+url, ud.Message.Chat.ID, false)
			continue
		}

		torrentInfo, err := b.Client.GetTorrent(torrent.ID)
		if err == nil && torrentInfo != nil && torrentInfo.SizeWhenDone > 0 {
			b.Send(ctx, fmt.Sprintf("*Added:* `<%d>` *%s* (%s)",
				torrent.ID, b.mdReplacer.Replace(torrent.Name), humanize.Bytes(torrentInfo.SizeWhenDone)), ud.Message.Chat.ID, true)
		} else {
			b.Send(ctx, fmt.Sprintf("*Added:* `<%d>` %s", torrent.ID, torrent.Name), ud.Message.Chat.ID, false)
		}
	}
}

// receiveTorrent gets an update that potentially has a .torrent file to add
func (b *Bot) receiveTorrent(ctx context.Context, ud *models.Update) {
	if ud.Message.Document == nil {
		return // has no document
	}

	file, err := b.API.GetFile(ctx, &tgbot.GetFileParams{
		FileID: ud.Message.Document.FileID,
	})
	if err != nil {
		b.Send(ctx, "*receiver:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	// Download the file in this process and add it by raw metainfo. The
	// Telegram file URL contains the bot token, so it must never be handed
	// to Transmission; downloading in-process keeps that credential from
	// being disclosed to Transmission.
	data, err := downloadTelegramFile(ctx, b.API.FileDownloadLink(file))
	if err != nil {
		b.Logger.Warn("Telegram file download failed", "error", err)
		b.Send(ctx, "*receiver:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	torrent, err := b.Client.AddTorrentByData(ctx, data)
	if err != nil {
		b.Send(ctx, "*receiver:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}
	if torrent == nil || torrent.Name == "" {
		b.Send(ctx, "*receiver:* error adding torrent", ud.Message.Chat.ID, false)
		return
	}

	b.Send(ctx, fmt.Sprintf("*Added:* `<%d>` %s", torrent.ID, b.mdReplacer.Replace(torrent.Name)), ud.Message.Chat.ID, true)
}

// maxTorrentDownloadSize is the largest .torrent file we will download from
// Telegram; real .torrent files are small, so anything bigger is rejected.
const maxTorrentDownloadSize = 10 << 20 // 10 MiB

// torrentDownloadClient has an explicit timeout for Telegram file downloads;
// the passed context can still cancel the download earlier.
var torrentDownloadClient = &http.Client{Timeout: 60 * time.Second}

// downloadTelegramFile downloads a Telegram file by URL. The URL contains the
// bot token, so it must only ever be requested from this process.
func downloadTelegramFile(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, privateDownloadError{cause: err}
	}
	resp, err := torrentDownloadClient.Do(req)
	if err != nil {
		return nil, privateDownloadError{cause: err}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}
	// Read at most maxTorrentDownloadSize+1 bytes so an oversize response is
	// detected without buffering it unboundedly.
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxTorrentDownloadSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxTorrentDownloadSize {
		return nil, fmt.Errorf("file exceeds maximum size of %d bytes", maxTorrentDownloadSize)
	}
	return data, nil
}

// privateDownloadError retains the underlying error for errors.Is/As while
// ensuring its token-bearing Telegram URL is never rendered in a log or chat.
type privateDownloadError struct {
	cause error
}

func (e privateDownloadError) Error() string {
	return "could not download file from Telegram"
}

func (e privateDownloadError) Unwrap() error {
	return e.cause
}

// del takes an id or more, and delete the corresponding torrent/s
func (b *Bot) del(ctx context.Context, ud *models.Update, args []string) {
	// make sure that we got an argument
	if len(args) == 0 {
		b.Send(ctx, "*del:* needs an ID", ud.Message.Chat.ID, false)
		return
	}

	// loop over args to read each potential id
	for _, id := range args {
		num, err := strconv.Atoi(id)
		if err != nil {
			b.Send(ctx, fmt.Sprintf("*del:* %s is not an ID", id), ud.Message.Chat.ID, false)
			continue
		}

		name, err := b.Client.DeleteTorrent(num, false)
		if err != nil {
			b.Send(ctx, "*del:* "+err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		b.Send(ctx, "*Deleted:* "+name, ud.Message.Chat.ID, false)
	}
}

// deldata takes an id or more, and delete the corresponding torrent/s with their data
func (b *Bot) deldata(ctx context.Context, ud *models.Update, args []string) {
	// make sure that we got an argument
	if len(args) == 0 {
		b.Send(ctx, "*deldata:* needs an ID", ud.Message.Chat.ID, false)
		return
	}
	// loop over args to read each potential id
	for _, id := range args {
		num, err := strconv.Atoi(id)
		if err != nil {
			b.Send(ctx, fmt.Sprintf("*deldata:* %s is not an ID", id), ud.Message.Chat.ID, false)
			continue
		}

		name, err := b.Client.DeleteTorrent(num, true)
		if err != nil {
			b.Send(ctx, "*deldata:* "+err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		b.Send(ctx, "Deleted with data: "+name, ud.Message.Chat.ID, false)
	}
}

// stop takes id[s] of torrent[s] or 'all' to stop them
func (b *Bot) stop(ctx context.Context, ud *models.Update, args []string) {
	// make sure that we got at least one argument
	if len(args) == 0 {
		b.Send(ctx, "*stop:* needs an argument", ud.Message.Chat.ID, false)
		return
	}

	// if the first argument is 'all' then stop all torrents
	if args[0] == "all" {
		if err := b.Client.StopAll(); err != nil {
			b.Send(ctx, "*stop:* error occurred while stopping some torrents", ud.Message.Chat.ID, false)
			return
		}
		b.Send(ctx, "Stopped all torrents", ud.Message.Chat.ID, false)
		return
	}

	for _, id := range args {
		num, err := strconv.Atoi(id)
		if err != nil {
			b.Send(ctx, fmt.Sprintf("*stop:* %s is not a number", id), ud.Message.Chat.ID, false)
			continue
		}
		status, err := b.Client.StopTorrent(num)
		if err != nil {
			b.Send(ctx, "*stop:* "+err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		torrent, err := b.Client.GetTorrent(num)
		if err != nil {
			b.Send(ctx, fmt.Sprintf("[fail] *stop:* No torrent with an ID of %d", num), ud.Message.Chat.ID, false)
			continue
		}
		b.Send(ctx, fmt.Sprintf("[%s] *stop:* %s", status, torrent.Name), ud.Message.Chat.ID, false)
	}
}

// start takes id[s] of torrent[s] or 'all' to start them
func (b *Bot) start(ctx context.Context, ud *models.Update, args []string) {
	// make sure that we got at least one argument
	if len(args) == 0 {
		b.Send(ctx, "*start:* needs an argument", ud.Message.Chat.ID, false)
		return
	}

	// if the first argument is 'all' then start all torrents
	if args[0] == "all" {
		if err := b.Client.StartAll(); err != nil {
			b.Send(ctx, "*start:* error occurred while starting some torrents", ud.Message.Chat.ID, false)
			return
		}
		b.Send(ctx, "Started all torrents", ud.Message.Chat.ID, false)
		return
	}

	for _, id := range args {
		num, err := strconv.Atoi(id)
		if err != nil {
			b.Send(ctx, fmt.Sprintf("*start:* %s is not a number", id), ud.Message.Chat.ID, false)
			continue
		}
		status, err := b.Client.StartTorrent(num)
		if err != nil {
			b.Send(ctx, "*start:* "+err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		torrent, err := b.Client.GetTorrent(num)
		if err != nil {
			b.Send(ctx, fmt.Sprintf("[fail] *start:* No torrent with an ID of %d", num), ud.Message.Chat.ID, false)
			continue
		}
		b.Send(ctx, fmt.Sprintf("[%s] *start:* %s", status, torrent.Name), ud.Message.Chat.ID, false)
	}
}

// check takes id[s] of torrent[s] or 'all' to verify them
func (b *Bot) check(ctx context.Context, ud *models.Update, args []string) {
	// make sure that we got at least one argument
	if len(args) == 0 {
		b.Send(ctx, "*check:* needs an argument", ud.Message.Chat.ID, false)
		return
	}

	// if the first argument is 'all' then verify all torrents
	if args[0] == "all" {
		if err := b.Client.VerifyAll(); err != nil {
			b.Send(ctx, "*check:* error occurred while verifying some torrents", ud.Message.Chat.ID, false)
			return
		}
		b.Send(ctx, "Verifying all torrents", ud.Message.Chat.ID, false)
		return
	}

	for _, id := range args {
		num, err := strconv.Atoi(id)
		if err != nil {
			b.Send(ctx, fmt.Sprintf("*check:* %s is not a number", id), ud.Message.Chat.ID, false)
			continue
		}
		status, err := b.Client.VerifyTorrent(num)
		if err != nil {
			b.Send(ctx, "*check:* "+err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		torrent, err := b.Client.GetTorrent(num)
		if err != nil {
			b.Send(ctx, fmt.Sprintf("[fail] *check:* No torrent with an ID of %d", num), ud.Message.Chat.ID, false)
			continue
		}
		b.Send(ctx, fmt.Sprintf("[%s] *check:* %s", status, torrent.Name), ud.Message.Chat.ID, false)
	}
}
