// services/config/config.go
package config

import (
	"os/exec"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"

	appconfig "smegg.me/thughunter/common/config"
	"smegg.me/thughunter/core/catcher"
)

const (
	EventConfigChanged      = "config:changed"
	EventPreferencesChanged = "config:preferences:changed"
)

type Service struct{}

// GetConfig returns the full resolved application config.
func (s *Service) GetConfig() appconfig.Config {
	return *appconfig.Get()
}

// SetConfig replaces the entire config on disk and in memory.
func (s *Service) SetConfig(c appconfig.Config) error {
	if err := appconfig.SetConfig(c); err != nil {
		return err
	}
	emitConfigEvents()
	return nil
}

// GetPreferences returns the current user preferences.
func (s *Service) GetPreferences() appconfig.Preferences {
	return appconfig.Get().Preferences
}

// SetPreferences persists updated preferences to the config file and
// updates the in-memory config.
func (s *Service) SetPreferences(p appconfig.Preferences) error {
	if err := appconfig.SetPreferences(p); err != nil {
		return err
	}
	emitConfigEvents()
	return nil
}

// GetLoggerConfig returns the current logger configuration.
func (s *Service) GetLoggerConfig() appconfig.LoggerConfig {
	return appconfig.Get().Logger
}

// SetLoggerConfig persists updated logger configuration.
func (s *Service) SetLoggerConfig(lc appconfig.LoggerConfig) error {
	if err := appconfig.SetLoggerConfig(lc); err != nil {
		return err
	}
	emitConfigEvents()
	return nil
}

// GetImapConfig returns the current IMAP configuration.
func (s *Service) GetImapConfig() appconfig.Config {
	c := *appconfig.Get()
	// Return a config with only IMAP populated (the binding exposes the full struct).
	return c
}

// SetImapConfig persists updated IMAP configuration.
func (s *Service) SetImapConfig(c appconfig.Config) error {
	if err := appconfig.SetImapConfig(c); err != nil {
		return err
	}
	emitConfigEvents()
	return nil
}

// GetScraperConfig returns the current scraper configuration.
func (s *Service) GetScraperConfig() appconfig.Config {
	return *appconfig.Get()
}

// SetScraperConfig persists updated scraper configuration.
func (s *Service) SetScraperConfig(c appconfig.Config) error {
	if err := appconfig.SetScraperConfig(c); err != nil {
		return err
	}
	emitConfigEvents()
	return nil
}

// GetScannerConfig returns the current scanner configuration.
func (s *Service) GetScannerConfig() appconfig.Config {
	return *appconfig.Get()
}

// SetScannerConfig persists updated scanner configuration.
func (s *Service) SetScannerConfig(c appconfig.Config) error {
	if err := appconfig.SetScannerConfig(c); err != nil {
		return err
	}
	emitConfigEvents()
	return nil
}

// IsVirtualDisplayAvailable reports whether a virtual display (Xvfb) can be used.
// This is only possible on Linux when the xvfb-run or Xvfb binary is found on PATH.
func (s *Service) IsVirtualDisplayAvailable() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	if _, err := exec.LookPath("Xvfb"); err == nil {
		return true
	}
	if _, err := exec.LookPath("xvfb-run"); err == nil {
		return true
	}
	return false
}

// VerifyImapConnection tests the full IMAP flow: connect, authenticate, and select mailbox.
// Errors are returned as "code: detail" strings where code is a localisation key.
func (s *Service) VerifyImapConnection() error {
	cfg := appconfig.Get()

	if cfg.Imap.Host == "" {
		return &catcher.ValidationError{Code: "host_not_configured"}
	}
	if cfg.Imap.Port <= 0 || cfg.Imap.Port > 65535 {
		return &catcher.ValidationError{Code: "port_invalid"}
	}
	if cfg.Imap.CatchAllUsername == "" || cfg.Imap.CatchAllPassword == "" {
		return &catcher.ValidationError{Code: "credentials_not_configured"}
	}
	if cfg.Imap.Mbox == "" {
		return &catcher.ValidationError{Code: "mailbox_not_configured"}
	}

	return catcher.ValidateCredentials(
		cfg.Imap.Host,
		int(cfg.Imap.Port),
		cfg.Imap.CatchAllUsername,
		cfg.Imap.CatchAllPassword,
		cfg.Imap.Mbox,
		cfg.Imap.UseTls,
	)
}

func emitConfigEvents() {
	if app := application.Get(); app != nil {
		updated := appconfig.Get()
		app.Event.Emit(EventPreferencesChanged, updated.Preferences)
		app.Event.Emit(EventConfigChanged, *updated)
	}
}
