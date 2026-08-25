package bot

import (
	"reflect"
	"testing"

	"github.com/pyed/transmission"
)

func TestProgressBar(t *testing.T) {
	tests := []struct {
		percent float64
		width   int
		want    string
	}{
		{0.0, 10, "[░░░░░░░░░░]"},
		{0.5, 10, "[█████░░░░░]"},
		{1.0, 10, "[██████████]"},
		{0.23, 10, "[██░░░░░░░░]"},
		{0.87, 10, "[████████░░]"},
		{-0.5, 10, "[░░░░░░░░░░]"}, // clamp lower
		{1.5, 10, "[██████████]"},  // clamp upper
		{0.5, 6, "[███░░░]"},       // custom width
	}

	for _, tt := range tests {
		got := progressBar(tt.percent, tt.width)
		if got != tt.want {
			t.Errorf("progressBar(%f, %d) = %q, want %q", tt.percent, tt.width, got, tt.want)
		}
	}
}

func TestTrackerHosts(t *testing.T) {
	trackers := []transmission.Tracker{
		{Announce: "https://Tracker.Example.com:443/announce?passkey=secret"},
		{Announce: "udp://tracker.example.com:6969/announce"},
		{Announce: "http://second.example.org/announce"},
		{Announce: "udp://[2001:db8::1]:6969/announce"},
		{Announce: "://invalid"},
	}

	want := []string{"2001:db8::1", "second.example.org", "tracker.example.com"}
	if got := trackerHosts(trackers); !reflect.DeepEqual(got, want) {
		t.Fatalf("trackerHosts() = %v, want %v", got, want)
	}
}
