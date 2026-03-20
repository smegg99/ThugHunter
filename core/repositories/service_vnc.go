// core/repositories/service_vnc.go
package repositories

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"smegg.me/thughunter/core/datastore"
	"smegg.me/thughunter/core/models"
)

// VNCServiceRepository extends ServiceRepository with VNC-specific queries.
type VNCServiceRepository struct {
	*ServiceRepository[models.VNCService]
}

// GetVNCServiceRepository returns a VNCServiceRepository backed by the global DB.
func GetVNCServiceRepository() *VNCServiceRepository {
	return &VNCServiceRepository{ServiceRepository: NewServiceRepository[models.VNCService]()}
}

// UpdateProbeResult upserts the VNC service probe data (auth type, RFB version,
// latency) for an existing ip+port record. If the record doesn't exist yet it
// creates it. This is called by the scanner after a successful VNC handshake.
func (r *VNCServiceRepository) UpdateProbeResult(svc *models.VNCService) error {
	db := datastore.Get()
	var existing models.VNCService
	err := db.Where("ip = ? AND port = ?", svc.IP, svc.Port).First(&existing).Error
	if err == nil {
		updates := map[string]any{
			"host_id":     svc.HostID,
			"rfb_version": svc.RFBVersion,
			"auth_type":   svc.AuthType,
			"no_auth":     svc.NoAuth,
			"latency_ms":  svc.LatencyMs,
		}
		if len(svc.Screenshot) > 0 {
			updates["screenshot"] = svc.Screenshot
			updates["screenshot_at"] = svc.ScreenshotAt
			updates["stale_screenshot"] = false
		}
		return db.Model(&existing).Updates(updates).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("update vnc probe ip=%q port=%d: %w", svc.IP, svc.Port, err)
	}
	return db.Create(svc).Error
}

// VNCListFilters bundles all filter/sort/pagination parameters for VNC listing.
type VNCListFilters struct {
	Page             int
	PageSize         int
	SortBy           string
	SortOrder        string
	Search           string
	Countries        []string
	Labels           []string
	Hardware         string
	AuthFilter       string
	ScreenshotFilter string
}

// VNCHostListItem holds a VNC service enriched with host location info.
type VNCHostListItem struct {
	models.VNCService
	CountryCode   string `json:"country_code"`
	City          string `json:"city"`
	OS            string `json:"os"`
	Hardware      string `json:"hardware"`
	Labels        string `json:"labels"`
	HasScreenshot bool   `json:"has_screenshot"`
	ScreenshotAt  string `json:"screenshot_at_str"`
}

// allowedVNCSortCols maps user-facing sort keys to qualified column names.
var allowedVNCSortCols = map[string]string{
	"ip":         "vnc_services.ip",
	"port":       "vnc_services.port",
	"no_auth":    "vnc_services.no_auth",
	"latency_ms": "vnc_services.latency_ms",
}

// ListFilteredWithHost returns a paginated list of VNC services joined with
// host data, applying all filters, search and sorting from f.
func (r *VNCServiceRepository) ListFilteredWithHost(f VNCListFilters) ([]VNCHostListItem, int64, error) {
	db := datastore.Get()
	_, pageSize, offset := normalizePage(f.Page, f.PageSize)

	base := db.Table("vnc_services").
		Joins("LEFT JOIN hosts ON hosts.id = vnc_services.host_id")

	if s := strings.TrimSpace(f.Search); s != "" {
		base = base.Where("vnc_services.ip LIKE ?", "%"+s+"%")
	}
	if len(f.Countries) > 0 {
		base = base.Where("hosts.country_code IN ?", f.Countries)
	}
	for _, lbl := range f.Labels {
		base = base.Where("hosts.labels LIKE ?", "%"+lbl+"%")
	}
	if hw := strings.TrimSpace(f.Hardware); hw != "" {
		base = base.Where("hosts.hardware LIKE ?", "%"+hw+"%")
	}
	switch f.AuthFilter {
	case "open":
		base = base.Where("vnc_services.no_auth = 1")
	case "closed":
		base = base.Where("vnc_services.no_auth = 0")
	}
	switch f.ScreenshotFilter {
	case "has":
		base = base.Where("length(vnc_services.screenshot) > 0")
	case "none":
		base = base.Where("vnc_services.screenshot IS NULL OR length(vnc_services.screenshot) = 0")
	}

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count vnc services: %w", err)
	}

	col := "vnc_services.ip"
	if mapped, ok := allowedVNCSortCols[f.SortBy]; ok {
		col = mapped
	}
	ord := "asc"
	if strings.ToLower(f.SortOrder) == "desc" {
		ord = "desc"
	}

	favFirst := `CASE WHEN COALESCE(vnc_services.is_favorite, 0) = 1 THEN 0 ELSE 1 END`
	unprobedLast := `CASE WHEN vnc_services.latency_ms = 0 THEN 1 ELSE 0 END`
	orderExpr := fmt.Sprintf(`%s, %s, %s %s, vnc_services.latency_ms ASC`, favFirst, unprobedLast, col, ord)

	var items []VNCHostListItem
	err := base.
		Select(`vnc_services.id, vnc_services.created_at, vnc_services.updated_at,
			vnc_services.host_id, vnc_services.ip, vnc_services.port,
			vnc_services.service_type, vnc_services.latency_ms,
			vnc_services.rfb_version, vnc_services.auth_type, vnc_services.no_auth,
			COALESCE(vnc_services.is_favorite, 0) as is_favorite,
			CASE WHEN length(vnc_services.screenshot) > 0 THEN 1 ELSE 0 END as has_screenshot,
			vnc_services.screenshot_at,
			COALESCE(vnc_services.stale_screenshot, 0) as stale_screenshot,
			hosts.country_code, hosts.city, hosts.os, hosts.hardware, hosts.labels`).
		Order(orderExpr).
		Offset(offset).
		Limit(pageSize).
		Find(&items).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list vnc services: %w", err)
	}
	return items, total, nil
}
