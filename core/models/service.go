// core/models/scraper_service.go
package models

import "gorm.io/gorm"

type ServiceType string

const (
	ServiceTypeVNC      ServiceType = "VNC"
	ServiceTypeRDP      ServiceType = "RDP"
	ServiceTypeSPICE    ServiceType = "SPICE"
	ServiceTypePKCamera ServiceType = "PKCamera"
)

func (s ServiceType) String() string {
	return string(s)
}

type Service interface {
	Type() ServiceType
	Base() *ServiceBase
}

type ServiceBase struct {
	gorm.Model
	HostID      uint        `gorm:"index;not null" json:"host_id"`
	IP          string      `json:"ip"`
	Port        int         `json:"port"`
	ServiceType ServiceType `json:"service_type"`
}
