package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func loadConfig() (map[string]any, string, error) {
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		return loadConfigFromPath(p)
	}

	if format.IsInstalled() {
		return loadInstalledConfig()
	}
	return loadPortableConfig()
}

// loadConfigFromPath loads the config from an explicitly specified path.
func loadConfigFromPath(p string) (map[string]any, string, error) {
	m, err := readJSONFile(p)
	if err != nil {
		return nil, "", fmt.Errorf("load config %s: %w", p, err)
	}
	return m, abs(p), nil
}

func loadPortableConfig() (map[string]any, string, error) {
	m, p, err := findPortableConfig()
	if err == nil {
		return m, abs(p), nil
	}

	p = portableSearchPaths[0]
	ap := abs(p)
	if err := generateDefaultConfig(p); err != nil {
		return nil, "", fmt.Errorf("generate default config: %w", err)
	}

	return loadGeneratedConfig(ap)
}

// findPortableConfig searches for a portable config file in standard locations.
func findPortableConfig() (map[string]any, string, error) {
	for _, p := range portableSearchPaths {
		m, err := readJSONFile(p)
		if err == nil {
			return m, p, nil
		}
	}
	return nil, "", fmt.Errorf("not found")
}

func tryFailOnDefaultConfigGenerateError(path string) error {
	if !failOnDefaultConfigGenerate {
		return nil
	}

	errDefaultGenerated := &ErrDefaultGenerated{Path: path}
	ShowDefaultGeneratedDialog(errDefaultGenerated)
	return errDefaultGenerated
}

// loadGeneratedConfig loads a config file that was just generated.
func loadGeneratedConfig(ap string) (map[string]any, string, error) {
	if err := tryFailOnDefaultConfigGenerateError(ap); err != nil {
		return nil, "", err
	}

	m, err := readJSONFile(portableSearchPaths[0])
	if err != nil {
		return nil, "", fmt.Errorf("read generated config %s: %w", portableSearchPaths[0], err)
	}
	return m, ap, nil
}

// loadInstalledConfig handles installed / read-only formats. It looks for
// user config in the platform config directory, falling back to a bundled
// config that gets copied on first launch.
func loadInstalledConfig() (map[string]any, string, error) {
	cfgDir, err := userConfigDir()
	if err != nil {
		return nil, "", fmt.Errorf("determine user config dir: %w", err)
	}
	userPath := filepath.Join(cfgDir, configFileName)

	// Try user config first
	if m, err := readJSONFile(userPath); err == nil {
		return m, userPath, nil
	}

	// Try bundled config
	if m, err := loadBundledConfig(userPath); err == nil {
		return m, userPath, nil
	}

	// Fall back to generating default
	return loadOrGenerateDefaultConfig(userPath)
}

func loadBundledConfig(userPath string) (map[string]any, error) {
	bundledDir := format.BundledConfigDir()
	if bundledDir == "" {
		return nil, fmt.Errorf("no bundled config dir")
	}

	bundledPath := filepath.Join(bundledDir, configFileName)
	if err := copyFile(bundledPath, userPath); err != nil {
		return nil, err
	}

	return readJSONFile(userPath)
}

func loadOrGenerateDefaultConfig(userPath string) (map[string]any, string, error) {
	if err := generateDefaultConfig(userPath); err != nil {
		return nil, "", fmt.Errorf("generate default config: %w", err)
	}

	if err := tryFailOnDefaultConfigGenerateError(userPath); err != nil {
		return nil, "", err
	}

	m, err := readJSONFile(userPath)
	if err != nil {
		return nil, "", fmt.Errorf("read generated config %s: %w", userPath, err)
	}
	return m, userPath, nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

func readJSONFile(path string) (map[string]any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func abs(p string) string {
	if a, err := filepath.Abs(p); err == nil {
		return a
	}
	return p
}
