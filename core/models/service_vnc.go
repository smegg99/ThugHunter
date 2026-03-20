// core/models/service_vnc.go
package models

import "time"

// VNCAuthType represents the RFB security type advertised by a VNC server.
type VNCAuthType int

const (
	VNCAuthNone     VNCAuthType = 1 // No authentication
	VNCAuthPassword VNCAuthType = 2 // VNC password authentication
	VNCAuthUnknown  VNCAuthType = -1
)

func (a VNCAuthType) String() string {
	switch a {
	case VNCAuthNone:
		return "None"
	case VNCAuthPassword:
		return "VNCAuth"
	default:
		return "Unknown"
	}
}

// VNCService stores probe results for a VNC endpoint.
type VNCService struct {
	ServiceBase
	RFBVersion      string      `json:"rfb_version"`
	AuthType        VNCAuthType `json:"auth_type"`
	NoAuth          bool        `gorm:"index" json:"no_auth"`
	IsFavorite      bool        `gorm:"default:false" json:"is_favorite"`
	Screenshot      []byte      `gorm:"type:blob" json:"-"`
	ScreenshotAt    *time.Time  `json:"screenshot_at"`
	StaleScreenshot bool        `gorm:"default:false" json:"stale_screenshot"`
}

func (s *VNCService) Type() ServiceType  { return ServiceTypeVNC }
func (s *VNCService) Base() *ServiceBase { return &s.ServiceBase }
