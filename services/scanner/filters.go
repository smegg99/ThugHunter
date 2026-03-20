// services/scanner/filters.go
package scanner

import (
	"fmt"

	"smegg.me/thughunter/core/datastore"
	"smegg.me/thughunter/core/models"
)

// FilterOptions contains the available filter values for the browse UI.
type FilterOptions struct {
	Countries []string `json:"countries"`
	Labels    []string `json:"labels"`
}

// GetFilterOptions returns distinct countries and labels from all hosts.
func (s *Service) GetFilterOptions() (*FilterOptions, error) {
	db := datastore.Get()

	var countries []string
	err := db.Model(&models.Host{}).
		Where("country_code != ''").
		Distinct("country_code").
		Order("country_code ASC").
		Pluck("country_code", &countries).Error
	if err != nil {
		return nil, fmt.Errorf("get distinct countries: %w", err)
	}

	// Labels are stored as JSON arrays; extract unique values.
	var rawLabels []string
	err = db.Model(&models.Host{}).
		Where("labels != '' AND labels != '[]' AND labels != 'null'").
		Pluck("labels", &rawLabels).Error
	if err != nil {
		return nil, fmt.Errorf("get labels: %w", err)
	}

	seen := make(map[string]struct{})
	var labels []string
	for _, raw := range rawLabels {
		// Labels are serialized as JSON arrays, e.g. ["web","linux"].
		// Simple parsing: strip brackets, split by comma, trim quotes.
		parsed := parseJSONStringArray(raw)
		for _, lbl := range parsed {
			if lbl != "" {
				if _, ok := seen[lbl]; !ok {
					seen[lbl] = struct{}{}
					labels = append(labels, lbl)
				}
			}
		}
	}

	return &FilterOptions{
		Countries: countries,
		Labels:    labels,
	}, nil
}

// parseJSONStringArray does a lightweight parse of a JSON string array.
func parseJSONStringArray(s string) []string {
	// Strip leading/trailing whitespace and brackets.
	s = trimByte(s, '[')
	s = trimByte(s, ']')
	if s == "" {
		return nil
	}
	var result []string
	for _, part := range splitUnquoted(s, ',') {
		part = trimByte(part, '"')
		part = trimByte(part, ' ')
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func trimByte(s string, b byte) string {
	for len(s) > 0 && s[0] == b {
		s = s[1:]
	}
	for len(s) > 0 && s[len(s)-1] == b {
		s = s[:len(s)-1]
	}
	return s
}

func splitUnquoted(s string, sep byte) []string {
	var result []string
	inQuote := false
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '"' {
			inQuote = !inQuote
		} else if s[i] == sep && !inQuote {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}
