// common/config/defaults.go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"

	"smegg.me/thughunter/schema"
)

func generateDefaultConfig(path string) error {
	result, err := extractDefaultsFromSchema()
	if err != nil {
		return err
	}

	if err := ensureConfigDirectoryExists(path); err != nil {
		return err
	}

	return writeDefaultConfigFile(path, result)
}

func extractDefaultsFromSchema() (map[string]any, error) {
	ctx := cuecontext.New()

	cueVal := ctx.CompileString(schema.CUE)
	if cueVal.Err() != nil {
		return nil, fmt.Errorf("compile schema: %w", cueVal.Err())
	}

	def := cueVal.LookupPath(cue.ParsePath("#Config"))
	if def.Err() != nil {
		return nil, fmt.Errorf("lookup #Config: %w", def.Err())
	}

	m, err := extractDefaults(def)
	if err != nil {
		return nil, fmt.Errorf("extract defaults: %w", err)
	}

	result, ok := m.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected object at top level")
	}

	return result, nil
}

func ensureConfigDirectoryExists(path string) error {
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create config dir: %w", err)
		}
	}
	return nil
}

func writeDefaultConfigFile(path string, data map[string]any) error {
	pretty, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("format defaults: %w", err)
	}

	if err := os.WriteFile(path, append(pretty, '\n'), 0o644); err != nil {
		return fmt.Errorf("write default config: %w", err)
	}

	return nil
}

func extractDefaults(v cue.Value) (any, error) {
	// Apply default value if it exists
	if d, ok := v.Default(); ok {
		v = d
	}

	switch v.IncompleteKind() {
	case cue.StructKind:
		return extractStructDefaults(v)
	case cue.ListKind:
		return extractListDefaults(v)
	case cue.StringKind:
		return extractStringDefault(v)
	case cue.IntKind:
		return extractIntDefault(v)
	case cue.FloatKind:
		return extractFloatDefault(v)
	case cue.BoolKind:
		return extractBoolDefault(v)
	case cue.NumberKind:
		return extractNumberDefault(v)
	default:
		return nil, nil
	}
}

func extractStructDefaults(v cue.Value) (any, error) {
	m := make(map[string]any)
	iter, err := v.Fields(cue.Optional(false))
	if err != nil {
		return nil, err
	}
	for iter.Next() {
		child, err := extractDefaults(iter.Value())
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", iter.Selector(), err)
		}
		m[iter.Selector().String()] = child
	}
	return m, nil
}

func extractListDefaults(v cue.Value) (any, error) {
	var list []any
	iter, err := v.List()
	if err != nil {
		return nil, err
	}
	for iter.Next() {
		elem, err := extractDefaults(iter.Value())
		if err != nil {
			return nil, err
		}
		list = append(list, elem)
	}
	if list == nil {
		list = []any{}
	}
	return list, nil
}

func extractStringDefault(v cue.Value) (any, error) {
	if s, err := v.String(); err == nil {
		return s, nil
	}
	return "", nil
}

func extractIntDefault(v cue.Value) (any, error) {
	if i, err := v.Int64(); err == nil {
		return i, nil
	}
	return int64(0), nil
}

func extractFloatDefault(v cue.Value) (any, error) {
	if f, err := v.Float64(); err == nil {
		return f, nil
	}
	return float64(0), nil
}

func extractBoolDefault(v cue.Value) (any, error) {
	if b, err := v.Bool(); err == nil {
		return b, nil
	}
	return false, nil
}

func extractNumberDefault(v cue.Value) (any, error) {
	if i, err := v.Int64(); err == nil {
		return i, nil
	}
	if f, err := v.Float64(); err == nil {
		return f, nil
	}
	return int64(0), nil
}
