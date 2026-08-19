package notify

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/pyed/transmission"
)

type TorrentsProvider func(ctx context.Context) (transmission.Torrents, error)

// StartWatcher monitors Transmission torrent completion via periodic RPC polling.
// This enables completion notifications without requiring access to a local log file.
func StartWatcher(ctx context.Context, interval time.Duration, getTorrents TorrentsProvider, getChatID ChatIDProvider, send SendFunc, logger *slog.Logger) {
	if interval <= 0 {
		interval = 10 * time.Second
	}

	go func() {
		knownIncomplete := make(map[int]bool)
		initialized := false
		var mu sync.Mutex

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				torrents, err := getTorrents(ctx)
				if err != nil {
					if logger != nil {
						logger.Warn("Notification watcher failed to get torrents", "error", err)
					}
					continue
				}

				cid := getChatID()
				mu.Lock()

				currentIDs := make(map[int]struct{}, len(torrents))

				for _, t := range torrents {
					currentIDs[t.ID] = struct{}{}
					isComplete := t.PercentDone >= 1.0 || t.Status == transmission.StatusSeeding

					if !initialized {
						if !isComplete {
							knownIncomplete[t.ID] = true
						}
						continue
					}

					// If previously incomplete and now complete
					if knownIncomplete[t.ID] && isComplete {
						delete(knownIncomplete, t.ID)
						if cid != 0 {
							msg := fmt.Sprintf("Completed: %s", t.Name)
							send(msg, cid, false)
						}
					} else if !isComplete {
						knownIncomplete[t.ID] = true
					}
				}

				// Clean up deleted torrents from knownIncomplete map
				for id := range knownIncomplete {
					if _, exists := currentIDs[id]; !exists {
						delete(knownIncomplete, id)
					}
				}

				initialized = true
				mu.Unlock()
			}
		}
	}()
}
