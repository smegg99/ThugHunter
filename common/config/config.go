// common/config/config.go
//
//go:generate bash ../../scripts/gen-types.sh
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Main config filename
const configFileName = "config.json"

// Controls whether to error if the default config is generated instead of loaded from a file.
const failOnDefaultConfigGenerate = true

var (
	portableSearchPaths = []string{
		configFileName,
		"config/" + configFileName,
	}

	cfg             Config
	publicCfg       map[string]any
	format          Format
	initOnce        sync.Once
	resolvedCfgPath string
)

// Initialize loads, resolves, validates, and decodes the config file.
// It returns the resolved config path and any error encountered.
func Initialize() (string, error) {
	var initErr error
	var configPath string
	initOnce.Do(func() {
		initializeFormat()

		m, path, err := loadConfig()
		if err != nil {
			initErr = err
			return
		}
		configPath = path
		resolvedCfgPath = path

		initErr = initializeConfigWithPath(path, m)
	})
	return configPath, initErr
}

func initializeFormat() {
	format = DetectFormat()
	loadDotenv()
}

func initializeConfigWithPath(path string, m map[string]any) error {
	if err := initDirRefs(path); err != nil {
		return fmt.Errorf("init directory refs: %w", err)
	}

	private := make(privateTracker)
	if err := resolveMap(m, "", private); err != nil {
		return fmt.Errorf("resolve refs: %w", err)
	}

	return validateAndDecode(path, m, private)
}

// Get returns the loaded config struct.
func Get() *Config { return &cfg }

// GetFormat returns the detected build format.
func GetFormat() Format { return format }

// GetConfigPath returns the resolved path to the config file.
func GetConfigPath() string { return resolvedCfgPath }

// GetConfigDir returns the resolved config directory.
func GetConfigDir() string { return resolvedCfgDir }

// GetDataDir returns the resolved data directory.
func GetDataDir() string { return resolvedDataDir }

// PublicJSON returns the public (non-private) config as JSON bytes.
func PublicJSON() ([]byte, error) {
	return json.Marshal(publicCfg)
}

// ExportPublicJSON writes the public config to a JSON file.
func ExportPublicJSON(path string) error {
	b, err := json.MarshalIndent(publicCfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal public config: %w", err)
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

// LaunchedViaAppImage reports whether the process is running inside an AppImage.
func LaunchedViaAppImage() bool {
	if os.Getenv("APPIMAGE") != "" || os.Getenv("APPDIR") != "" {
		return true
	}
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	return strings.Contains(filepath.Clean(exe), "/tmp/.mount_")
}

func initDirRefs(configPath string) error {
	if format.IsInstalled() {
		cfgDir, err := userConfigDir()
		if err != nil {
			return fmt.Errorf("config dir: %w", err)
		}
		resolvedCfgDir = cfgDir

		dataDir, err := userDataDir()
		if err != nil {
			return fmt.Errorf("data dir: %w", err)
		}
		resolvedDataDir = dataDir
	} else {
		abs, err := filepath.Abs(configPath)
		if err != nil {
			return fmt.Errorf("abs config path: %w", err)
		}
		dir := filepath.Dir(abs)
		resolvedCfgDir = dir
		resolvedDataDir = dir
	}
	return nil
}
