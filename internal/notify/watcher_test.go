package notify

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/pyed/transmission"
)

func TestStartWatcher(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	torrentsList := transmission.Torrents{
		{ID: 1, Name: "Torrent 1", PercentDone: 0.5, Status: transmission.StatusDownloading},
		{ID: 2, Name: "Torrent 2", PercentDone: 1.0, Status: transmission.StatusSeeding},
	}

	getTorrents := func(ctx context.Context) (transmission.Torrents, error) {
		mu.Lock()
		defer mu.Unlock()
		copyList := make(transmission.Torrents, len(torrentsList))
		for i, tor := range torrentsList {
			tCopy := *tor
			copyList[i] = &tCopy
		}
		return copyList, nil
	}

	var sentMessages []string
	var sendMu sync.Mutex

	send := func(text string, chatID int64, markdown bool) int {
		sendMu.Lock()
		defer sendMu.Unlock()
		sentMessages = append(sentMessages, text)
		return 1
	}

	getChatID := func() int64 {
		return 123456789
	}

	// Start watcher with rapid 20ms interval
	StartWatcher(ctx, 20*time.Millisecond, getTorrents, getChatID, send, nil)

	// Allow initial poll
	time.Sleep(50 * time.Millisecond)

	sendMu.Lock()
	if len(sentMessages) != 0 {
		t.Fatalf("expected 0 messages on startup, got %d", len(sentMessages))
	}
	sendMu.Unlock()

	// Transition Torrent 1 to Complete
	mu.Lock()
	torrentsList[0].PercentDone = 1.0
	torrentsList[0].Status = transmission.StatusSeeding
	mu.Unlock()

	// Wait for next poll
	time.Sleep(50 * time.Millisecond)

	sendMu.Lock()
	if len(sentMessages) != 1 {
		t.Fatalf("expected 1 notification message, got %d", len(sentMessages))
	}
	if sentMessages[0] != "Completed: Torrent 1" {
		t.Errorf("expected 'Completed: Torrent 1', got %q", sentMessages[0])
	}
	sendMu.Unlock()

	// Next poll without changes — ensure no duplicate messages
	time.Sleep(50 * time.Millisecond)
	sendMu.Lock()
	if len(sentMessages) != 1 {
		t.Errorf("expected still 1 notification (no duplicates), got %d", len(sentMessages))
	}
	sendMu.Unlock()
}

func TestWatcherReportsOutageAndRecoveryOnce(t *testing.T) {
	state := newWatcherState()
	providerErr := errors.New("connection refused")
	getTorrents := func(context.Context) (transmission.Torrents, error) {
		if providerErr != nil {
			return nil, providerErr
		}
		return transmission.Torrents{}, nil
	}

	var messages []string
	send := func(text string, _ int64, _ bool) int {
		messages = append(messages, text)
		return 1
	}
	chatID := func() int64 { return 123456789 }

	for i := 0; i < outageFailureThreshold+2; i++ {
		state.poll(context.Background(), getTorrents, chatID, send, nil)
	}
	if len(messages) != 1 || messages[0] != outageMessage {
		t.Fatalf("expected one outage notification, got %v", messages)
	}

	providerErr = nil
	state.poll(context.Background(), getTorrents, chatID, send, nil)
	state.poll(context.Background(), getTorrents, chatID, send, nil)
	if len(messages) != 2 || messages[1] != recoveryMessage {
		t.Fatalf("expected one recovery notification, got %v", messages)
	}
}

func TestWatcherSuppressesTransientOutage(t *testing.T) {
	state := newWatcherState()
	failing := true
	getTorrents := func(context.Context) (transmission.Torrents, error) {
		if failing {
			return nil, errors.New("temporary failure")
		}
		return transmission.Torrents{}, nil
	}

	var messages []string
	send := func(text string, _ int64, _ bool) int {
		messages = append(messages, text)
		return 1
	}
	for i := 0; i < outageFailureThreshold-1; i++ {
		state.poll(context.Background(), getTorrents, func() int64 { return 1 }, send, nil)
	}
	failing = false
	state.poll(context.Background(), getTorrents, func() int64 { return 1 }, send, nil)
	if len(messages) != 0 {
		t.Fatalf("transient failures produced notifications: %v", messages)
	}
}
