# transmission-telegram

#### Manage your Transmission BitTorrent client through Telegram.

<img src="https://raw.github.com/pyed/transmission-telegram/master/demo.gif" width="400" />

## Installation

### Binary Releases
Download the pre-compiled binary for your OS and architecture from the [Releases](https://github.com/pyed/transmission-telegram/releases) page and place `transmission-telegram` in your `$PATH`.

### From Source (Go 1.22+)
```bash
go install github.com/pyed/transmission-telegram/cmd/bot@latest
```

---

## Configuration

You can configure `transmission-telegram` via command-line flags or environment variables:

| Flag | Environment Variable | Description |
|---|---|---|
| `-token` | `TT_BOTT` | Telegram Bot Token (*Required*) |
| `-master` | — | Telegram username allowed to control the bot (*Required*, repeatable for multiple users) |
| `-url` | `TR_URL` | Transmission RPC URL (default: `http://localhost:9091/transmission/rpc`) |
| `-username` | `TR_AUTH` (`user:pass`) | Transmission RPC Username |
| `-password` | `TR_AUTH` (`user:pass`) | Transmission RPC Password |
| `-logfile` | — | Path to log file (default: stdout) |
| `-transmission-logfile` | — | Transmission log file to monitor for torrent completion notifications |
| `-no-live` | — | Disable auto-refreshing message updates for torrent speeds/status |

### Example CLI Usage
```bash
transmission-telegram \
  -token="123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ" \
  -master="@YourTelegramUsername" \
  -url="http://localhost:9091/transmission/rpc" \
  -username="admin" \
  -password="password"
```

---

## Docker

### Standalone
```bash
docker run -d --name transmission-telegram \
  --restart unless-stopped \
  -e TT_BOTT="<Your Bot Token>" \
  -e TR_AUTH="<username>:<password>" \
  kevinhalpin/transmission-telegram:latest \
  -master="<Your Username>" \
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
    image: kevinhalpin/transmission-telegram:latest
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

## Documentation & Commands

See the [Wiki](https://github.com/pyed/transmission-telegram/wiki) for full command documentation and advanced setup.
