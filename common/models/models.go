// common/models/models.go
package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSONStringSlice []string

func (j *JSONStringSlice) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return fmt.Errorf("cannot scan type %T into JSONStringSlice", value)
	}
}

func (j JSONStringSlice) Value() (driver.Value, error) {
	return json.Marshal(j)
}

type JSONServiceMap map[string]int

func (m *JSONServiceMap) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, m)
	case string:
		return json.Unmarshal([]byte(v), m)
	default:
		return fmt.Errorf("cannot scan type %T into JSONServiceMap", value)
	}
}

func (m JSONServiceMap) Value() (driver.Value, error) {
	return json.Marshal(m)
}

type Host struct {
	IP       string `gorm:"primaryKey"`
	Hostname string
	Labels   JSONStringSlice `gorm:"type:text"`
	Location string
	Services JSONServiceMap `gorm:"type:text"`
}
