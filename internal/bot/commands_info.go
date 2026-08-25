package bot

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/go-telegram/bot/models"
	"github.com/pyed/transmission"
)

// info takes an id of a torrent and returns some info about it
func (b *Bot) info(ctx context.Context, ud *models.Update, args []string) {
	if len(args) == 0 {
		b.Send(ctx, "*info:* needs a torrent ID number", ud.Message.Chat.ID, false)
		return
	}

	for _, id := range args {
		torrentID, err := strconv.Atoi(id)
		if err != nil {
			b.Send(ctx, fmt.Sprintf("*info:* %s is not a number", id), ud.Message.Chat.ID, false)
			continue
		}

		// get the torrent
		torrent, err := b.Client.GetTorrent(torrentID)
		if err != nil {
			b.Send(ctx, fmt.Sprintf("*info:* Can't find a torrent with an ID of %d", torrentID), ud.Message.Chat.ID, false)
			continue
		}

		// get the trackers using 'trackerRegex'
		var trackers string
		for _, tracker := range torrent.Trackers {
			sm := b.trackerRegex.FindSubmatch([]byte(tracker.Announce))
			if len(sm) > 1 {
				trackers += string(sm[1]) + " "
			}
		}

		// format the info
		// format the info
		torrentName := b.mdReplacer.Replace(torrent.Name) // escape markdown
		info := fmt.Sprintf("`<%d>` *%s*\n%s %s *%.1f%%* (%s / %s) ↓ *%s*  ↑ *%s* R: *%s*\nDL: *%s* UP: *%s*\nAdded: *%s*, ETA: *%s*\nTrackers: `%s`",
			torrent.ID, torrentName, progressBar(torrent.PercentDone, 10), torrent.TorrentStatus(),
			torrent.PercentDone*100, humanize.Bytes(torrent.Have()), humanize.Bytes(torrent.SizeWhenDone),
			humanize.Bytes(torrent.RateDownload), humanize.Bytes(torrent.RateUpload), torrent.Ratio(),
			humanize.Bytes(torrent.DownloadedEver), humanize.Bytes(torrent.UploadedEver), time.Unix(torrent.AddedDate, 0).Format(time.Stamp),
			torrent.ETA(), trackers)

		keyboard := infoKeyboard(torrent.ID, torrent.HashString)

		// send it
		msgID := b.SendWithKeyboard(ctx, info, ud.Message.Chat.ID, true, keyboard)

		// this go-routine will make the info live for 'duration * interval'
		go func(torrentID, msgID int, trackers string) {
			b.liveUpdateWithKeyboard(ctx, ud.Message.Chat.ID, msgID, func() string {
				torrent, err := b.Client.GetTorrent(torrentID)
				if err != nil {
					return "" // skip this iteration if there's an error
				}

				torrentName := b.mdReplacer.Replace(torrent.Name)
				return fmt.Sprintf("`<%d>` *%s*\n%s %s *%.1f%%* (%s / %s) ↓ *%s*  ↑ *%s* R: *%s*\nDL: *%s* UP: *%s*\nAdded: *%s*, ETA: *%s*\nTrackers: `%s`",
					torrent.ID, torrentName, progressBar(torrent.PercentDone, 10), torrent.TorrentStatus(),
					torrent.PercentDone*100, humanize.Bytes(torrent.Have()), humanize.Bytes(torrent.SizeWhenDone),
					humanize.Bytes(torrent.RateDownload), humanize.Bytes(torrent.RateUpload), torrent.Ratio(),
					humanize.Bytes(torrent.DownloadedEver), humanize.Bytes(torrent.UploadedEver), time.Unix(torrent.AddedDate, 0).Format(time.Stamp),
					torrent.ETA(), trackers)
			}, func() string {
				// fetch one final time for the dashes display
				torrent, err := b.Client.GetTorrent(torrentID)
				if err != nil {
					return ""
				}
				torrentName := b.mdReplacer.Replace(torrent.Name)
				return fmt.Sprintf("`<%d>` *%s*\n%s %s *%.1f%%* (%s / %s) ↓ *- B*  ↑ *- B* R: *%s*\nDL: *%s* UP: *%s*\nAdded: *%s*, ETA: *-*\nTrackers: `%s`",
					torrent.ID, torrentName, progressBar(torrent.PercentDone, 10), torrent.TorrentStatus(),
					torrent.PercentDone*100, humanize.Bytes(torrent.Have()), humanize.Bytes(torrent.SizeWhenDone),
					torrent.Ratio(), humanize.Bytes(torrent.DownloadedEver), humanize.Bytes(torrent.UploadedEver),
					time.Unix(torrent.AddedDate, 0).Format(time.Stamp), trackers)
			}, keyboard)
		}(torrentID, msgID, trackers)
	}
}

