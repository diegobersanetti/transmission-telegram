package notify

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestParseCompletionLog(t *testing.T) {
	tests := []struct {
		name         string
		line         string
		expectedName string
		expectedOk   bool
	}{
		{
			name:         "standard transmission log line",
			line:         `[2017-02-22 21:00:00.898] Ubuntu 22.04 LTS State changed from "Incomplete" to "Complete" (torrent.c:2218)`,
			expectedName: "Ubuntu 22.04 LTS",
			expectedOk:   true,
		},
		{
			name:         "different timestamp and line number",
			line:         `[2026-08-19 03:45:00] Debian 12.5 DVD State changed from "Incomplete" to "Complete" (torrent.c:9999)`,
			expectedName: "Debian 12.5 DVD",
			expectedOk:   true,
		},
		{
			name:         "line without timestamp",
			line:         `Arch Linux 2026 State changed from "Incomplete" to "Complete" (torrent.c:100)`,
			expectedName: "Arch Linux 2026",
			expectedOk:   true,
		},
		{
			name:         "unrelated log line",
			line:         `[2026-08-19 03:45:00] Transmission 4.0.5 started (session.c:123)`,
			expectedName: "",
			expectedOk:   false,
		},
		{
			name:         "incomplete state change (e.g. stopped)",
			line:         `[2026-08-19 03:45:00] Ubuntu State changed from "Complete" to "Stopped" (torrent.c:2218)`,
			expectedName: "",
			expectedOk:   false,
		},
		{
			name:         "empty line",
			line:         "",
			expectedName: "",
			expectedOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, ok := ParseCompletionLog(tt.line)
			if ok != tt.expectedOk {
				t.Fatalf("expected ok=%v, got %v", tt.expectedOk, ok)
			}
			if name != tt.expectedName {
				t.Fatalf("expected name=%q, got %q", tt.expectedName, name)
			}
		})
	}
}

func TestStartTailer(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "transmission.log")

	// Create initial log file with some prior data
	if err := os.WriteFile(logFile, []byte("initial line\n"), 0644); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var sentMessages []string
	var mu sync.Mutex

	send := func(text string, chatID int64, markdown bool) int {
		mu.Lock()
		defer mu.Unlock()
		sentMessages = append(sentMessages, text)
		return 1
	}

	getChatID := func() int64 {
		return 123456
	}

	StartTailer(ctx, logFile, getChatID, send, nil)

	// Wait a moment for tailer to seek to EOF
	time.Sleep(50 * time.Millisecond)

	// Append a completion log line
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("failed to open log for append: %v", err)
	}
	f.WriteString(`[2026-08-19 04:00:00] Fedora 40 Workstation State changed from "Incomplete" to "Complete" (torrent.c:1234)` + "\n")
	f.Close()

	// Wait for tailer to read and process line
	time.Sleep(600 * time.Millisecond)

	mu.Lock()
	if len(sentMessages) != 1 {
		t.Fatalf("expected 1 completion message, got %d", len(sentMessages))
	}
	if sentMessages[0] != "Completed: Fedora 40 Workstation" {
		t.Errorf("expected 'Completed: Fedora 40 Workstation', got %q", sentMessages[0])
	}
	mu.Unlock()
}
