// common/models/models.go
package models

import (
	"database/sql/driver"
	"encoding/json"
)

type JSONStringSlice []string

func (j *JSONStringSlice) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), j)
}

func (j JSONStringSlice) Value() (driver.Value, error) {
	return json.Marshal(j)
}

type JSONServiceMap map[string]int

func (m *JSONServiceMap) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), m)
}

func (m JSONServiceMap) Value() (driver.Value, error) {
	return json.Marshal(m)
}

type Host struct {
	IP       string `gorm:"primaryKey"`
	Hostname string
	Labels   JSONStringSlice `gorm:"type:json"`
	Location string
	Services JSONServiceMap `gorm:"type:json"`
}
