// core/models/host.go
//
// Host model for discovered network hosts.
package models

import (
	"strings"

	"gorm.io/gorm"
)

// Host represents a discovered network host with location, services, and labels.
type Host struct {
	gorm.Model
	IP          string              `gorm:"uniqueIndex;not null" json:"ip"`
	City        string              `json:"city"`
	Region      string              `json:"region"`
	CountryCode string              `json:"country_code"`
	OS          string              `json:"os"`
	Hardware    string              `json:"hardware"`
	Labels      []string            `gorm:"serializer:json" json:"labels"`
	Services    map[string][]string `gorm:"serializer:json" json:"services"` // Service name to list of ports
	Software    []string            `gorm:"serializer:json" json:"software"`
	PingMs      float64             `json:"ping_ms"`
	IsFavorite  bool                `gorm:"default:false" json:"is_favorite"`
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

// NewHost creates a Host from scraped data, parsing the location string.
func NewHost(ip, locationString, os, hardware string) *Host {
	h := &Host{
		IP:       ip,
		OS:       os,
		Hardware: hardware,
	}

	h.City, h.Region, h.CountryCode = parseLocationString(locationString)

	return h
}
