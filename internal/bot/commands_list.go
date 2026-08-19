package bot

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/go-telegram/bot/models"
)

// list takes an optional argument which is a query to match against trackers
// to list only torrents that has a tracker that matches.
func (b *Bot) list(ctx context.Context, ud *models.Update, args []string) {
	torrents, err := b.Client.GetTorrents()
	if err != nil {
		b.Send(ctx, "*list:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	// if it gets a query, it will list torrents that has trackers that match the query
	if len(args) != 0 {
		// (?i) for case insensitivity
		regx, err := regexp.Compile("(?i)" + regexp.QuoteMeta(args[0]))
		if err != nil {
			b.Send(ctx, "*list:* "+err.Error(), ud.Message.Chat.ID, false)
			return
		}

		for i := range torrents {
			if regx.MatchString(torrents[i].GetTrackers()) {
				buf.WriteString(fmt.Sprintf("<%d> %s\n", torrents[i].ID, torrents[i].Name))
			}
		}
	} else { // if we did not get a query, list all torrents
		for i := range torrents {
			buf.WriteString(fmt.Sprintf("<%d> %s\n", torrents[i].ID, torrents[i].Name))
		}
	}

	if buf.Len() == 0 {
		// if we got a tracker query show different message
		if len(args) != 0 {
			b.Send(ctx, fmt.Sprintf("*list:* No tracker matches: *%s*", args[0]), ud.Message.Chat.ID, true)
			return
		}
		b.Send(ctx, "*list:* no torrents", ud.Message.Chat.ID, false)
		return
	}

	b.Send(ctx, buf.String(), ud.Message.Chat.ID, false)
}

// head will list the first 5 or n torrents
func (b *Bot) head(ctx context.Context, ud *models.Update, args []string) {
	var (
		n   = 5 // default to 5
		err error
	)

	if len(args) > 0 {
		n, err = strconv.Atoi(args[0])
		if err != nil {
			b.Send(ctx, "*head:* argument must be a number", ud.Message.Chat.ID, false)
			return
		}
	}

	torrents, err := b.Client.GetTorrents()
	if err != nil {
		b.Send(ctx, "*head:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	// make sure that we stay in the boundaries
	if n <= 0 || n > len(torrents) {
		n = len(torrents)
	}

	buf := new(bytes.Buffer)
	for i := range torrents[:n] {
		torrentName := b.mdReplacer.Replace(torrents[i].Name) // escape markdown
		buf.WriteString(fmt.Sprintf("`<%d>` *%s*\n%s %s *%.1f%%* (%s / %s) ↓ *%s*  ↑ *%s* R: *%s*\n\n",
			torrents[i].ID, torrentName, progressBar(torrents[i].PercentDone, 10), torrents[i].TorrentStatus(),
			torrents[i].PercentDone*100, humanize.Bytes(torrents[i].Have()), humanize.Bytes(torrents[i].SizeWhenDone),
			humanize.Bytes(torrents[i].RateDownload), humanize.Bytes(torrents[i].RateUpload), torrents[i].Ratio()))
	}

	if buf.Len() == 0 {
		b.Send(ctx, "*head:* no torrents", ud.Message.Chat.ID, false)
		return
	}

	msgID := b.Send(ctx, buf.String(), ud.Message.Chat.ID, true)

	b.liveUpdate(ctx, ud.Message.Chat.ID, msgID, func() string {
		buf.Reset()
		torrents, err = b.Client.GetTorrents()
		if err != nil {
			return "" // try again next iteration
		}
		if len(torrents) < 1 {
			return ""
		}
		// make sure that we stay in the boundaries
		if n <= 0 || n > len(torrents) {
			n = len(torrents)
		}
		for _, torrent := range torrents[:n] {
			torrentName := b.mdReplacer.Replace(torrent.Name)
			buf.WriteString(fmt.Sprintf("`<%d>` *%s*\n%s %s *%.1f%%* (%s / %s) ↓ *%s*  ↑ *%s* R: *%s*\n\n",
				torrent.ID, torrentName, progressBar(torrent.PercentDone, 10), torrent.TorrentStatus(),
				torrent.PercentDone*100, humanize.Bytes(torrent.Have()), humanize.Bytes(torrent.SizeWhenDone),
				humanize.Bytes(torrent.RateDownload), humanize.Bytes(torrent.RateUpload), torrent.Ratio()))
		}
		return buf.String()
	}, nil)
}

// tail lists the last 5 or n torrents
func (b *Bot) tail(ctx context.Context, ud *models.Update, args []string) {
	var (
		n   = 5 // default to 5
		err error
	)

	if len(args) > 0 {
		n, err = strconv.Atoi(args[0])
		if err != nil {
			b.Send(ctx, "*tail:* argument must be a number", ud.Message.Chat.ID, false)
			return
		}
	}

	torrents, err := b.Client.GetTorrents()
	if err != nil {
		b.Send(ctx, "*tail:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	// make sure that we stay in the boundaries
	if n <= 0 || n > len(torrents) {
		n = len(torrents)
	}

	buf := new(bytes.Buffer)
	for _, torrent := range torrents[len(torrents)-n:] {
		torrentName := b.mdReplacer.Replace(torrent.Name) // escape markdown
		buf.WriteString(fmt.Sprintf("`<%d>` *%s*\n%s %s *%.1f%%* (%s / %s) ↓ *%s*  ↑ *%s* R: *%s*\n\n",
			torrent.ID, torrentName, progressBar(torrent.PercentDone, 10), torrent.TorrentStatus(),
			torrent.PercentDone*100, humanize.Bytes(torrent.Have()), humanize.Bytes(torrent.SizeWhenDone),
			humanize.Bytes(torrent.RateDownload), humanize.Bytes(torrent.RateUpload), torrent.Ratio()))
	}

	if buf.Len() == 0 {
		b.Send(ctx, "*tail:* no torrents", ud.Message.Chat.ID, false)
		return
	}

	msgID := b.Send(ctx, buf.String(), ud.Message.Chat.ID, true)

	b.liveUpdate(ctx, ud.Message.Chat.ID, msgID, func() string {
		buf.Reset()
		torrents, err = b.Client.GetTorrents()
		if err != nil {
			return ""
		}
		if len(torrents) < 1 {
			return ""
		}
		if n <= 0 || n > len(torrents) {
			n = len(torrents)
		}
		for _, torrent := range torrents[len(torrents)-n:] {
			torrentName := b.mdReplacer.Replace(torrent.Name)
			buf.WriteString(fmt.Sprintf("`<%d>` *%s*\n%s %s *%.1f%%* (%s / %s) ↓ *%s*  ↑ *%s* R: *%s*\n\n",
				torrent.ID, torrentName, progressBar(torrent.PercentDone, 10), torrent.TorrentStatus(),
				torrent.PercentDone*100, humanize.Bytes(torrent.Have()), humanize.Bytes(torrent.SizeWhenDone),
				humanize.Bytes(torrent.RateDownload), humanize.Bytes(torrent.RateUpload), torrent.Ratio()))
		}
		return buf.String()
	}, nil)
}

// latest takes n and returns the latest n torrents
func (b *Bot) latest(ctx context.Context, ud *models.Update, args []string) {
	var (
		n   = 5 // default to 5
		err error
	)

	if len(args) > 0 {
		n, err = strconv.Atoi(args[0])
		if err != nil {
			b.Send(ctx, "*latest:* argument must be a number", ud.Message.Chat.ID, false)
			return
		}
	}

	torrents, err := b.Client.GetTorrents()
	if err != nil {
		b.Send(ctx, "*latest:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	// make sure that we stay in the boundaries
	if n <= 0 || n > len(torrents) {
		n = len(torrents)
	}

	// sort by age, and set reverse to true to get the latest first
	torrents.SortAge(true)

	buf := new(bytes.Buffer)
	for i := range torrents[:n] {
		buf.WriteString(fmt.Sprintf("<%d> %s\n", torrents[i].ID, torrents[i].Name))
	}
	if buf.Len() == 0 {
		b.Send(ctx, "*latest:* No torrents", ud.Message.Chat.ID, false)
		return
	}
	b.Send(ctx, buf.String(), ud.Message.Chat.ID, false)
}

// search takes a query and returns torrents with match
func (b *Bot) search(ctx context.Context, ud *models.Update, args []string) {
	// make sure that we got a query
	if len(args) == 0 {
		b.Send(ctx, "*search:* needs an argument", ud.Message.Chat.ID, false)
		return
	}

	query := strings.Join(args, " ")
	// "(?i)" for case insensitivity
	regx, err := regexp.Compile("(?i)" + regexp.QuoteMeta(query))
	if err != nil {
		b.Send(ctx, "*search:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	torrents, err := b.Client.GetTorrents()
	if err != nil {
		b.Send(ctx, "*search:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for i := range torrents {
		if regx.MatchString(torrents[i].Name) {
			buf.WriteString(fmt.Sprintf("<%d> %s\n", torrents[i].ID, torrents[i].Name))
		}
	}
	if buf.Len() == 0 {
		b.Send(ctx, "No matches!", ud.Message.Chat.ID, false)
		return
	}
	b.Send(ctx, buf.String(), ud.Message.Chat.ID, false)
}
