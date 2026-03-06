//go:build !linux

package theme

import "time"

const pollingInterval = 5 * time.Second

func watchAccentChanges(onChange func()) (stop func()) {
	done := make(chan struct{})
	go func() {
		prev := Snapshot()
		ticker := time.NewTicker(pollingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				cur := Snapshot()
				if cur.AccentColor != prev.AccentColor || cur.DarkMode != prev.DarkMode {
					prev = cur
					onChange()
				}
			case <-done:
				return
			}
		}
	}()
	return func() { close(done) }
}
