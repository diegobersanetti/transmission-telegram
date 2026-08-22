package bot

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/go-telegram/bot/models"
	"github.com/pyed/transmission"
)

// sortUsage is the usage message for the /sort command.
const sortUsage = `*sort* takes one of:
			(*id, name, age, size, progress, downspeed, upspeed, download, upload, ratio*)
			optionally start with (*rev*) for reversed order
			e.g. "*sort rev size*" to get biggest torrents first.`

// sort changes torrents sorting
func (b *Bot) sort(ctx context.Context, ud *models.Update, args []string) {
	if len(args) == 0 {
		b.Send(ctx, sortUsage, ud.Message.Chat.ID, true)
		return
	}

	var reversed bool
	if strings.ToLower(args[0]) == "rev" {
		reversed = true
		args = args[1:]
		// "/sort rev" without a key: used to panic on args[0] below,
		// crashing the whole process. Reply with usage instead.
		if len(args) == 0 {
			b.Send(ctx, sortUsage, ud.Message.Chat.ID, true)
			return
		}
	}

	switch strings.ToLower(args[0]) {
	case "id":
		if reversed {
			b.Client.SetSort(transmission.SortRevID)
			break
		}
		b.Client.SetSort(transmission.SortID)
	case "name":
		if reversed {
			b.Client.SetSort(transmission.SortRevName)
			break
		}
		b.Client.SetSort(transmission.SortName)
	case "age":
		if reversed {
			b.Client.SetSort(transmission.SortRevAge)
			break
		}
		b.Client.SetSort(transmission.SortAge)
	case "size":
		if reversed {
			b.Client.SetSort(transmission.SortRevSize)
			break
		}
		b.Client.SetSort(transmission.SortSize)
	case "progress":
		if reversed {
			b.Client.SetSort(transmission.SortRevProgress)
			break
		}
		b.Client.SetSort(transmission.SortProgress)
	case "downspeed":
		if reversed {
			b.Client.SetSort(transmission.SortRevDownSpeed)
			break
		}
		b.Client.SetSort(transmission.SortDownSpeed)
	case "upspeed":
		if reversed {
			b.Client.SetSort(transmission.SortRevUpSpeed)
			break
		}
		b.Client.SetSort(transmission.SortUpSpeed)
	case "download":
		if reversed {
			b.Client.SetSort(transmission.SortRevDownloaded)
			break
		}
		b.Client.SetSort(transmission.SortDownloaded)
	case "upload":
		if reversed {
			b.Client.SetSort(transmission.SortRevUploaded)
			break
		}
		b.Client.SetSort(transmission.SortUploaded)
	case "ratio":
		if reversed {
			b.Client.SetSort(transmission.SortRevRatio)
			break
		}
		b.Client.SetSort(transmission.SortRatio)
	default:
		b.Send(ctx, "unknown sorting method", ud.Message.Chat.ID, false)
		return
	}

	if reversed {
		b.Send(ctx, "*sort:* reversed "+args[0], ud.Message.Chat.ID, false)
		return
	}
	b.Send(ctx, "*sort:* "+args[0], ud.Message.Chat.ID, false)
}

