package bot

import (
	"net/url"
	"sort"
	"strings"

	"github.com/pyed/transmission"
)

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

func trackerHosts(trackers []transmission.Tracker) []string {
	unique := make(map[string]struct{}, len(trackers))
	for _, tracker := range trackers {
		u, err := url.Parse(strings.TrimSpace(tracker.Announce))
		if err != nil {
			continue
		}
		host := strings.ToLower(u.Hostname())
		if host != "" {
			unique[host] = struct{}{}
		}
	}

	hosts := make([]string, 0, len(unique))
	for host := range unique {
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)
	return hosts
}
