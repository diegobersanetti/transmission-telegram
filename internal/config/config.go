package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	VERSION = "v1.4.1"

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

type MasterSlice []string

func (masters *MasterSlice) String() string {
	return fmt.Sprintf("%s", *masters)
}

func (masters *MasterSlice) Set(master string) error {
	*masters = append(*masters, strings.ToLower(master))
	return nil
}

func (masters MasterSlice) Contains(username string) bool {
	username = strings.ToLower(username)
	for _, master := range masters {
		if master == username {
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
	cfg := &Config{
		Interval: 5 * time.Second,
		Duration: 10,
	}

	flag.StringVar(&cfg.BotToken, "token", "", "Telegram bot token, Can be passed via environment variable 'TT_BOTT'")
	flag.Var(&cfg.Masters, "master", "Your telegram handler, So the bot will only respond to you. Can specify more than one")
	flag.StringVar(&cfg.RPCURL, "url", "http://localhost:9091/transmission/rpc", "Transmission RPC URL")
	flag.StringVar(&cfg.Username, "username", "", "Transmission username")
	flag.StringVar(&cfg.Password, "password", "", "Transmission password")
	flag.StringVar(&cfg.LogFile, "logfile", "", "Send logs to a file")
	flag.StringVar(&cfg.TransLogFile, "transmission-logfile", "", "Open transmission logfile to monitor torrents completion")
	flag.BoolVar(&cfg.NoLive, "no-live", false, "Don't edit and update info after sending")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: transmission-telegram <-token=TOKEN> <-master=@tuser> [-master=@yuser2] [-url=http://] [-username=user] [-password=pass]\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if cfg.BotToken == "" {
		if token := os.Getenv("TT_BOTT"); len(token) > 1 {
			cfg.BotToken = token
		}
	}

	if cfg.BotToken == "" || len(cfg.Masters) < 1 {
		fmt.Fprintf(os.Stderr, "Error: Mandatory argument missing! (-token or -master)\n\n")
		flag.Usage()
		os.Exit(1)
	}

	for i := range cfg.Masters {
		cfg.Masters[i] = strings.Replace(cfg.Masters[i], "@", "", -1)
	}

	if cfg.Username == "" {
		if values := strings.SplitN(os.Getenv("TR_AUTH"), ":", 2); len(values) > 1 {
			cfg.Username, cfg.Password = values[0], values[1]
		}
	}

	return cfg
}
