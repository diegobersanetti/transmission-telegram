package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

const Help = `
*Torrent lists*
/list (optional tracker) — all torrents, optionally filtered by tracker (/li, /ls)
/head (optional n) — first n torrents with live speeds (/he)
/tail (optional n) — last n torrents with live speeds (/ta)
/active — actively transferring torrents (/ac)
/downs — downloading and queued torrents (/dg)
/seeding — seeding and queued torrents (/sd)
/paused — stopped torrents (/pa)
/checking — verifying and queued torrents (/ch)
/errors — torrents with errors (/er)
/latest (optional n) — newest torrents; default 5 (/la)
/search <name> — search torrent names (/se)
/info <id...> — details and inline controls (/in)

*Torrent control*
/add <url|magnet...> — add one or more torrents (/ad)
You can also send a .torrent document directly.
/stop <id...|all> — pause torrents (/sp)
/start <id...|all> — resume torrents (/st)
/check <id...|all> — verify torrents (/ck)
/del <id...> — remove torrents, keeping local data (/rm)
/deldata <id...> — permanently remove torrents and local data
/reannounce <id...|all> — contact trackers now (/ra)

*Session and settings*
/speed — current transfer speeds (/ss)
/stats — session and cumulative statistics (/sa)
/count — torrent counts by status (/co)
/trackers — tracker hosts and torrent counts (/tr)
/sort (optional rev) <method> — set listing order (/so)
/turtle (optional on|off) — alternative speed limits (/alt, /tu)
/free (optional path) — available disk space (/space, /disk)
/downlimit <KB/s> — global download limit (/dl)
/uplimit <KB/s> — global upload limit (/ul)
/downloaddir <path> — set the default download directory (/dd)
/version — Transmission and bot versions (/ver)
/help — show this message

Only authorized users are accepted. In groups, Telegram may display commands as /command@BotName.
Be careful with /deldata: it deletes downloaded files immediately.
Report issues: https://github.com/pyed/transmission-telegram`

// Version is overridden by release builds. Tagged go install builds fall back
// to the module version recorded in Go build information.
var Version = "dev"

func init() {
	if Version != "dev" {
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		Version = info.Main.Version
	}
}

type MasterSlice []string

func (masters *MasterSlice) String() string {
	return fmt.Sprintf("%s", *masters)
}

func (masters *MasterSlice) Set(master string) error {
	master = strings.TrimSpace(master)
	if strings.Trim(master, "@") == "" {
		return errors.New("master cannot be empty")
	}
	*masters = append(*masters, strings.ToLower(master))
	return nil
}

func (masters MasterSlice) Contains(username string, userID int64) bool {
	cleanUsername := strings.ToLower(strings.TrimPrefix(username, "@"))
	userIDStr := strconv.FormatInt(userID, 10)

	for _, master := range masters {
		cleanMaster := strings.ToLower(strings.TrimPrefix(master, "@"))
		if cleanUsername != "" && cleanMaster == cleanUsername {
			return true
		}
		if userID > 0 && cleanMaster == userIDStr {
			return true
		}
	}
	return false
}

type Config struct {
	BotToken     string
	Masters      MasterSlice
	RPCURL       string
	Username     string
	Password     string
	LogFile      string
	TransLogFile string
	NoLive       bool
	Interval     time.Duration
	Duration     int
}

func Parse() *Config {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: transmission-telegram <-token=TOKEN> <-master=@tuser|123456789> [-master=@yuser2] [-url=http://] [-username=user] [-password=pass]\n\n")
		flag.PrintDefaults()
	}

	cfg, err := ParseFlags(flag.CommandLine, os.Args[1:], os.Getenv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n\n", err)
		flag.Usage()
		os.Exit(1)
	}

	return cfg
}

func ParseFlags(fs *flag.FlagSet, args []string, getenv func(string) string) (*Config, error) {
	if getenv == nil {
		getenv = os.Getenv
	}

	cfg := &Config{
		Interval: 5 * time.Second,
		Duration: 10,
	}
	tokenDefault := strings.TrimSpace(getenv("TT_BOT_TOKEN"))
	if tokenDefault == "" {
		tokenDefault = strings.TrimSpace(getenv("TT_BOTT"))
	}
	rpcURLDefault := strings.TrimSpace(getenv("TR_URL"))
	if rpcURLDefault == "" {
		rpcURLDefault = "http://localhost:9091/transmission/rpc"
	}
	authUsernameDefault, authPasswordDefault := "", ""
	if values := strings.SplitN(getenv("TR_AUTH"), ":", 2); len(values) == 2 {
		authUsernameDefault, authPasswordDefault = values[0], values[1]
	}

	fs.StringVar(&cfg.BotToken, "token", tokenDefault, "Telegram bot token (env: TT_BOT_TOKEN or legacy TT_BOTT)")
	fs.Var(&cfg.Masters, "master", "Telegram username or numeric user ID allowed to use the bot; repeatable")
	fs.StringVar(&cfg.RPCURL, "url", rpcURLDefault, "Transmission RPC URL (env: TR_URL)")
	fs.StringVar(&cfg.Username, "username", authUsernameDefault, "Transmission username (env: TR_AUTH user:password)")
	fs.StringVar(&cfg.Password, "password", authPasswordDefault, "Transmission password (env: TR_AUTH user:password)")
	fs.StringVar(&cfg.LogFile, "logfile", "", "Send logs to a file")
	fs.StringVar(&cfg.TransLogFile, "transmission-logfile", "", "Open transmission logfile to monitor torrents completion")
	fs.BoolVar(&cfg.NoLive, "no-live", false, "Don't edit and update info after sending")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	cfg.BotToken = strings.TrimSpace(cfg.BotToken)
	cfg.RPCURL = strings.TrimSpace(cfg.RPCURL)
	if cfg.BotToken == "" {
		return nil, errors.New("telegram bot token is required (-token or TT_BOT_TOKEN)")
	}
	if len(cfg.Masters) < 1 {
		return nil, errors.New("at least one authorized user is required (-master)")
	}

	for i := range cfg.Masters {
		cfg.Masters[i] = strings.TrimPrefix(strings.TrimSpace(cfg.Masters[i]), "@")
	}

	return cfg, nil
}
