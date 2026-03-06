package config

import (
	"path/filepath"

	"github.com/joho/godotenv"
)

const envFileName = ".env"

// Loads env file based on how the app was installed/launched.
func loadDotenv() {
	if format.IsInstalled() {
		if cfgDir, err := userConfigDir(); err == nil {
			_ = godotenv.Load(filepath.Join(cfgDir, envFileName))
		}
		if bundledDir := format.BundledConfigDir(); bundledDir != "" {
			_ = godotenv.Load(filepath.Join(bundledDir, envFileName))
		}
		return
	}
	_ = godotenv.Load()
}
