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

const (
	HELP = `
	*list* or *li* or *ls*
	Lists all the torrents, takes an optional argument which is a query to list only torrents that has a tracker matches the query, or some of it.

	*head* or *he*
	Lists the first n number of torrents, n defaults to 5 if no argument is provided.

	*tail* or *ta*
	Lists the last n number of torrents, n defaults to 5 if no argument is provided.

	*downs* or *dg*
	Lists torrents with the status of _Downloading_ or in the queue to download.

	*seeding* or *sd*
	Lists torrents with the status of _Seeding_ or in the queue to seed.

	*paused* or *pa*
	Lists _Paused_ torrents.

	*checking* or *ch*
	Lists torrents with the status of _Verifying_ or in the queue to verify.

	*active* or *ac*
	Lists torrents that are actively uploading or downloading.

	*errors* or *er*
	Lists torrents with with errors along with the error message.

	*sort* or *so*
	Manipulate the sorting of the aforementioned commands. Call it without arguments for more.

	*trackers* or *tr*
	Lists all the trackers along with the number of torrents.

	*downloaddir* or *dd*
	Set download directory to the specified path. Transmission will automatically create a
	directory in case you provided an inexistent one.

	*add* or *ad*
	Takes one or many URLs or magnets to add them. You can send a ".torrent" file via Telegram to add it.

	*search* or *se*
	Takes a query and lists torrents with matching names.

	*latest* or *la*
	Lists the newest n torrents, n defaults to 5 if no argument is provided.

	*info* or *in*
	Takes one or more torrent's IDs to list more info about them.

	*stop* or *sp*
	Takes one or more torrent's IDs to stop them, or _all_ to stop all torrents.

	*start* or *st*
	Takes one or more torrent's IDs to start them, or _all_ to start all torrents.

	*check* or *ck*
	Takes one or more torrent's IDs to verify them, or _all_ to verify all torrents.

	*del* or *rm*
	Takes one or more torrent's IDs to delete them.

	*deldata*
	Takes one or more torrent's IDs to delete them and their data.

	*stats* or *sa*
	Shows Transmission's stats.

	*downlimit* or *dl*
	Set global limit for download speed in kilobytes.

	*uplimit* or *ul*
	Set global limit for upload speed in kilobytes.

	*speed* or *ss*
	Shows the upload and download speeds.

	*turtle* or *alt* or *tu*
	Toggle alternative speed limits ("Turtle Mode") on or off.

	*free* or *space* or *disk*
	Shows available and total storage space for the download directory.

	*reannounce* or *ra*
	Forces an immediate tracker re-announce for specified torrent IDs or _all_.

	*count* or *co*
	Shows the torrents counts per status.

	*help*
	Shows this help message.

	*version* or *ver*
	Shows version numbers.

	- Prefix commands with '/' if you want to talk to your bot in a group. 
	- report any issues [here](https://github.com/pyed/transmission-telegram)
	`
)

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
		return nil, errors.New("Telegram bot token is required (-token or TT_BOT_TOKEN)")
	}
	if len(cfg.Masters) < 1 {
		return nil, errors.New("at least one authorized user is required (-master)")
	}

	for i := range cfg.Masters {
		cfg.Masters[i] = strings.TrimPrefix(strings.TrimSpace(cfg.Masters[i]), "@")
	}

	return cfg, nil
}
