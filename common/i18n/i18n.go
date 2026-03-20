// common/i18n/i18n.go
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed locales/*.json
var localeFS embed.FS

// locale is the currently active locale code (e.g. "en", "pl").
var (
	mu     sync.RWMutex
	locale = "en"
	msgs   map[string]string // flat key -> translated string
)

func init() {
	loadLocale("en")
}

// SetLocale switches the active locale and loads its messages.
// Unknown locales silently fall back to "en".
func SetLocale(code string) {
	mu.Lock()
	defer mu.Unlock()

	code = normalizeCode(code)
	if code == locale {
		return
	}
	loadLocale(code)
}

// GetLocale returns the currently active locale code.
func GetLocale() string {
	mu.RLock()
	defer mu.RUnlock()
	return locale
}

// T returns the translated string for key, or the key itself if missing.
func T(key string) string {
	mu.RLock()
	defer mu.RUnlock()

	if s, ok := msgs[key]; ok {
		return s
	}
	return key
}

// Tf returns the translated string for key with fmt.Sprintf arguments.
func Tf(key string, args ...any) string {
	mu.RLock()
	defer mu.RUnlock()

	if s, ok := msgs[key]; ok {
		return fmt.Sprintf(s, args...)
	}
	return fmt.Sprintf(key, args...)
}

// normalizeCode strips region suffixes (e.g. "en-US" into "en").
func normalizeCode(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	if i := strings.IndexAny(code, "-_"); i > 0 {
		code = code[:i]
	}
	if code == "" {
		return "en"
	}
	return code
}

// loadLocale reads and parses a locale JSON file into msgs.
// On any failure it falls back to "en". Must be called with mu held.
func loadLocale(code string) {
	m, err := readLocaleFile(code)
	if err != nil && code != "en" {
		m, err = readLocaleFile("en")
		code = "en"
	}
	if err != nil {
		msgs = map[string]string{}
		locale = "en"
		return
	}
	locale = code
	msgs = flatten(m, "")
}

func readLocaleFile(code string) (map[string]any, error) {
	data, err := localeFS.ReadFile("locales/" + code + ".json")
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// flatten recursively converts a nested map to dot-separated keys.
func flatten(m map[string]any, prefix string) map[string]string {
	out := make(map[string]string)
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case string:
			out[key] = val
		case map[string]any:
			for fk, fv := range flatten(val, key) {
				out[fk] = fv
			}
		}
	}
	return out
}
