package notify

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
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

// StartTailer monitors a Transmission log file for completion events using standard library file operations.
func StartTailer(ctx context.Context, logFile string, getChatID ChatIDProvider, send SendFunc, logger *slog.Logger) {
	go func() {
		file, err := os.Open(logFile)
		if err != nil {
			if logger != nil {
				logger.Error("Failed to open transmission logfile", "path", logFile, "error", err)
			}
			return
		}
		defer file.Close()

		// Seek to end of file so we only tail newly appended logs
		if _, err := file.Seek(0, io.SeekEnd); err != nil {
			if logger != nil {
				logger.Error("Failed to seek transmission logfile", "error", err)
			}
			return
		}

		reader := bufio.NewReader(file)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				line, err := reader.ReadString('\n')
				if err != nil {
					if errors.Is(err, io.EOF) {
						time.Sleep(500 * time.Millisecond)
						continue
					}
					if logger != nil {
						logger.Error("Error reading transmission logfile", "error", err)
					}
					return
				}

				torrentName, ok := ParseCompletionLog(line)
				if ok {
					cid := getChatID()
					if cid != 0 {
						msg := fmt.Sprintf("Completed: %s", torrentName)
						send(msg, cid, false)
					}
				}
			}
		}
	}()
}
