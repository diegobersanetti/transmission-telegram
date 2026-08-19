package bot

import (
	"strings"
	"sync/atomic"
	"testing"

	"github.com/pyed/transmission-telegram/internal/config"
)

func TestBot_ChatID(t *testing.T) {
	b := &Bot{}
	if b.ChatID() != 0 {
		t.Errorf("expected initial ChatID to be 0")
	}

	atomic.StoreInt64(&b.chatID, 123456789)
	if b.ChatID() != 123456789 {
		t.Errorf("expected ChatID to be 123456789, got %d", b.ChatID())
	}
}

func TestBot_MarkdownReplacer(t *testing.T) {
	replacer := strings.NewReplacer(
		"*", "•",
		"[", "(",
		"]", ")",
		"_", "-",
		"`", "'",
	)

	input := "[Torrent_Name] *2026* `1080p`"
	expected := "(Torrent-Name) •2026• '1080p'"
	got := replacer.Replace(input)
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestBot_CommandRegistration(t *testing.T) {
	b := &Bot{
		Config: &config.Config{},
	}
	b.registerCommands()

	expectedCommands := []string{
		"list", "li", "ls",
		"head", "he",
		"tail", "ta",
		"downs", "dg",
		"seeding", "sd",
		"paused", "pa",
		"checking", "ch",
		"active", "ac",
		"errors", "er",
		"sort", "so",
		"trackers", "tr",
		"downloaddir", "dd",
		"add", "ad",
		"search", "se",
		"latest", "la",
		"info", "in",
		"stop", "sp",
		"start", "st",
		"check", "ck",
		"stats", "sa",
		"downlimit", "dl",
		"uplimit", "ul",
		"speed", "ss",
		"count", "co",
		"del", "rm",
		"deldata",
		"help",
		"version", "ver",
	}

	for _, cmd := range expectedCommands {
		if _, ok := b.commands[cmd]; !ok {
			t.Errorf("expected command %q to be registered", cmd)
		}
	}
}
