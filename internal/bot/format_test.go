package bot

import "testing"

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
