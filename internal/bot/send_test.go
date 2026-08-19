package bot

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSplitMessage_ShortMessage(t *testing.T) {
	msg := "Hello world"
	chunks := splitMessage(msg, 4096)
	if len(chunks) != 1 || chunks[0] != msg {
		t.Fatalf("expected single chunk %q, got %v", msg, chunks)
	}
}

func TestSplitMessage_EmptyMessage(t *testing.T) {
	chunks := splitMessage("", 4096)
	if len(chunks) != 0 {
		t.Fatalf("expected 0 chunks for empty message, got %v", chunks)
	}
}

func TestSplitMessage_NewlineSplit(t *testing.T) {
	// 3 lines of 2000 chars each = 6000 chars total
	line1 := strings.Repeat("a", 2000)
	line2 := strings.Repeat("b", 2000)
	line3 := strings.Repeat("c", 2000)
	text := line1 + "\n" + line2 + "\n" + line3

	chunks := splitMessage(text, 4096)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}

	if chunks[0] != line1+"\n"+line2 {
		t.Errorf("chunk 0 mismatch")
	}
	if chunks[1] != line3 {
		t.Errorf("chunk 1 mismatch")
	}
}

func TestSplitMessage_NoNewlineFallback(t *testing.T) {
	// 5000 characters without any newlines
	text := strings.Repeat("x", 5000)
	chunks := splitMessage(text, 4096)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	if len(chunks[0]) != 4096 {
		t.Errorf("expected chunk 0 len 4096, got %d", len(chunks[0]))
	}
	if len(chunks[1]) != 904 {
		t.Errorf("expected chunk 1 len 904, got %d", len(chunks[1]))
	}
}

func TestSplitMessage_UnicodeMultiByte(t *testing.T) {
	// Unicode runes: "🚀" is 4 bytes each
	rocket := "🚀"
	text := strings.Repeat(rocket, 5000) // 5000 runes

	chunks := splitMessage(text, 4096)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}

	if utf8.RuneCountInString(chunks[0]) != 4096 {
		t.Errorf("expected 4096 runes in chunk 0, got %d", utf8.RuneCountInString(chunks[0]))
	}
	if utf8.RuneCountInString(chunks[1]) != 904 {
		t.Errorf("expected 904 runes in chunk 1, got %d", utf8.RuneCountInString(chunks[1]))
	}

	// Verify valid UTF-8
	if !utf8.ValidString(chunks[0]) || !utf8.ValidString(chunks[1]) {
		t.Errorf("chunks must be valid UTF-8 strings")
	}
}
