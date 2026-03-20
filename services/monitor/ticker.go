// services/monitor/ticker.go
package monitor

import (
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

var (
	tickerMu   sync.Mutex
	tickerStop chan struct{}
)

// startPolling begins a background goroutine that emits system stats events
// at the given interval. Only one poller runs at a time.
func startPolling(intervalMs int64) {
	tickerMu.Lock()
	defer tickerMu.Unlock()

	if tickerStop != nil {
		close(tickerStop)
	}

	stop := make(chan struct{})
	tickerStop = stop

	go poll(time.Duration(intervalMs)*time.Millisecond, stop)
}

// stopPolling halts the background polling goroutine if running.
func stopPolling() {
	tickerMu.Lock()
	defer tickerMu.Unlock()

	if tickerStop != nil {
		close(tickerStop)
		tickerStop = nil
	}
}

func poll(interval time.Duration, stop <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			emitSnapshot(interval)
		}
	}
}

func emitSnapshot(interval time.Duration) {
	app := application.Get()
	if app == nil {
		return
	}

	now := time.Now().UnixMilli()

	cpuStats, cpuErr := collectCPU()
	ramStats, ramErr := collectRAM()
	swapStats, swapErr := collectSwap()

	if cpuErr == nil {
		app.Event.Emit(EventMonitorCPU, cpuStats)
	}
	if ramErr == nil {
		app.Event.Emit(EventMonitorRAM, ramStats)
	}
	if swapErr == nil {
		app.Event.Emit(EventMonitorSwap, swapStats)
	}

	if cpuErr == nil && ramErr == nil && swapErr == nil {
		app.Event.Emit(EventMonitorSystem, SystemSnapshot{
			CPU:             cpuStats,
			RAM:             ramStats,
			Swap:            swapStats,
			IntervalMs:      interval.Milliseconds(),
			SampledAtUnixMs: now,
		})
	}
}
