// services/scanner/services.go
package scanner

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/datastore"
	"smegg.me/thughunter/core/models"
	"smegg.me/thughunter/core/repositories"
	"smegg.me/thughunter/core/screenshot"
)

// vncServiceInfo holds lightweight fields for screenshot capture decisions.
type vncServiceInfo struct {
	ID     uint
	IP     string
	Port   int
	NoAuth bool
	HasSS  bool
}

// vncScreenshotRow holds a service ID and its screenshot blob.
type vncScreenshotRow struct {
	ID         uint
	Screenshot []byte
}

// VNCListItem extends VNCService with host location info for the frontend.
type VNCListItem = repositories.VNCHostListItem

// inFlightCaptures tracks service IDs currently being captured to prevent
// duplicate concurrent captures of the same service.
var inFlightCaptures sync.Map

// VNCPage holds a page of VNC list items with total count.
type VNCPage struct {
	Items []VNCListItem `json:"items"`
	Total int64         `json:"total"`
}

// ListVNCServices returns a paginated list of VNC services enriched with host
// location info. Supports sorting, searching and filtering by country, labels, etc.
func (s *Service) ListVNCServices(page, pageSize int, sortBy, sortOrder, search string, countries, labels []string, hardware, authFilter, screenshotFilter string) (*VNCPage, error) {
	repo := repositories.GetVNCServiceRepository()
	items, total, err := repo.ListFilteredWithHost(repositories.VNCListFilters{
		Page:             page,
		PageSize:         pageSize,
		SortBy:           sortBy,
		SortOrder:        sortOrder,
		Search:           search,
		Countries:        countries,
		Labels:           labels,
		Hardware:         hardware,
		AuthFilter:       authFilter,
		ScreenshotFilter: screenshotFilter,
	})
	if err != nil {
		return nil, err
	}
	return &VNCPage{Items: items, Total: total}, nil
}

// GetVNCScreenshot returns the screenshot for a VNC service as a
// base64-encoded string. Returns empty string if no screenshot exists.
func (s *Service) GetVNCScreenshot(id uint) (string, error) {
	db := datastore.Get()
	var svc models.VNCService
	if err := db.Select("screenshot").First(&svc, id).Error; err != nil {
		return "", fmt.Errorf("get vnc screenshot: %w", err)
	}
	if len(svc.Screenshot) == 0 {
		return "", nil
	}
	return base64.StdEncoding.EncodeToString(svc.Screenshot), nil
}

// ToggleVNCFavorite flips the is_favorite flag on a VNC service and returns the new value.
func (s *Service) ToggleVNCFavorite(id uint) (bool, error) {
	db := datastore.Get()
	var svc models.VNCService
	if err := db.First(&svc, id).Error; err != nil {
		return false, fmt.Errorf("toggle vnc favorite: %w", err)
	}
	svc.IsFavorite = !svc.IsFavorite
	if err := db.Model(&svc).Update("is_favorite", svc.IsFavorite).Error; err != nil {
		return false, fmt.Errorf("toggle vnc favorite: %w", err)
	}
	return svc.IsFavorite, nil
}

// ScreenshotResult holds one id→base64 pair for the frontend.
type ScreenshotResult struct {
	ID         uint   `json:"id"`
	Screenshot string `json:"screenshot"`
}

// RefreshScreenshots captures fresh VNC screenshots on-demand for the
// requested service IDs. For each no-auth service without a cached
// screenshot it connects via go-vnc, grabs the framebuffer, and stores
// the PNG in the DB. Returns the base64-encoded images for all IDs that
// have a screenshot (cached or freshly captured).
func (s *Service) RefreshScreenshots(ids []uint) ([]ScreenshotResult, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	if len(ids) > 200 {
		ids = ids[:200]
	}

	db := datastore.Get()

	var infos []vncServiceInfo
	err := db.Table("vnc_services").
		Select("id, ip, port, no_auth, (CASE WHEN length(screenshot) > 0 THEN 1 ELSE 0 END) as has_ss").
		Where("id IN ?", ids).
		Find(&infos).Error
	if err != nil {
		return nil, fmt.Errorf("query vnc services: %w", err)
	}

	// Determine which services need a fresh capture, skipping any already
	// being captured by a concurrent RefreshScreenshots call.
	var needCapture []vncServiceInfo
	for _, info := range infos {
		if !info.HasSS && info.NoAuth {
			if _, loaded := inFlightCaptures.LoadOrStore(info.ID, struct{}{}); !loaded {
				needCapture = append(needCapture, info)
			}
		}
	}

	// Capture concurrently - CaptureScreenshot already rate-limits via
	// its global semaphore, so no local semaphore needed here.
	var wg sync.WaitGroup

	// Snapshot config outside the goroutines to avoid racing with
	// concurrent config updates via patchConfig.
	cfg := config.Get()
	timeoutSec := int(cfg.Scanner.Workers.ScreenshotTimeoutSeconds)
	if timeoutSec <= 0 {
		timeoutSec = 7
	}
	ssCfg := screenshot.Config{
		Template:     cfg.Scanner.Templates.ScreenshotCommandTemplate,
		DelaySeconds: int(cfg.Scanner.Workers.ScreenshotDelaySeconds),
		PauseSeconds: int(cfg.Scanner.Workers.ScreenshotPauseSeconds),
		RejectBlank:  cfg.Scanner.RejectBlankScreenshots,
	}

	for _, info := range needCapture {
		wg.Add(1)
		go func(si vncServiceInfo) {
			defer wg.Done()
			defer inFlightCaptures.Delete(si.ID)

			// CaptureScreenshot applies the timeout internally after
			// acquiring its semaphore slot, so no outer timeout needed.
			res := screenshot.Capture(context.Background(), si.IP, si.Port, timeoutSec, ssCfg)
			if res.Err != nil {
				logger.Debug().Err(res.Err).Str("ip", si.IP).Int("port", si.Port).Msg("on-demand screenshot failed")
				return
			}

			if err := db.Model(&models.VNCService{}).
				Where("id = ?", si.ID).
				Update("screenshot", res.Data).Error; err != nil {
				logger.Warn().Err(err).Uint("id", si.ID).Msg("failed to save on-demand screenshot")
				return
			}
			logger.Info().Str("ip", si.IP).Int("port", si.Port).Int("bytes", len(res.Data)).Msg("on-demand screenshot captured")
		}(info)
	}
	wg.Wait()

	// Now read all screenshots for the requested IDs.
	var rows []vncScreenshotRow
	if err := db.Table("vnc_services").
		Select("id, screenshot").
		Where("id IN ? AND length(screenshot) > 0", ids).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("read screenshots: %w", err)
	}

	results := make([]ScreenshotResult, 0, len(rows))
	for _, row := range rows {
		results = append(results, ScreenshotResult{
			ID:         row.ID,
			Screenshot: base64.StdEncoding.EncodeToString(row.Screenshot),
		})
	}
	return results, nil
}
