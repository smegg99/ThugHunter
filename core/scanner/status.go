// core/scanner/status.go
//
// Thread-safe scan statistics.
package scanner

import (
	"fmt"
	"sync/atomic"
	"time"
)

// Stats tracks scan progress counters. All methods are safe for concurrent use.
type Stats struct {
	TotalHosts      atomic.Int64
	Scanned         atomic.Int64
	PingOK          atomic.Int64
	PingTimeout     atomic.Int64
	PingError       atomic.Int64
	ProbeOK         atomic.Int64
	ProbeTimeout    atomic.Int64
	ProbeError      atomic.Int64
	Saved           atomic.Int64
	ScreenshotTotal atomic.Int64
	ScreenshotDone  atomic.Int64
	ScreenshotStage atomic.Int32 // see ScreenshotStage constants
	StartedAt       time.Time

	// Screenshot method breakdown.
	SSNativeOK   atomic.Int64
	SSNativeFail atomic.Int64
	SSExtOK      atomic.Int64
	SSExtFail    atomic.Int64
	SSBlank      atomic.Int64
}

func newStats(totalHosts int) *Stats {
	s := &Stats{StartedAt: time.Now()}
	s.TotalHosts.Store(int64(totalHosts))
	return s
}

func (s *Stats) addPing(status Status) {
	switch status {
	case StatusOK:
		s.PingOK.Add(1)
	case StatusTimeout:
		s.PingTimeout.Add(1)
	case StatusError:
		s.PingError.Add(1)
	}
}

func (s *Stats) addProbe(status Status) {
	switch status {
	case StatusOK:
		s.ProbeOK.Add(1)
	case StatusTimeout:
		s.ProbeTimeout.Add(1)
	case StatusError:
		s.ProbeError.Add(1)
	}
}

func (s *Stats) markScanned() { s.Scanned.Add(1) }
func (s *Stats) markSaved()   { s.Saved.Add(1) }

func (s *Stats) setScreenshotTotal(n int) {
	s.ScreenshotTotal.Store(int64(n))
	s.ScreenshotStage.Store(int32(ScreenshotRunning))
}
func (s *Stats) markScreenshot()    { s.ScreenshotDone.Add(1) }
func (s *Stats) finishScreenshots() { s.ScreenshotStage.Store(int32(ScreenshotDone)) }

// Snapshot returns a read-consistent copy of current stats.
type Snapshot struct {
	TotalHosts      int64
	Scanned         int64
	PingOK          int64
	PingTimeout     int64
	PingError       int64
	ProbeOK         int64
	ProbeTimeout    int64
	ProbeError      int64
	Saved           int64
	ScreenshotTotal int64
	ScreenshotDone  int64
	ScreenshotStage ScreenshotStage
	Elapsed         time.Duration
}

func (s *Stats) Snapshot() Snapshot {
	return Snapshot{
		TotalHosts:      s.TotalHosts.Load(),
		Scanned:         s.Scanned.Load(),
		PingOK:          s.PingOK.Load(),
		PingTimeout:     s.PingTimeout.Load(),
		PingError:       s.PingError.Load(),
		ProbeOK:         s.ProbeOK.Load(),
		ProbeTimeout:    s.ProbeTimeout.Load(),
		ProbeError:      s.ProbeError.Load(),
		Saved:           s.Saved.Load(),
		ScreenshotTotal: s.ScreenshotTotal.Load(),
		ScreenshotDone:  s.ScreenshotDone.Load(),
		ScreenshotStage: ScreenshotStage(s.ScreenshotStage.Load()),
		Elapsed:         time.Since(s.StartedAt),
	}
}

func (snap Snapshot) String() string {
	pings := snap.PingOK + snap.PingTimeout + snap.PingError
	var pingRate float64
	if pings > 0 {
		pingRate = float64(snap.PingOK) / float64(pings) * 100
	}
	probes := snap.ProbeOK + snap.ProbeTimeout + snap.ProbeError
	var probeRate float64
	if probes > 0 {
		probeRate = float64(snap.ProbeOK) / float64(probes) * 100
	}
	return fmt.Sprintf(
		"hosts=%d/%d  ping_ok=%d timeout=%d err=%d (%.0f%%)  probe_ok=%d timeout=%d err=%d (%.0f%%)  saved=%d  elapsed=%s",
		snap.Scanned, snap.TotalHosts,
		snap.PingOK, snap.PingTimeout, snap.PingError, pingRate,
		snap.ProbeOK, snap.ProbeTimeout, snap.ProbeError, probeRate,
		snap.Saved,
		snap.Elapsed.Truncate(time.Second),
	)
}
