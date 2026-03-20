// services/scanner/service.go
package scanner

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"smegg.me/thughunter/common/logger"
	corescanner "smegg.me/thughunter/core/scanner"
)

// ScanMode identifies which kind of scan is active.
type ScanMode string

const (
	ScanModeHosts       ScanMode = "hosts"       // host ping + service probe
	ScanModeScreenshots ScanMode = "screenshots" // screenshot-only pass
)

// ScanProgress is a point-in-time snapshot of a running or completed scan.
type ScanProgress struct {
	Running         bool                        `json:"running"`
	Mode            ScanMode                    `json:"mode"`
	TotalHosts      int64                       `json:"total_hosts"`
	Scanned         int64                       `json:"scanned"`
	PingOK          int64                       `json:"ping_ok"`
	ProbeOK         int64                       `json:"probe_ok"`
	Saved           int64                       `json:"saved"`
	ElapsedSecs     float64                     `json:"elapsed_secs"`
	ScreenshotStage corescanner.ScreenshotStage `json:"screenshot_stage"`
	ScreenshotTotal int64                       `json:"screenshot_total"`
	ScreenshotDone  int64                       `json:"screenshot_done"`
}

// Service is the Wails-bound scanner service.
type Service struct {
	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
	mode   ScanMode
}

// Start begins a full scan of all known hosts in a background goroutine.
// Returns an error if the scanner is already running.
func (s *Service) Start() error {
	return s.startRun(ScanModeHosts, func(ctx context.Context, sc *corescanner.Scanner) error {
		return sc.Run(ctx)
	})
}

// StartHostScan begins a host-only scan (ping + probe, no screenshots).
func (s *Service) StartHostScan() error {
	return s.startRun(ScanModeHosts, func(ctx context.Context, sc *corescanner.Scanner) error {
		return sc.RunHostsOnly(ctx)
	})
}

// StartScreenshots begins a screenshot-only run for no-auth VNC services.
func (s *Service) StartScreenshots() error {
	return s.startRun(ScanModeScreenshots, func(ctx context.Context, sc *corescanner.Scanner) error {
		return sc.RunScreenshotsOnly(ctx)
	})
}

// startRun is the shared entry point for all scan modes.
func (s *Service) startRun(mode ScanMode, runFn func(context.Context, *corescanner.Scanner) error) error {
	sc := corescanner.Get()
	if sc == nil {
		return fmt.Errorf("scanner not initialized")
	}

	s.mu.Lock()
	if s.cancel != nil {
		s.mu.Unlock()
		return fmt.Errorf("scanner is already running")
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	s.cancel = cancel
	s.done = done
	s.mode = mode
	s.mu.Unlock()

	emitEvent(EventScanStarted, string(mode))

	go s.runScan(ctx, cancel, done, sc, runFn)
	return nil
}

// runScan is the background goroutine that drives the scan lifecycle.
func (s *Service) runScan(ctx context.Context, cancel context.CancelFunc, done chan struct{}, sc *corescanner.Scanner, runFn func(context.Context, *corescanner.Scanner) error) {
	defer func() {
		emitEvent(EventScanCompleted, buildProgress(sc, s.currentMode()))
		s.mu.Lock()
		s.cancel = nil
		s.done = nil
		s.mode = ""
		s.mu.Unlock()
		cancel()
		close(done)
	}()

	stopPoll := startProgressPoller(ctx, sc, s)
	defer stopPoll()

	if err := runFn(ctx, sc); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error().Err(err).Msg("scanner: scan run failed")
		emitEvent(EventScanError, err.Error())
	}
}

// currentMode returns the current scan mode (thread-safe).
func (s *Service) currentMode() ScanMode {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mode
}

// Stop cancels the active scan and waits for it to finish.
// Gives in-flight captures up to 10 seconds to drain after cancellation.
func (s *Service) Stop() {
	s.mu.Lock()
	cancel := s.cancel
	done := s.done
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if done != nil {
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			logger.Warn().Msg("scanner: stop timed out waiting for captures to drain")
		}
	}
}

// IsRunning reports whether a scan is currently active.
func (s *Service) IsRunning() bool {
	sc := corescanner.Get()
	if sc == nil {
		return false
	}
	return sc.Running()
}

// GetProgress returns a point-in-time snapshot of the current scan state.
func (s *Service) GetProgress() *ScanProgress {
	return buildProgress(corescanner.Get(), s.currentMode())
}

// buildProgress translates a scanner.Snapshot into a ScanProgress response.
// Returns nil when idle so the frontend can hide the progress indicator.
func buildProgress(sc *corescanner.Scanner, mode ScanMode) *ScanProgress {
	if sc == nil {
		return nil
	}
	snap := sc.Progress()
	if snap == nil && !sc.Running() {
		return nil
	}
	p := &ScanProgress{Running: sc.Running(), Mode: mode}
	if snap == nil {
		return p
	}
	p.TotalHosts = snap.TotalHosts
	p.Scanned = snap.Scanned
	p.PingOK = snap.PingOK
	p.ProbeOK = snap.ProbeOK
	p.Saved = snap.Saved
	p.ElapsedSecs = snap.Elapsed.Seconds()
	p.ScreenshotStage = snap.ScreenshotStage
	p.ScreenshotTotal = snap.ScreenshotTotal
	p.ScreenshotDone = snap.ScreenshotDone
	return p
}

// startProgressPoller emits scan progress events at a fixed interval while
// the context is live.  Returns a stop function that waits for the goroutine
// to fully exit before returning, preventing concurrent event emission
// between the poller and scan-completion code.
func startProgressPoller(ctx context.Context, sc *corescanner.Scanner, svc *Service) func() {
	done := make(chan struct{})
	stop := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				emitEvent(EventScanProgress, buildProgress(sc, svc.currentMode()))
			}
		}
	}()
	return func() {
		close(stop)
		<-done
	}
}
