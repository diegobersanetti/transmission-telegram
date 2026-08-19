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

func TestNormalizeCommand(t *testing.T) {
	tests := []struct {
		input    string
		wantCmd  string
		wantLink bool
	}{
		{"/list", "list", false},
		{"/list@MyTransmissionBot", "list", false},
		{"/LiSt@my_bot", "list", false},
		{"list", "list", false},
		{"/info@TransmissionBot", "info", false},
		{"/help", "help", false},
		{"magnet:?xt=urn:btih:xyz", "add", true},
		{"http://example.com/test.torrent", "add", true},
		{"https://example.com/test.torrent", "add", true},
		{"/add", "add", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cmd, isLink := normalizeCommand(tt.input)
			if cmd != tt.wantCmd || isLink != tt.wantLink {
				t.Errorf("normalizeCommand(%q) = (%q, %v), want (%q, %v)",
					tt.input, cmd, isLink, tt.wantCmd, tt.wantLink)
			}
		})
	}
}

func TestDefaultBotCommands(t *testing.T) {
	b := &Bot{
		Config: &config.Config{},
	}
	b.registerCommands()

	cmds := defaultBotCommands()
	if len(cmds) == 0 {
		t.Fatalf("expected non-empty defaultBotCommands")
	}

	for _, c := range cmds {
		if strings.HasPrefix(c.Command, "/") {
			t.Errorf("command %q should not have leading slash", c.Command)
		}
		if c.Description == "" {
			t.Errorf("command %q has empty description", c.Command)
		}
		if _, ok := b.commands[c.Command]; !ok {
			t.Errorf("menu command %q is not registered in bot handlers", c.Command)
		}
	}
}
