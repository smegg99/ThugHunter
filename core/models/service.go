// core/models/service.go
//
// Shared service types and base struct for all protocol-specific services.
package models

import "gorm.io/gorm"

// ServiceType identifies the protocol of a discovered service.
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

// Service is implemented by all protocol-specific service models.
type Service interface {
	Type() ServiceType
	Base() *ServiceBase
}

// ServiceBase holds fields common to all discovered services.
type ServiceBase struct {
	gorm.Model
	HostID      uint        `gorm:"index;not null" json:"host_id"`
	IP          string      `json:"ip"`
	Port        int         `json:"port"`
	ServiceType ServiceType `json:"service_type"`
	LatencyMs   float64     `json:"latency_ms"`
}
