// core/scanner/store.go
package scanner

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/datastore"
	"smegg.me/thughunter/core/models"
	"smegg.me/thughunter/core/repositories"
	"smegg.me/thughunter/core/screenshot"
)

// screenshotTarget identifies a VNC service that needs a screenshot.
type screenshotTarget struct {
	ID   uint
	IP   string
	Port int
}

// screenshotResult carries a successful capture from a worker to the writer.
type screenshotResult struct {
	ID     uint
	IP     string
	Port   int
	Data   []byte
	Method string
}

// storeResult saves useful data from a HostResult to the database.
func storeResult(_ context.Context, result HostResult, stats *Stats) {
	host := result.Host

	// Update host ping latency.
	if result.Ping.Status == StatusOK {
		host.PingMs = float64(result.Ping.Latency.Milliseconds())
		if err := repositories.GetHostRepository().Upsert(host); err != nil {
			logger.Warn().Err(err).Str("ip", host.IP).Msg("failed to update host ping")
		}
	}

	logger.Debug().Str("ip", host.IP).Int("services", len(result.ServiceResults)).Msg("storing scan results")

	for _, sr := range result.ServiceResults {
		if !sr.Open || sr.Status != StatusOK {
			continue
		}

		switch sr.Service {
		case models.ServiceTypeVNC:
			storeVNCResult(host, sr, stats)
		}
	}
}

// storeVNCResult upserts VNC probe results to the vnc_services table.
// When the probe captured a screenshot (NoAuth server), it is saved too.
func storeVNCResult(host *models.Host, sr ServiceResult, stats *Stats) {
	detail, ok := sr.Detail.(*VNCDetail)
	if !ok || detail == nil {
		return
	}

	svc := &models.VNCService{
		ServiceBase: models.ServiceBase{
			HostID:      host.ID,
			IP:          host.IP,
			Port:        sr.Port,
			ServiceType: models.ServiceTypeVNC,
			LatencyMs:   float64(sr.Latency.Milliseconds()),
		},
		RFBVersion: detail.RFBVersion,
		AuthType:   detail.AuthType,
		NoAuth:     detail.NoAuth,
		Screenshot: detail.Screenshot,
	}
	if len(detail.Screenshot) > 0 {
		now := time.Now()
		svc.ScreenshotAt = &now
	}

	repo := repositories.GetVNCServiceRepository()
	if err := repo.UpdateProbeResult(svc); err != nil {
		logger.Warn().Err(err).Str("ip", host.IP).Int("port", sr.Port).Msg("failed to save VNC service")
		return
	}

	stats.markSaved()
	logger.Info().
		Str("ip", host.IP).
		Int("port", sr.Port).
		Str("auth", detail.AuthType.String()).
		Bool("no_auth", detail.NoAuth).
		Bool("has_screenshot", len(detail.Screenshot) > 0).
		Msg("saved VNC probe result")
}

// fixVNCAuthFromScreenshots updates VNC services that have a screenshot but
// are not marked as no-auth. Having a screenshot proves no password was
// required to connect. This runs once after all scan goroutines finish.
func fixVNCAuthFromScreenshots() {
	db := datastore.Get()
	result := db.Model(&models.VNCService{}).
		Where("length(screenshot) > 0 AND (no_auth = 0 OR auth_type != ?)", models.VNCAuthNone).
		Updates(map[string]any{
			"no_auth":   true,
			"auth_type": models.VNCAuthNone,
		})
	if result.Error != nil {
		logger.Warn().Err(result.Error).Msg("failed to fix VNC auth from screenshots")
		return
	}
	if result.RowsAffected > 0 {
		logger.Info().Int64("fixed", result.RowsAffected).Msg("post-scan: corrected VNC auth for services with screenshots")
	}
}

