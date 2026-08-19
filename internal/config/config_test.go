package config

import (
	"flag"
	"strings"
	"testing"
	"time"
)

func TestMasterSlice(t *testing.T) {
	var masters MasterSlice

	if err := masters.Set("Sheriff"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := masters.Set("@Pyed"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test Contains (case-insensitivity and @ handling)
	if !masters.Contains("sheriff") {
		t.Errorf("expected Contains('sheriff') to be true")
	}
	if !masters.Contains("SHERIFF") {
		t.Errorf("expected Contains('SHERIFF') to be true")
	}
	if !masters.Contains("@sheriff") {
		t.Errorf("expected Contains('@sheriff') to be true")
	}
	if !masters.Contains("pyed") {
		t.Errorf("expected Contains('pyed') to be true")
	}
	if !masters.Contains("@Pyed") {
		t.Errorf("expected Contains('@Pyed') to be true")
	}
	if masters.Contains("other_user") {
		t.Errorf("expected Contains('other_user') to be false")
	}

	// Test String
	str := masters.String()
	if !strings.Contains(str, "sheriff") || !strings.Contains(str, "@pyed") {
		t.Errorf("unexpected String output: %s", str)
	}
}

func TestParseFlags_Success(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	args := []string{
		"-token=test-token",
		"-master=@User1",
		"-master=user2",
		"-url=http://127.0.0.1:9091/transmission/rpc",
		"-username=admin",
		"-password=secret",
		"-no-live",
	}

	cfg, err := ParseFlags(fs, args, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.BotToken != "test-token" {
		t.Errorf("expected BotToken='test-token', got %q", cfg.BotToken)
	}
	if len(cfg.Masters) != 2 || cfg.Masters[0] != "user1" || cfg.Masters[1] != "user2" {
		t.Errorf("unexpected Masters: %v", cfg.Masters)
	}
	if cfg.RPCURL != "http://127.0.0.1:9091/transmission/rpc" {
		t.Errorf("unexpected RPCURL: %s", cfg.RPCURL)
	}
	if cfg.Username != "admin" || cfg.Password != "secret" {
		t.Errorf("unexpected credentials: %s / %s", cfg.Username, cfg.Password)
	}
	if !cfg.NoLive {
		t.Errorf("expected NoLive=true")
	}
	if cfg.Interval != 5*time.Second {
		t.Errorf("expected Interval=5s, got %v", cfg.Interval)
	}
	if cfg.Duration != 10 {
		t.Errorf("expected Duration=10, got %d", cfg.Duration)
	}
}

func TestParseFlags_EnvVars(t *testing.T) {
	env := map[string]string{
		"TT_BOTT": "env-token-123",
		"TR_AUTH": "user:pass:with:colons",
	}
	getenv := func(key string) string {
		return env[key]
	}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	args := []string{"-master=owner"}

	cfg, err := ParseFlags(fs, args, getenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.BotToken != "env-token-123" {
		t.Errorf("expected BotToken from env, got %q", cfg.BotToken)
	}
	if cfg.Username != "user" || cfg.Password != "pass:with:colons" {
		t.Errorf("expected password with colons preserved, got %q / %q", cfg.Username, cfg.Password)
	}
}

func TestParseFlags_MissingMandatory(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	_, err := ParseFlags(fs, []string{}, nil)
	if err == nil {
		t.Fatalf("expected error when mandatory flags are missing")
	}

	fs2 := flag.NewFlagSet("test", flag.ContinueOnError)
	_, err = ParseFlags(fs2, []string{"-token=tok"}, nil)
	if err == nil {
		t.Fatalf("expected error when master is missing")
	}

	fs3 := flag.NewFlagSet("test", flag.ContinueOnError)
	_, err = ParseFlags(fs3, []string{"-master=user"}, nil)
	if err == nil {
		t.Fatalf("expected error when token is missing")
	}
}
