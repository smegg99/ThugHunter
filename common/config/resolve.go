// common/config/resolve.go
package config

import (
	"fmt"
	"strings"
)

type privateTracker map[string]bool

func joinPath(parts ...string) string {
	var nonEmpty []string
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	return strings.Join(nonEmpty, ".")
}

// Recursively resolves all config references in the given map, tracking private fields.
func resolveMap(m map[string]any, prefix string, private privateTracker) error {
	for k, v := range m {
		path := joinPath(prefix, k)

		switch val := v.(type) {
		case string:
			if err := resolveStringField(m, k, path, val, private); err != nil {
				return err
			}

		case map[string]any:
			if err := resolveMap(val, path, private); err != nil {
				return err
			}

		case []any:
			if err := resolveSlice(k, val, path, private); err != nil {
				return err
			}
		}
	}
	return nil
}

// Attempts to resolve a string value as a config reference.
func resolveStringField(m map[string]any, key, path, val string, private privateTracker) error {
	r, ok := parseRef(val)
	if !ok {
		return nil
	}

	if r.isPrivate() {
		private[path] = true
	}

	resolved, err := r.resolve()
	if err != nil {
		return fmt.Errorf("field %q: %w", key, err)
	}
	m[key] = resolved
	return nil
}

// Attempts to resolve all slice elements as config references.
func resolveSlice(key string, slice []any, path string, private privateTracker) error {
	for i, elem := range slice {
		if err := resolveSliceElement(slice, i, key, path, elem, private); err != nil {
			return err
		}
	}
	return nil
}

// Attempts to resolve a single slice element as a config reference.
func resolveSliceElement(slice []any, index int, key, path string, elem any, private privateTracker) error {
	s, ok := elem.(string)
	if !ok {
		return nil
	}

	r, ok := parseRef(s)
	if !ok {
		return nil
	}

	if r.isPrivate() {
		private[path] = true
	}

	resolved, err := r.resolve()
	if err != nil {
		return fmt.Errorf("field %q[%d]: %w", key, index, err)
	}
	slice[index] = resolved
	return nil
}

// Recursively builds a public-safe map, filtering out private fields.
func buildPublicMap(m map[string]any, prefix string, private privateTracker) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		path := joinPath(prefix, k)
		if private[path] {
			continue
		}
		out[k] = buildPublicValue(v, path, private)
	}
	return out
}

// Recursively builds public-safe values, filtering out private fields.
func buildPublicValue(v any, path string, private privateTracker) any {
	switch val := v.(type) {
	case map[string]any:
		return buildPublicMap(val, path, private)
	case []any:
		cp := make([]any, len(val))
		copy(cp, val)
		return cp
	default:
		return v
	}
}
