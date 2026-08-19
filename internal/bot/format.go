package bot

import "strings"

// progressBar generates a visual progress bar of the given width (e.g. [█████░░░░░]).
func progressBar(percent float64, width int) string {
	if width <= 0 {
		width = 10
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 1.0 {
		percent = 1.0
	}

	filled := int(percent * float64(width))
	if filled > width {
		filled = width
	}

	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}
