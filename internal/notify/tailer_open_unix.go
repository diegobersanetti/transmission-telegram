//go:build !windows

package notify

import "os"

func openLogFile(name string) (*os.File, error) {
	return os.Open(name)
}
