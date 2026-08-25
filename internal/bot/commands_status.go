package bot

import (
	"bytes"
	"context"
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/go-telegram/bot/models"
	"github.com/pyed/transmission"
)

// downs will send the names of torrents with status 'Downloading' or in queue to
func (b *Bot) downs(ctx context.Context, ud *models.Update, args []string) {
	torrents, err := b.Client.GetTorrents(ctx)
	if err != nil {
		b.Send(ctx, "*downs:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for i := range torrents {
		// Downloading or in queue to download
		if torrents[i].Status == transmission.StatusDownloading ||
			torrents[i].Status == transmission.StatusDownloadPending {
			buf.WriteString(fmt.Sprintf("<%d> %s\n", torrents[i].ID, torrents[i].Name))
		}
	}

	if buf.Len() == 0 {
		b.Send(ctx, "No downloads", ud.Message.Chat.ID, false)
		return
	}
	b.Send(ctx, buf.String(), ud.Message.Chat.ID, false)
}

// seeding will send the names of the torrents with the status 'Seeding' or in the queue to
func (b *Bot) seeding(ctx context.Context, ud *models.Update, args []string) {
	torrents, err := b.Client.GetTorrents(ctx)
	if err != nil {
		b.Send(ctx, "*seeding:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for i := range torrents {
		if torrents[i].Status == transmission.StatusSeeding ||
			torrents[i].Status == transmission.StatusSeedPending {
			buf.WriteString(fmt.Sprintf("<%d> %s\n", torrents[i].ID, torrents[i].Name))
		}
	}

	if buf.Len() == 0 {
		b.Send(ctx, "No torrents seeding", ud.Message.Chat.ID, false)
		return
	}

	b.Send(ctx, buf.String(), ud.Message.Chat.ID, false)
}

// paused will send the names of the torrents with status 'Paused'
func (b *Bot) paused(ctx context.Context, ud *models.Update, args []string) {
	torrents, err := b.Client.GetTorrents(ctx)
	if err != nil {
		b.Send(ctx, "*paused:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for i := range torrents {
		if torrents[i].Status == transmission.StatusStopped {
			buf.WriteString(fmt.Sprintf("<%d> %s\n%s %s (%.1f%%) DL: %s UL: %s  R: %s\n\n",
				torrents[i].ID, torrents[i].Name, progressBar(torrents[i].PercentDone, 10), torrents[i].TorrentStatus(),
				torrents[i].PercentDone*100, humanize.Bytes(torrents[i].DownloadedEver),
				humanize.Bytes(torrents[i].UploadedEver), torrents[i].Ratio()))
		}
	}

	if buf.Len() == 0 {
		b.Send(ctx, "No paused torrents", ud.Message.Chat.ID, false)
		return
	}

	b.Send(ctx, buf.String(), ud.Message.Chat.ID, false)
}

// checking will send the names of torrents with the status 'verifying' or in the queue to
func (b *Bot) checking(ctx context.Context, ud *models.Update, args []string) {
	torrents, err := b.Client.GetTorrents(ctx)
	if err != nil {
		b.Send(ctx, "*checking:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for i := range torrents {
		if torrents[i].Status == transmission.StatusChecking ||
			torrents[i].Status == transmission.StatusCheckPending {
			buf.WriteString(fmt.Sprintf("<%d> %s\n%s %s (%.1f%%)\n\n",
				torrents[i].ID, torrents[i].Name, progressBar(torrents[i].PercentDone, 10), torrents[i].TorrentStatus(),
				torrents[i].PercentDone*100))
		}
	}

	if buf.Len() == 0 {
		b.Send(ctx, "No torrents verifying", ud.Message.Chat.ID, false)
		return
	}

	b.Send(ctx, buf.String(), ud.Message.Chat.ID, false)
}

// active will send torrents that are actively downloading or uploading
func (b *Bot) active(ctx context.Context, ud *models.Update, args []string) {
	torrents, err := b.Client.GetTorrents(ctx)
	if err != nil {
		b.Send(ctx, "*active:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for i := range torrents {
		if torrents[i].RateDownload > 0 ||
			torrents[i].RateUpload > 0 {
			torrentName := b.mdReplacer.Replace(torrents[i].Name)
			buf.WriteString(fmt.Sprintf("`<%d>` *%s*\n%s %s *%.1f%%* (%s / %s) ↓ *%s*  ↑ *%s* R: *%s*\n\n",
				torrents[i].ID, torrentName, progressBar(torrents[i].PercentDone, 10), torrents[i].TorrentStatus(),
				torrents[i].PercentDone*100, humanize.Bytes(torrents[i].Have()),
				humanize.Bytes(torrents[i].SizeWhenDone), humanize.Bytes(torrents[i].RateDownload),
				humanize.Bytes(torrents[i].RateUpload), torrents[i].Ratio()))
		}
	}
	if buf.Len() == 0 {
		b.Send(ctx, "No active torrents", ud.Message.Chat.ID, false)
		return
	}

	msgID := b.Send(ctx, buf.String(), ud.Message.Chat.ID, true)

	b.liveUpdate(ctx, ud.Message.Chat.ID, msgID, func() string {
		buf.Reset()
		torrents, err = b.Client.GetTorrents(ctx)
		if err != nil {
			return "" // if there was error getting torrents, skip to the next iteration
		}
		for i := range torrents {
			if torrents[i].RateDownload > 0 ||
				torrents[i].RateUpload > 0 {
				torrentName := b.mdReplacer.Replace(torrents[i].Name)
				buf.WriteString(fmt.Sprintf("`<%d>` *%s*\n%s %s *%.1f%%* (%s / %s) ↓ *%s*  ↑ *%s* R: *%s*\n\n",
					torrents[i].ID, torrentName, progressBar(torrents[i].PercentDone, 10), torrents[i].TorrentStatus(),
					torrents[i].PercentDone*100, humanize.Bytes(torrents[i].Have()),
					humanize.Bytes(torrents[i].SizeWhenDone), humanize.Bytes(torrents[i].RateDownload),
					humanize.Bytes(torrents[i].RateUpload), torrents[i].Ratio()))
			}
		}
		return buf.String()
	}, func() string {
		// replace the speed with dashes to indicate that we are done being live
		buf.Reset()
		torrents, err = b.Client.GetTorrents(ctx)
		if err != nil {
			return ""
		}
		for i := range torrents {
			if torrents[i].RateDownload > 0 ||
				torrents[i].RateUpload > 0 {
				torrentName := b.mdReplacer.Replace(torrents[i].Name)
				buf.WriteString(fmt.Sprintf("`<%d>` *%s*\n%s %s *%.1f%%* (%s / %s) ↓ *-*  ↑ *-* R: *%s*\n\n",
					torrents[i].ID, torrentName, progressBar(torrents[i].PercentDone, 10), torrents[i].TorrentStatus(),
					torrents[i].PercentDone*100, humanize.Bytes(torrents[i].Have()),
					humanize.Bytes(torrents[i].SizeWhenDone), torrents[i].Ratio()))
			}
		}
		return buf.String()
	})
}

// errors will send torrents with errors
func (b *Bot) errors(ctx context.Context, ud *models.Update, args []string) {
	torrents, err := b.Client.GetTorrents(ctx)
	if err != nil {
		b.Send(ctx, "*errors:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for i := range torrents {
		if torrents[i].Error != 0 {
			buf.WriteString(fmt.Sprintf("<%d> %s\n%s\n",
				torrents[i].ID, torrents[i].Name, torrents[i].ErrorString))
		}
	}
	if buf.Len() == 0 {
		b.Send(ctx, "No errors", ud.Message.Chat.ID, false)
		return
	}
	b.Send(ctx, buf.String(), ud.Message.Chat.ID, false)
}