// infoKeyboard returns inline action buttons for controlling a torrent. The
// callback includes a torrent fingerprint so buttons from an old message
// cannot affect a different torrent that later reuses the same numeric ID.
func infoKeyboard(torrentID int, hash string) *models.InlineKeyboardMarkup {
	fingerprint := torrentFingerprint(hash)
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "⏸ Pause", CallbackData: fmt.Sprintf("cmd:stop:%d:%s", torrentID, fingerprint)},
				{Text: "▶ Resume", CallbackData: fmt.Sprintf("cmd:start:%d:%s", torrentID, fingerprint)},
				{Text: "🗑 Delete", CallbackData: fmt.Sprintf("cmd:del:%d:%s", torrentID, fingerprint)},
			},
		},
	}
}

func deleteConfirmationKeyboard(torrentID int, hash string) *models.InlineKeyboardMarkup {
	fingerprint := torrentFingerprint(hash)
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "⚠️ Confirm delete", CallbackData: fmt.Sprintf("cmd:confirm-del:%d:%s", torrentID, fingerprint)},
				{Text: "Cancel", CallbackData: fmt.Sprintf("cmd:cancel:%d:%s", torrentID, fingerprint)},
			},
		},
	}
}

// stats echo back transmission stats
func (b *Bot) stats(ctx context.Context, ud *models.Update, args []string) {
	stats, err := b.Client.GetStats()
	if err != nil {
		b.Send(ctx, "*stats:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	msg := fmt.Sprintf(
		`
		Total: *%d*
		Active: *%d*
		Paused: *%d*

		_Current Stats_
		Downloaded: *%s*
		Uploaded: *%s*
		Running time: *%s*

		_Accumulative Stats_
		Sessions: *%d*
		Downloaded: *%s*
		Uploaded: *%s*
		Total Running time: *%s*
		`,

		stats.TorrentCount,
		stats.ActiveTorrentCount,
		stats.PausedTorrentCount,
		humanize.Bytes(stats.CurrentStats.DownloadedBytes),
		humanize.Bytes(stats.CurrentStats.UploadedBytes),
		stats.CurrentActiveTime(),
		stats.CumulativeStats.SessionCount,
		humanize.Bytes(stats.CumulativeStats.DownloadedBytes),
		humanize.Bytes(stats.CumulativeStats.UploadedBytes),
		stats.CumulativeActiveTime(),
	)

	b.Send(ctx, msg, ud.Message.Chat.ID, true)
}

// speed will echo back the current download and upload speeds
func (b *Bot) speed(ctx context.Context, ud *models.Update, args []string) {
	stats, err := b.Client.GetStats()
	if err != nil {
		b.Send(ctx, "*speed:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	msg := fmt.Sprintf("↓ %s  ↑ %s", humanize.Bytes(stats.DownloadSpeed), humanize.Bytes(stats.UploadSpeed))

	msgID := b.Send(ctx, msg, ud.Message.Chat.ID, false)

	b.liveUpdate(ctx, ud.Message.Chat.ID, msgID, func() string {
		stats, err := b.Client.GetStats()
		if err != nil {
			return ""
		}
		return fmt.Sprintf("↓ %s  ↑ %s", humanize.Bytes(stats.DownloadSpeed), humanize.Bytes(stats.UploadSpeed))
	}, func() string {
		// show dashes to indicate that we are done updating.
		return "↓ - B  ↑ - B"
	})
}

// count returns current torrents count per status
func (b *Bot) count(ctx context.Context, ud *models.Update, args []string) {
	torrents, err := b.Client.GetTorrents()
	if err != nil {
		b.Send(ctx, "*count:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	var downloading, seeding, stopped, checking, downloadingQ, seedingQ, checkingQ int

	for i := range torrents {
		switch torrents[i].Status {
		case transmission.StatusDownloading:
			downloading++
		case transmission.StatusSeeding:
			seeding++
		case transmission.StatusStopped:
			stopped++
		case transmission.StatusChecking:
			checking++
		case transmission.StatusDownloadPending:
			downloadingQ++
		case transmission.StatusSeedPending:
			seedingQ++
		case transmission.StatusCheckPending:
			checkingQ++
		}
	}

	msg := fmt.Sprintf("Downloading: %d\nSeeding: %d\nPaused: %d\nVerifying: %d\n\n- Waiting to -\nDownload: %d\nSeed: %d\nVerify: %d\n\nTotal: %d",
		downloading, seeding, stopped, checking, downloadingQ, seedingQ, checkingQ, len(torrents))

	b.Send(ctx, msg, ud.Message.Chat.ID, false)
}
