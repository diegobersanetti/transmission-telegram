package notify

import (
	"fmt"
	"log"
	"strings"

	"github.com/pyed/tailer"
)

type ChatIDProvider func() int64
type SendFunc func(text string, chatID int64, markdown bool) int

func StartTailer(logFile string, getChatID ChatIDProvider, send SendFunc, logger *log.Logger) {
	go func() {
		ft := tailer.RunFileTailer(logFile, false, nil)

		const (
			substring = `"Incomplete" to "Complete"`
			start     = len(`[2017-02-22 21:00:00.898] `)
			end       = len(` State changed from "Incomplete" to "Complete" (torrent.c:2218)`)
		)

		for {
			select {
			case line := <-ft.Lines():
				if strings.Contains(line, substring) {
					cid := getChatID()
					if cid == 0 {
						continue
					}

					msg := fmt.Sprintf("Completed: %s", line[start:len(line)-end])
					send(msg, cid, false)
				}
			case err := <-ft.Errors():
				logger.Printf("[ERROR] tailing transmission log: %s", err)
				return
			}
		}
	}()
}
