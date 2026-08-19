package notify

import (
	"fmt"
	"log"
	"strings"

	"github.com/pyed/tailer"
)

type ChatIDProvider func() int64
type SendFunc func(text string, chatID int64, markdown bool) int

// ParseCompletionLog extracts the torrent name from a Transmission completion log line.
// Example format: [2017-02-22 21:00:00.898] Torrent Name State changed from "Incomplete" to "Complete" (torrent.c:2218)
func ParseCompletionLog(line string) (string, bool) {
	const marker = `"Incomplete" to "Complete"`
	if !strings.Contains(line, marker) {
		return "", false
	}

	stateIdx := strings.Index(line, " State changed from")
	if stateIdx == -1 {
		return "", false
	}

	// Find the end of the timestamp prefix "[...]"
	timeEndIdx := strings.Index(line, "] ")
	startIdx := 0
	if timeEndIdx != -1 && timeEndIdx+2 < stateIdx {
		startIdx = timeEndIdx + 2
	}

	name := strings.TrimSpace(line[startIdx:stateIdx])
	if name == "" {
		return "", false
	}

	return name, true
}

func StartTailer(logFile string, getChatID ChatIDProvider, send SendFunc, logger *log.Logger) {
	go func() {
		ft := tailer.RunFileTailer(logFile, false, nil)

		for {
			select {
			case line := <-ft.Lines():
				torrentName, ok := ParseCompletionLog(line)
				if ok {
					cid := getChatID()
					if cid == 0 {
						continue
					}

					msg := fmt.Sprintf("Completed: %s", torrentName)
					send(msg, cid, false)
				}
			case err := <-ft.Errors():
				logger.Printf("[ERROR] tailing transmission log: %s", err)
				return
			}
		}
	}()
}
