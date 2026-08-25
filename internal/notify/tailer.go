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

// StartTailer monitors a Transmission log file for completion events using
// standard library file operations. It starts at EOF, but follows later file
// replacement or truncation so normal log rotation does not stop alerts.
func StartTailer(ctx context.Context, logFile string, getChatID ChatIDProvider, send SendFunc, logger *slog.Logger) error {
	file, err := openLogFile(logFile)
	if err != nil {
		return fmt.Errorf("open transmission logfile %q: %w", logFile, err)
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("stat transmission logfile %q: %w", logFile, err)
	}
	offset, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		file.Close()
		return fmt.Errorf("seek transmission logfile %q: %w", logFile, err)
	}

	go func() {
		defer func() { _ = file.Close() }()
		reader := bufio.NewReader(file)
		var pending string

		for {
			line, readErr := reader.ReadString('\n')
			if len(line) > 0 {
				offset += int64(len(line))
				pending += line
			}
			if strings.HasSuffix(pending, "\n") {
				if torrentName, ok := ParseCompletionLog(strings.TrimRight(pending, "\r\n")); ok {
					if cid := getChatID(); cid != 0 {
						send(fmt.Sprintf("Completed: %s", torrentName), cid, false)
					}
				}
				pending = ""
			}

			if readErr == nil {
				continue
			}
			if !errors.Is(readErr, io.EOF) {
				if logger != nil {
					logger.Error("Error reading transmission logfile", "error", readErr)
				}
				return
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(500 * time.Millisecond):
			}

			currentInfo, statErr := os.Stat(logFile)
			if statErr != nil {
				continue
			}
			if os.SameFile(info, currentInfo) && currentInfo.Size() >= offset {
				continue
			}

			newFile, openErr := openLogFile(logFile)
			if openErr != nil {
				continue
			}
			if closeErr := file.Close(); closeErr != nil && logger != nil {
				logger.Warn("Failed to close rotated transmission logfile", "error", closeErr)
			}
			file = newFile
			info = currentInfo
			offset = 0
			pending = ""
			reader = bufio.NewReader(file)
			if logger != nil {
				logger.Info("Reopened rotated transmission logfile", "path", logFile)
			}
		}
	}()

	return nil
}
