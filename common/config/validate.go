// common/config/validate.go
package config

import (
	"encoding/json"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueerrors "cuelang.org/go/cue/errors"
	cuejson "cuelang.org/go/encoding/json"

	"smegg.me/thughunter/schema"
)

func validateAndDecode(path string, m map[string]any, private privateTracker) error {
	unified, err := validateResolvedConfig(path, m)
	if err != nil {
		return err
	}

	if err := updatePublicAndInMemoryConfig(unified, private); err != nil {
		return err
	}

	return nil
}

// Marshals the config map and validates it with the CUE schema.
func validateResolvedConfig(path string, m map[string]any) (cue.Value, error) {
	resolved, err := json.Marshal(m)
	if err != nil {
		return cue.Value{}, fmt.Errorf("marshal resolved config: %w", err)
	}

	return validateCUE(path, resolved)
}

// Rebuilds the public config and updates the in-memory cfg struct.
func updatePublicAndInMemoryConfig(unified cue.Value, private privateTracker) error {
	b, err := unified.MarshalJSON()
	if err != nil {
		return fmt.Errorf("export config: %w", err)
	}

	// Unmarshal the CUE output (with defaults filled in) back to a map
	// so we can resolve any @{...} refs that came from CUE defaults.
	var fullMap map[string]any
	if err := json.Unmarshal(b, &fullMap); err != nil {
		return fmt.Errorf("unmarshal unified config: %w", err)
	}

	if err := resolveMap(fullMap, "", private); err != nil {
		return fmt.Errorf("resolve default refs: %w", err)
	}

	// Rebuild public config with fully resolved values (including CUE defaults).
	publicCfg = buildPublicMap(fullMap, "", private)

	b, err = json.Marshal(fullMap)
	if err != nil {
		return fmt.Errorf("re-marshal config: %w", err)
	}

	if err := json.Unmarshal(b, &cfg); err != nil {
		return fmt.Errorf("decode config: %w", err)
	}
	return nil
}

func validateCUE(path string, data []byte) (cue.Value, error) {
	cueVal, err := compileCUESchema()
	if err != nil {
		return cue.Value{}, err
	}

	def := getCUEConfigDefinition(cueVal)
	if def.Err() != nil {
		return cue.Value{}, fmt.Errorf("lookup #Config: %w", def.Err())
	}

	return validateUnifiedConfig(path, data, def)
}

// Compiles the CUE schema.
func compileCUESchema() (cue.Value, error) {
	ctx := cuecontext.New()

	cueVal := ctx.CompileString(schema.CUE)
	if cueVal.Err() != nil {
		return cue.Value{}, fmt.Errorf("compile schema: %w", cueVal.Err())
	}
	return cueVal, nil
}

// Retrieves the #Config definition from the compiled CUE schema.
func getCUEConfigDefinition(cueVal cue.Value) cue.Value {
	return cueVal.LookupPath(cue.ParsePath("#Config"))
}

// Unifies the data with the schema definition and validates it.
func validateUnifiedConfig(path string, data []byte, def cue.Value) (cue.Value, error) {
	ctx := cuecontext.New()

	dataExpr, err := cuejson.Extract(path, data)
	if err != nil {
		return cue.Value{}, fmt.Errorf("parse json: %w", err)
	}

	unified := def.Unify(ctx.BuildExpr(dataExpr))
	if err := unified.Validate(cue.Concrete(true)); err != nil {
		details := strings.TrimSpace(cueerrors.Details(err, nil))
		return cue.Value{}, fmt.Errorf("invalid config:\n%s", details)
	}
	return unified, nil
}