// captureScreenshots runs a dedicated post-scan phase that captures
// screenshots for VNC services.
func captureScreenshots(ctx context.Context, sc scanConfig, stats *Stats, refreshAll bool) {
	db := datastore.Get()

	var targets []screenshotTarget
	query := db.WithContext(ctx).Table("vnc_services").Select("id, ip, port")
	if refreshAll {
		// All VNC services for a full refresh.
		// no filter
	} else {
		query = query.Where("screenshot IS NULL OR length(screenshot) = 0")
	}
	err := query.Find(&targets).Error
	if err != nil {
		logger.Warn().Err(err).Msg("post-scan: failed to query VNC services for screenshots")
		return
	}
	if len(targets) == 0 {
		logger.Debug().Msg("post-scan: no VNC services need screenshots")
		return
	}

	// Warn if external tool timing budget is impossible.
	extBudgetSec := sc.screenshot.DelaySeconds + sc.screenshot.PauseSeconds
	if sc.screenshot.Template != "" && extBudgetSec >= sc.screenshotTimeout {
		logger.Warn().
			Int("delay_s", sc.screenshot.DelaySeconds).
			Int("pause_s", sc.screenshot.PauseSeconds).
			Int("timeout_s", sc.screenshotTimeout).
			Msg("post-scan: external tool delay+pause >= timeout; external captures will always fail - reduce screenshot_delay_seconds or screenshot_pause_seconds, or increase screenshot_timeout_seconds")
	}

	// When refreshing all, mark existing screenshots as stale.
	// Successfully captured screenshots will clear this flag.
	if refreshAll {
		if res := db.WithContext(ctx).Model(&models.VNCService{}).Where("length(screenshot) > 0").Update("stale_screenshot", true); res.Error != nil {
			logger.Warn().Err(res.Error).Msg("post-scan: failed to mark existing screenshots as stale")
		} else if res.RowsAffected > 0 {
			logger.Info().Int64("count", res.RowsAffected).Msg("post-scan: marked existing screenshots as stale")
		}
	}

	stats.setScreenshotTotal(len(targets))
	logger.Info().Int("count", len(targets)).Int("workers", sc.screenshotMaxWorkers).Msg("post-scan: capturing screenshots for VNC services")

	var captured atomic.Int64

	results := make(chan screenshotResult, sc.screenshotMaxWorkers*2)

	// Opens its own SQLite connection so BLOB writes never compete with
	// the shared MaxOpenConns(1) pool used by frontend queries.
	var writerWg sync.WaitGroup
	writerWg.Add(1)
	go func() {
		defer writerWg.Done()

		wDB, err := datastore.OpenWriter()
		if err != nil {
			logger.Error().Err(err).Msg("post-scan: failed to open writer connection; falling back to shared pool")
			wDB = db // graceful degradation
		} else {
			defer func() {
				if sqlDB, err := wDB.DB(); err == nil {
					sqlDB.Close()
				}
			}()
		}

		for r := range results {
			if err := wDB.Model(&models.VNCService{}).Where("id = ?", r.ID).Updates(map[string]any{
				"screenshot":       r.Data,
				"screenshot_at":    time.Now(),
				"stale_screenshot": false,
			}).Error; err != nil {
				logger.Warn().Err(err).Str("ip", r.IP).Int("port", r.Port).Msg("failed to save post-scan screenshot")
				continue
			}
			captured.Add(1)
			logger.Info().Str("ip", r.IP).Int("port", r.Port).Str("method", r.Method).Msg("post-scan screenshot captured")
		}
	}()

	// Capture worker pool
	work := make(chan screenshotTarget, sc.screenshotMaxWorkers)
	var wg sync.WaitGroup
	for i := 0; i < sc.screenshotMaxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range work {
				// Check for cancellation before starting each capture
				// so Stop() takes effect between targets, not just mid-capture.
				if ctx.Err() != nil {
					stats.markScreenshot()
					continue
				}
				func(t screenshotTarget) {
					defer stats.markScreenshot()

					res := screenshot.Capture(ctx, t.IP, t.Port, sc.screenshotTimeout, sc.screenshot)

					// Track method-level stats.
					switch {
					case res.Method == "native":
						stats.SSNativeOK.Add(1)
					case res.Method == "external":
						stats.SSExtOK.Add(1)
					case errors.Is(res.Err, screenshot.ErrBlank):
						stats.SSBlank.Add(1)
					default:
						// Only count as ext_fail when the external tool was
						// actually attempted (template configured) and the
						// result was not a blank rejection. Otherwise it's a
						// native-only failure.
						if sc.screenshot.Template != "" && res.Method != "native" {
							stats.SSExtFail.Add(1)
						} else {
							stats.SSNativeFail.Add(1)
						}
					}

					if res.Err != nil {
						logger.Warn().Err(res.Err).Str("ip", t.IP).Int("port", t.Port).Msg("post-scan screenshot failed")
						return
					}
					if !sc.screenshot.RejectBlank && !screenshot.Validate(res.Data) {
						stats.SSBlank.Add(1)
						logger.Debug().Str("ip", t.IP).Int("port", t.Port).Msg("post-scan screenshot blank, discarding")
						return
					}

					// Send to writer goroutine (non-blocking on DB).
					select {
					case results <- screenshotResult{
						ID:     t.ID,
						IP:     t.IP,
						Port:   t.Port,
						Data:   res.Data,
						Method: res.Method,
					}:
					case <-ctx.Done():
					}
				}(t)
			}
		}()
	}

	// Enqueue targets; cancel-aware send so Stop() can unblock us.
	for _, t := range targets {
		select {
		case <-ctx.Done():
		case work <- t:
			continue
		}
		break
	}
	close(work)

	// Wait for all capture workers to finish, then close the results
	// channel so the writer drains and exits.
	wg.Wait()
	close(results)
	writerWg.Wait()

	stats.finishScreenshots()
	logger.Info().
		Int64("captured", captured.Load()).
		Int("total", len(targets)).
		Int64("native_ok", stats.SSNativeOK.Load()).
		Int64("native_fail", stats.SSNativeFail.Load()).
		Int64("ext_ok", stats.SSExtOK.Load()).
		Int64("ext_fail", stats.SSExtFail.Load()).
		Int64("blank", stats.SSBlank.Load()).
		Msg("post-scan: screenshot capture complete")
}
