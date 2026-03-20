// core/scanner/scanner.go
package scanner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/models"
	"smegg.me/thughunter/core/screenshot"
)

// Scanner is the top-level coordinator.
type Scanner struct {
	mu      sync.Mutex
	running bool
	stats   *Stats
}

var (
	instance *Scanner
	initOnce sync.Once
)

// Initialize creates the global Scanner singleton.
func Initialize() {
	initOnce.Do(func() {
		instance = &Scanner{}
		logger.Info().Msg("scanner initialized")
	})
}

// Get returns the global scanner instance.
func Get() *Scanner { return instance }

// Running reports whether a scan is in progress.
func (s *Scanner) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// Progress returns the current stats snapshot, or nil if not running.
func (s *Scanner) Progress() *Snapshot {
	s.mu.Lock()
	st := s.stats
	s.mu.Unlock()
	if st == nil {
		return nil
	}
	snap := st.Snapshot()
	return &snap
}

// scanConfig holds resolved config values for a single run.
type scanConfig struct {
	pingMode             PingMode
	icmpPing             bool
	pingTimeout          int
	connectTimeout       int
	bannerTimeout        int
	screenshotTimeout    int
	maxWorkers           int
	screenshotMaxWorkers int
	screenshot           screenshot.Config
}

// resolveScanConfig reads the config singleton and returns a scanConfig.
func resolveScanConfig() scanConfig {
	cfg := config.Get()
	sc := scanConfig{
		pingMode:          PingMode(cfg.Scanner.PingMode),
		icmpPing:          cfg.Scanner.IcmpPing,
		pingTimeout:       int(cfg.Scanner.Workers.PingTimeoutSeconds),
		connectTimeout:    int(cfg.Scanner.Workers.ConnectTimeoutSeconds),
		bannerTimeout:     int(cfg.Scanner.Workers.BannerTimeoutSeconds),
		screenshotTimeout: int(cfg.Scanner.Workers.ScreenshotTimeoutSeconds),
		maxWorkers:        int(cfg.Scanner.Workers.MaxWorkers),
	}
	if sc.pingMode == "" {
		sc.pingMode = PingModeSoft
	}
	sc.screenshotMaxWorkers = int(cfg.Scanner.Workers.ScreenshotMaxWorkers)
	if sc.screenshotTimeout <= 0 {
		sc.screenshotTimeout = 15
	}
	if sc.screenshotMaxWorkers <= 0 {
		sc.screenshotMaxWorkers = 32
	}

	sc.screenshot = screenshot.Config{
		Template:     cfg.Scanner.Templates.ScreenshotCommandTemplate,
		DelaySeconds: int(cfg.Scanner.Workers.ScreenshotDelaySeconds),
		PauseSeconds: int(cfg.Scanner.Workers.ScreenshotPauseSeconds),
		RejectBlank:  cfg.Scanner.RejectBlankScreenshots,
	}
	if sc.screenshot.DelaySeconds < 0 {
		sc.screenshot.DelaySeconds = 1
	}
	if sc.screenshot.PauseSeconds < 0 {
		sc.screenshot.PauseSeconds = 3
	}

	return sc
}

// ResetStats clears the stats from any previous run so Progress() returns nil.
func (s *Scanner) ResetStats() {
	s.mu.Lock()
	s.stats = nil
	s.mu.Unlock()
}

// Run starts a full scan: load hosts, fan out to workers, collect results,
// then capture screenshots.
// Blocks until all hosts are processed or ctx is cancelled.
func (s *Scanner) Run(ctx context.Context) error {
	if err := s.acquireRun(); err != nil {
		return err
	}
	defer s.releaseRun()

	sc, hosts, err := s.prepareHostScan()
	if err != nil {
		return err
	}

	s.logRunStart(hosts, sc)
	logger.Debug().
		Int("ping_timeout", sc.pingTimeout).
		Int("connect_timeout", sc.connectTimeout).
		Int("banner_timeout", sc.bannerTimeout).
		Int("screenshot_timeout", sc.screenshotTimeout).
		Int("screenshot_workers", sc.screenshotMaxWorkers).
		Msg("resolved scan config")

	stats := s.initStats(len(hosts))

	if err := s.runWorkerPool(ctx, hosts, sc, stats); err != nil {
		return err
	}

	// Post-scan: capture screenshots for no-auth VNC services, then fix
	// auth for any services where a screenshot proves no password was needed.
	logger.Debug().Int("screenshot_workers", sc.screenshotMaxWorkers).Msg("starting post-scan screenshot phase")
	screenshot.InitSem(sc.screenshotMaxWorkers)
	captureScreenshots(ctx, sc, stats, false)
	fixVNCAuthFromScreenshots()

	return nil
}