// trackers will send a list of trackers and how many torrents each one has
func (b *Bot) trackers(ctx context.Context, ud *models.Update, args []string) {
	torrents, err := b.Client.GetTorrents()
	if err != nil {
		b.Send(ctx, "*trackers:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	trackerMap := make(map[string]int)

	for i := range torrents {
		for _, tracker := range torrents[i].Trackers {
			sm := b.trackerRegex.FindSubmatch([]byte(tracker.Announce))
			if len(sm) > 1 {
				currentTracker := string(sm[1])
				trackerMap[currentTracker]++
			}
		}
	}

	buf := new(bytes.Buffer)
	for k, v := range trackerMap {
		buf.WriteString(fmt.Sprintf("%d - %s\n", v, k))
	}

	if buf.Len() == 0 {
		b.Send(ctx, "No trackers!", ud.Message.Chat.ID, false)
		return
	}
	b.Send(ctx, buf.String(), ud.Message.Chat.ID, false)
}

// downloaddir takes a path and sets it as the download directory
func (b *Bot) downloaddir(ctx context.Context, ud *models.Update, args []string) {
	if len(args) < 1 {
		b.Send(ctx, "Please, specify a path for downloaddir", ud.Message.Chat.ID, false)
		return
	}

	downloadDir := args[0]

	cmd := transmission.NewSessionSetCommand()
	cmd.SetDownloadDir(downloadDir)

	out, err := b.Client.ExecuteCommand(cmd)
	if err != nil {
		b.Send(ctx, "*downloaddir:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}
	if out.Result != "success" {
		b.Send(ctx, "*downloaddir:* "+out.Result, ud.Message.Chat.ID, false)
		return
	}

	b.Send(
		ctx,
		"*downloaddir:* downloaddir has been successfully changed to "+downloadDir,
		ud.Message.Chat.ID, false,
	)
}

// downlimit sets the global downlimit to a provided value in kilobytes
func (b *Bot) downlimit(ctx context.Context, ud *models.Update, args []string) {
	b.speedLimit(ctx, ud, args, transmission.DownloadLimitType)
}

// uplimit sets the global uplimit to a provided value in kilobytes
func (b *Bot) uplimit(ctx context.Context, ud *models.Update, args []string) {
	b.speedLimit(ctx, ud, args, transmission.UploadLimitType)
}

// speedLimit sets either a download or upload limit
func (b *Bot) speedLimit(ctx context.Context, ud *models.Update, args []string, limitType transmission.SpeedLimitType) {
	if len(args) < 1 {
		b.Send(ctx, "Please, specify the limit", ud.Message.Chat.ID, false)
		return
	}

	limit, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		b.Send(ctx, "Please, specify the limit as number of kilobytes", ud.Message.Chat.ID, false)
		return
	}

	speedLimitCmd := transmission.NewSpeedLimitCommand(limitType, uint(limit))
	if speedLimitCmd == nil {
		b.Send(ctx, fmt.Sprintf("*%s:* internal error", limitType), ud.Message.Chat.ID, false)
		return
	}

	out, err := b.Client.ExecuteCommand(speedLimitCmd)
	if err != nil {
		b.Send(ctx, fmt.Sprintf("*%s:* %v", limitType, err.Error()), ud.Message.Chat.ID, false)
		return
	}
	if out.Result != "success" {
		b.Send(ctx, fmt.Sprintf("*%s:* %v", limitType, out.Result), ud.Message.Chat.ID, false)
		return
	}

	b.Send(
		ctx,
		fmt.Sprintf("*%s:* limit has been successfully changed to %d KB/s", limitType, limit),
		ud.Message.Chat.ID, false,
	)
}

// turtle toggles or sets alternative speed limits ("Turtle Mode")
func (b *Bot) turtle(ctx context.Context, ud *models.Update, args []string) {
	sess, err := b.Client.GetSession(ctx)
	if err != nil {
		b.Send(ctx, "*turtle:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	enabled := !sess.AltSpeedEnabled
	if len(args) > 0 {
		switch strings.ToLower(args[0]) {
		case "on", "1", "true", "enable":
			enabled = true
		case "off", "0", "false", "disable":
			enabled = false
		}
	}

	if err := b.Client.SetAltSpeedEnabled(ctx, enabled); err != nil {
		b.Send(ctx, "*turtle:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	if enabled {
		b.Send(ctx, fmt.Sprintf("🐢 *Turtle mode enabled* (Limit: ↓ %d KB/s  ↑ %d KB/s)", sess.AltSpeedDown, sess.AltSpeedUp), ud.Message.Chat.ID, true)
	} else {
		b.Send(ctx, "🐇 *Turtle mode disabled*", ud.Message.Chat.ID, true)
	}
}

// free checks and displays available disk space for the download directory
func (b *Bot) free(ctx context.Context, ud *models.Update, args []string) {
	path := ""
	if len(args) > 0 {
		path = args[0]
	} else {
		sess, err := b.Client.GetSession(ctx)
		if err != nil {
			b.Send(ctx, "*free:* "+err.Error(), ud.Message.Chat.ID, false)
			return
		}
		path = sess.DownloadDir
	}

	freeBytes, totalBytes, err := b.Client.FreeSpace(ctx, path)
	if err != nil {
		b.Send(ctx, "*free:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	var msg string
	if totalBytes > 0 {
		pctFree := (float64(freeBytes) / float64(totalBytes)) * 100.0
		msg = fmt.Sprintf("💾 *Directory:* `%s`\n*Free Space:* *%s* of *%s* (%.1f%% free)",
			b.mdReplacer.Replace(path), humanize.Bytes(uint64(freeBytes)), humanize.Bytes(uint64(totalBytes)), pctFree)
	} else {
		msg = fmt.Sprintf("💾 *Directory:* `%s`\n*Free Space:* *%s*",
			b.mdReplacer.Replace(path), humanize.Bytes(uint64(freeBytes)))
	}

	b.Send(ctx, msg, ud.Message.Chat.ID, true)
}

// reannounce forces an immediate tracker re-announce
func (b *Bot) reannounce(ctx context.Context, ud *models.Update, args []string) {
	if len(args) == 0 {
		b.Send(ctx, "*reannounce:* please provide torrent IDs or _all_", ud.Message.Chat.ID, true)
		return
	}

	if strings.ToLower(args[0]) == "all" {
		torrents, err := b.Client.GetTorrents(ctx)
		if err != nil {
			b.Send(ctx, "*reannounce:* "+err.Error(), ud.Message.Chat.ID, false)
			return
		}
		if err := b.Client.ReannounceTorrents(ctx, torrents.GetIDs()...); err != nil {
			b.Send(ctx, "*reannounce:* "+err.Error(), ud.Message.Chat.ID, false)
			return
		}
		b.Send(ctx, "📡 Re-announced all torrents to trackers", ud.Message.Chat.ID, false)
		return
	}

	var ids []int
	for _, idStr := range args {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		b.Send(ctx, "*reannounce:* no valid IDs provided", ud.Message.Chat.ID, false)
		return
	}

	if err := b.Client.ReannounceTorrents(ctx, ids...); err != nil {
		b.Send(ctx, "*reannounce:* "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	b.Send(ctx, fmt.Sprintf("📡 Re-announced %d torrent(s) to trackers", len(ids)), ud.Message.Chat.ID, false)
}
