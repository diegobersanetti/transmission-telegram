package bot

import (
	"fmt"

	"github.com/pyed/transmission-telegram/internal/config"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

// registerCommands builds the command dispatch map.
// All existing command names and aliases are preserved.
func (b *Bot) registerCommands() {
	b.commands = map[string]commandFunc{
		"list": b.list, "li": b.list, "ls": b.list,
		"head": b.head, "he": b.head,
		"tail": b.tail, "ta": b.tail,
		"downs": b.downs, "dg": b.downs,
		"seeding": b.seeding, "sd": b.seeding,
		"paused": b.paused, "pa": b.paused,
		"checking": b.checking, "ch": b.checking,
		"active": b.active, "ac": b.active,
		"errors": b.errors, "er": b.errors,
		"sort": b.sort, "so": b.sort,
		"trackers": b.trackers, "tr": b.trackers,
		"downloaddir": b.downloaddir, "dd": b.downloaddir,
		"add": b.add, "ad": b.add,
		"search": b.search, "se": b.search,
		"latest": b.latest, "la": b.latest,
		"info": b.info, "in": b.info,
		"stop": b.stop, "sp": b.stop,
		"start": b.start, "st": b.start,
		"check": b.check, "ck": b.check,
		"stats": b.stats, "sa": b.stats,
		"downlimit": b.downlimit, "dl": b.downlimit,
		"uplimit": b.uplimit, "ul": b.uplimit,
		"speed": b.speed, "ss": b.speed,
		"count": b.count, "co": b.count,
		"del": b.del, "rm": b.del,
		"deldata": b.deldata,
		"help": b.help,
		"version": b.version, "ver": b.version,
	}
}

func (b *Bot) help(ud tgbotapi.Update, args []string) {
	b.Send(config.HELP, ud.Message.Chat.ID, true)
}

func (b *Bot) version(ud tgbotapi.Update, args []string) {
	b.Send(fmt.Sprintf("Transmission *%s*\nTransmission-telegram *%s*", b.Client.Version(), config.VERSION), ud.Message.Chat.ID, true)
}
