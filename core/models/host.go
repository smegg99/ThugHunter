// core/models/crawler_host.go
package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type Host struct {
	ID               uint              `gorm:"primaryKey" json:"id"`
	IP               string            `gorm:"uniqueIndex;not null" json:"ip"`
	City             string            `json:"city"`
	Region           string            `json:"region"`
	CountryCode      string            `json:"country_code"`
	OS               string            `json:"os"`
	Hardware         string            `json:"hardware"`
	VNCServices      []VNCService      `gorm:"foreignKey:HostID" json:"vnc_services"`
	RDPServices      []RDPService      `gorm:"foreignKey:HostID" json:"rdp_services"`
	SPICEServices    []SPICEService    `gorm:"foreignKey:HostID" json:"spice_services"`
	PKCameraServices []PKCameraService `gorm:"foreignKey:HostID" json:"pkcamera_services"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	DeletedAt        gorm.DeletedAt    `gorm:"index" json:"deleted_at,omitempty"`
}

func (h *Host) AddService(service Service) {
	switch s := service.(type) {
	case *VNCService:
		h.VNCServices = append(h.VNCServices, *s)
	case *RDPService:
		h.RDPServices = append(h.RDPServices, *s)
	case *SPICEService:
		h.SPICEServices = append(h.SPICEServices, *s)
	case *PKCameraService:
		h.PKCameraServices = append(h.PKCameraServices, *s)
	}
}

func (h *Host) GetAllServices() []any {
	var services []any
	for _, s := range h.VNCServices {
		services = append(services, s)
	}
	for _, s := range h.RDPServices {
		services = append(services, s)
	}
	for _, s := range h.SPICEServices {
		services = append(services, s)
	}
	for _, s := range h.PKCameraServices {
		services = append(services, s)
	}
	return services
}

func parseLocationString(locationString string) (city, region, countryCode string) {
	parts := strings.Split(locationString, ",")
	if len(parts) != 2 {
		return "", "", ""
	}

	city = strings.TrimSpace(parts[0])

	subParts := strings.Split(parts[1], "(")
	if len(subParts) != 2 {
		return "", "", ""
	}

	region = strings.TrimSpace(subParts[0])
	countryCode = strings.TrimSuffix(strings.TrimSpace(subParts[1]), ")")

	return city, region, countryCode
}

func NewHost(ip, locationString, os, hardware string) *Host {
	h := &Host{
		IP:       ip,
		OS:       os,
		Hardware: hardware,
	}

	h.City, h.Region, h.CountryCode = parseLocationString(locationString)

	return h
}
