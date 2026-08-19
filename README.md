# transmission-telegram

[![CI](https://github.com/pyed/transmission-telegram/actions/workflows/ci.yml/badge.svg)](https://github.com/pyed/transmission-telegram/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/pyed/transmission-telegram)](https://goreportcard.com/report/github.com/pyed/transmission-telegram)
[![Release](https://img.shields.io/github/v/release/pyed/transmission-telegram)](https://github.com/pyed/transmission-telegram/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

#### Manage and monitor your Transmission BitTorrent client through Telegram.

<img src="https://raw.github.com/pyed/transmission-telegram/master/demo.gif" width="450" />

---

## Key Features

- ⚡ **Modern Telegram Bot Engine**: Built with [`github.com/go-telegram/bot`](https://github.com/go-telegram/bot) with native graceful shutdown, context propagation, and structured `log/slog` logging.
- 🎛 **Interactive Inline Buttons**: Control torrents (`[ ⏸ Pause ]`, `[ ▶ Resume ]`, `[ 🗑 Delete ]`) directly from `/info` with instant feedback.
- 📊 **Visual Progress Bars**: Clean Unicode progress bars (`[█████░░░░░] 50.0%`) on `/head`, `/tail`, `/active`, `/paused`, `/checking`, and `/info`.
- 🔔 **Polling Completion Alerts**: Automatic completion notifications via Transmission RPC — works across Docker, Kubernetes, NAS, and remote seedboxes without needing local log file access.
- 🐢 **Turtle Mode & Utilities**: Instant speed toggling (`/turtle`), free disk space checker (`/free`), and manual tracker re-announcing (`/reannounce`).
- 🔒 **Flexible Authorization**: Secure `-master` list supporting both Telegram `@usernames` and permanent numeric user IDs.
- 👥 **Group Chat Ready**: Clean handling of group mentions (e.g. `/list@BotName`).
- 📱 **Telegram Menu Auto-Registration**: Automatically configures the Telegram command suggestion menu on startup.

---

## Installation

### Pre-compiled Binaries
Download the pre-compiled binary for Linux, macOS, or Windows from the [Releases](https://github.com/pyed/transmission-telegram/releases) page.

### From Source (Go 1.21+)
```bash
go install github.com/pyed/transmission-telegram/cmd/bot@latest
```

---

## Configuration

`transmission-telegram` is configured via command-line flags or environment variables:

| Flag | Environment Variable | Description |
|---|---|---|
| `-token` | `TT_BOTT` | Telegram Bot Token (*Required*) |
| `-master` | — | Telegram `@username` or numeric user ID (*Required*, repeatable) |
| `-url` | `TR_URL` | Transmission RPC URL (default: `http://localhost:9091/transmission/rpc`) |
| `-username` | `TR_AUTH` (`user:pass`) | Transmission RPC username |
| `-password` | `TR_AUTH` (`user:pass`) | Transmission RPC password |
| `-logfile` | — | Send bot logs to a file (default: stdout) |
| `-transmission-logfile` | — | Path to Transmission daemon log file (optional; defaults to RPC polling) |
| `-no-live` | — | Disable real-time editing and live update loops |

### Example CLI Usage
```bash
transmission-telegram \
  -token="123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ" \
  -master="@YourUsername" \
  -master="123456789" \
  -url="http://localhost:9091/transmission/rpc" \
  -username="admin" \
  -password="password"
```

---

## Available Commands

| Command | Aliases | Description |
|---|---|---|
| `/list [query]` | `/li`, `/ls` | List all torrents (or filter by tracker query) |
| `/head [n]` | `/he` | Show the first *n* torrents with live speed updates (default: 5) |
| `/tail [n]` | `/ta` | Show the last *n* torrents with live speed updates (default: 5) |
| `/active` | `/ac` | List torrents actively uploading or downloading |
| `/downs` | `/dg` | List torrents currently downloading |
| `/seeding` | `/sd` | List torrents currently seeding |
| `/paused` | `/pa` | List paused torrents |
| `/checking` | `/ch` | List torrents verifying checksums |
| `/errors` | `/er` | List torrents with errors |
| `/latest [n]` | `/la` | List newest added torrents |
| `/search <query>` | `/se` | Search torrents by name |
| `/info <id...>` | `/in` | Show details, progress bar, and interactive control buttons |
| `/add <url/magnet>` | `/ad` | Add torrent by URL or magnet. Sending a `.torrent` file directly in chat also adds it |
| `/stop <id...\|all>` | `/sp` | Stop specific torrents or all |
| `/start <id...\|all>` | `/st` | Start specific torrents or all |
| `/check <id...\|all>` | `/ck` | Verify specific torrents or all |
| `/del <id...>` | `/rm` | Remove torrent from Transmission |
| `/deldata <id...>` | — | Remove torrent and delete local files from disk |
| `/turtle [on\|off]` | `/alt`, `/tu` | Toggle or set alternative speed limits ("Turtle Mode") |
| `/free [path]` | `/space`, `/disk` | Show free and total disk space on the download drive |
| `/reannounce <id...\|all>` | `/ra` | Force tracker re-announce |
| `/speed` | `/ss` | Display current download and upload speeds |
| `/stats` | `/sa` | Show Transmission session and cumulative statistics |
| `/downlimit <KB/s>` | `/dl` | Set global download speed limit |
| `/uplimit <KB/s>` | `/ul` | Set global upload speed limit |
| `/downloaddir <path>` | `/dd` | Set default download directory |
| `/sort <method>` | `/so` | Sort torrent listings (`id`, `name`, `age`, `size`, `progress`, `downspeed`, `upspeed`, `ratio`, prefix with `rev` for reverse) |
| `/trackers` | `/tr` | List trackers and torrent counts |
| `/count` | `/co` | Show torrent counts grouped by status |
| `/help` | — | Display command help |
| `/version` | `/ver` | Display Transmission daemon and bot version |

---

## Docker

### Standalone
```bash
docker run -d --name transmission-telegram \
  --restart unless-stopped \
  -e TT_BOTT="<YOUR_BOT_TOKEN>" \
  -e TR_AUTH="<username>:<password>" \
  pyed/transmission-telegram:latest \
  -master="@YourUsername" \
  -url="http://transmission:9091/transmission/rpc"
```

### Docker Compose
```yaml
version: '3.8'

services:
  transmission:
    image: linuxserver/transmission:latest
    container_name: transmission
    environment:
      - PUID=1000
      - PGID=1000
      - TZ=Etc/UTC
    volumes:
      - /path/to/config:/config
      - /path/to/downloads:/downloads
    ports:
      - "9091:9091"
      - "51413:51413"
      - "51413:51413/udp"
    restart: unless-stopped

  telegram-bot:
    image: pyed/transmission-telegram:latest
    container_name: transmission-telegram
    depends_on:
      - transmission
    restart: unless-stopped
    command: >
      -token="YOUR_BOT_TOKEN"
      -master="@YourTelegramUsername"
      -url="http://transmission:9091/transmission/rpc"
      -username="admin"
      -password="password"
```

---

## License

[MIT](LICENSE)
