package notify

import (
	"testing"
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
