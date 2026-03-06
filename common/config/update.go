// common/config/update.go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"cuelang.org/go/cue"
)

var mu sync.Mutex

// SetConfig replaces the full config on disk and in memory.
func SetConfig(c Config) error {
	return patchFromStruct(c)
}

// SetPreferences updates only the preferences section.
func SetPreferences(p Preferences) error {
	return patchFromStruct(struct {
		Preferences Preferences `json:"preferences"`
	}{Preferences: p})
}

// SetLoggerConfig updates only the logger section.
func SetLoggerConfig(lc LoggerConfig) error {
	return patchFromStruct(struct {
		Logger LoggerConfig `json:"logger"`
	}{Logger: lc})
}

// Marshals a typed struct into a map and applies it as a patch.
func patchFromStruct(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal patch: %w", err)
	}
	var patch map[string]any
	if err := json.Unmarshal(b, &patch); err != nil {
		return fmt.Errorf("unmarshal patch: %w", err)
	}
	return patchConfig(patch)
}

// Applies a partial update to the config. The patch is deep-merged
// into the on-disk JSON (preserving unmodified refs), validated with the CUE
// schema, and the in-memory config is refreshed.
func patchConfig(patch map[string]any) error {
	mu.Lock()
	defer mu.Unlock()

	raw, err := readAndMergePatch(patch)
	if err != nil {
		return err
	}

	unified, err := validateResolvedConfig(resolvedCfgPath, deepCopyMap(raw))
	if err != nil {
		return err
	}

	if err := writeMergedConfig(raw); err != nil {
		return err
	}

	return updateInMemoryConfigFromUnified(unified)
}

// Reads the current config and merges the patch into it.
func readAndMergePatch(patch map[string]any) (map[string]any, error) {
	raw, err := readJSONFile(resolvedCfgPath)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	deepMerge(raw, patch)
	return raw, nil
}

// Writes the merged config to disk in pretty JSON format.
func writeMergedConfig(raw map[string]any) error {
	pretty, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(resolvedCfgPath, append(pretty, '\n'), 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// Updates cfg and publicCfg from the validated CUE output.
func updateInMemoryConfigFromUnified(unified cue.Value) error {
	b, err := unified.MarshalJSON()
	if err != nil {
		return fmt.Errorf("export config: %w", err)
	}

	var fullMap map[string]any
	if err := json.Unmarshal(b, &fullMap); err != nil {
		return fmt.Errorf("unmarshal unified: %w", err)
	}

	private := make(privateTracker)
	if err := resolveMap(fullMap, "", private); err != nil {
		return fmt.Errorf("resolve defaults: %w", err)
	}

	publicCfg = buildPublicMap(fullMap, "", private)

	b, err = json.Marshal(fullMap)
	if err != nil {
		return fmt.Errorf("re-marshal: %w", err)
	}

	if err := json.Unmarshal(b, &cfg); err != nil {
		return fmt.Errorf("decode config: %w", err)
	}

	return nil
}

// Recursively merges src into dst. Nested maps are descended;
// all other types in src overwrite the corresponding key in dst.
func deepMerge(dst, src map[string]any) {
	for k, sv := range src {
		dv, exists := dst[k]
		if !exists {
			dst[k] = sv
			continue
		}
		dstMap, dstOk := dv.(map[string]any)
		srcMap, srcOk := sv.(map[string]any)
		if dstOk && srcOk {
			deepMerge(dstMap, srcMap)
		} else {
			dst[k] = sv
		}
	}
}

// Returns a deep copy of m via JSON round-trip.
func deepCopyMap(m map[string]any) map[string]any {
	b, _ := json.Marshal(m)
	var out map[string]any
	_ = json.Unmarshal(b, &out)
	return out
}