// RunHostsOnly starts a host scan (ping + probe) without the screenshot phase.
func (s *Scanner) RunHostsOnly(ctx context.Context) error {
	if err := s.acquireRun(); err != nil {
		return err
	}
	defer s.releaseRun()

	sc, hosts, err := s.prepareHostScan()
	if err != nil {
		return err
	}

	s.logRunStart(hosts, sc)
	stats := s.initStats(len(hosts))

	return s.runWorkerPool(ctx, hosts, sc, stats)
}

// RunScreenshotsOnly captures screenshots for no-auth VNC services without
// running a host scan first.
func (s *Scanner) RunScreenshotsOnly(ctx context.Context) error {
	if err := s.acquireRun(); err != nil {
		return err
	}
	defer s.releaseRun()

	sc := resolveScanConfig()
	stats := s.initStats(0)

	logger.Info().Int("workers", sc.screenshotMaxWorkers).Msg("starting screenshot-only run")
	screenshot.InitSem(sc.screenshotMaxWorkers)
	captureScreenshots(ctx, sc, stats, true)
	fixVNCAuthFromScreenshots()

	return nil
}

// prepareHostScan resets stats, resolves config, loads hosts, and validates
// that there is work to do. Returns the config and host list ready for scanning.
func (s *Scanner) prepareHostScan() (scanConfig, []*models.Host, error) {
	s.ResetStats()

	sc := resolveScanConfig()

	hosts, err := loadHosts()
	if err != nil {
		return sc, nil, err
	}
	if len(hosts) == 0 {
		return sc, nil, fmt.Errorf("no hosts to scan")
	}

	if sc.maxWorkers > len(hosts) {
		sc.maxWorkers = len(hosts)
	}

	return sc, hosts, nil
}

// initStats creates a new Stats instance and stores it on the scanner.
func (s *Scanner) initStats(total int) *Stats {
	stats := newStats(total)
	s.mu.Lock()
	s.stats = stats
	s.mu.Unlock()
	return stats
}

// runWorkerPool fans out host scanning to workers, collects results, and
// drains the pipeline. Blocks until all hosts are processed or ctx is cancelled.
func (s *Scanner) runWorkerPool(ctx context.Context, hosts []*models.Host, sc scanConfig, stats *Stats) error {
	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()

	queue := s.fillQueue(hosts)
	results := make(chan HostResult, sc.maxWorkers)

	var wg sync.WaitGroup
	for i := 0; i < sc.maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runWorker(workerCtx, queue, results, enabledProbes, sc, stats)
		}()
	}
	logger.Debug().Int("workers", sc.maxWorkers).Msg("scan workers spawned")

	go func() {
		wg.Wait()
		close(results)
	}()

	err := s.collectResults(ctx, results, stats)
	workerCancel()
	// Drain remaining results so the wg.Wait goroutine can finish.
	for range results {
	}

	return err
}

// acquireRun claims exclusive ownership of the scan lifecycle.
func (s *Scanner) acquireRun() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return fmt.Errorf("scanner is already running")
	}
	s.running = true
	return nil
}

// releaseRun marks the scan as finished.
func (s *Scanner) releaseRun() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
}

// logRunStart emits an info log with the run parameters.
func (s *Scanner) logRunStart(hosts []*models.Host, sc scanConfig) {
	logger.Info().
		Int("hosts", len(hosts)).
		Int("workers", sc.maxWorkers).
		Str("ping_mode", string(sc.pingMode)).
		Bool("icmp_ping", sc.icmpPing).
		Msg("starting scan")
}

// fillQueue pushes all hosts into a buffered channel and closes it.
func (s *Scanner) fillQueue(hosts []*models.Host) <-chan *models.Host {
	queue := make(chan *models.Host, len(hosts))
	for _, h := range hosts {
		queue <- h
	}
	close(queue)
	return queue
}

// collectResults drains the results channel, persists data, and logs progress.
func (s *Scanner) collectResults(ctx context.Context, results <-chan HostResult, stats *Stats) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case result, ok := <-results:
			if !ok {
				snap := stats.Snapshot()
				logger.Info().Str("summary", snap.String()).Msg("scan finished")
				return nil
			}
			stats.markScanned()
			storeResult(ctx, result, stats)

		case <-ticker.C:
			snap := stats.Snapshot()
			logger.Info().Str("progress", snap.String()).Msg("scan progress")

		case <-ctx.Done():
			snap := stats.Snapshot()
			logger.Warn().Str("summary", snap.String()).Msg("scan cancelled")
			return ctx.Err()
		}
	}
}
