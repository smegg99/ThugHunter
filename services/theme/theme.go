package theme

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	apptheme "smegg.me/thughunter/common/theme"
)

const (
	EventThemeChanged = "theme:changed"
)

// Service is the Wails-bound theme service. It delegates all detection
// logic to common/theme and bridges it to the frontend via Wails bindings
// and events.
type Service struct{}

// GetTheme returns the current device theme info.
func (s *Service) GetTheme() apptheme.Info {
	return currentTheme()
}

// RegisterThemeWatcher sets up listeners for OS theme changes and emits
// a theme:changed event to the frontend whenever the system theme or
// accent color changes.
func RegisterThemeWatcher(app *application.App) {
	app.Event.OnApplicationEvent(events.Common.ThemeChanged, func(event *application.ApplicationEvent) {
		app.Event.Emit(EventThemeChanged, currentTheme())
	})

	apptheme.WatchAccentChanges(func() {
		app.Event.Emit(EventThemeChanged, currentTheme())
	})
}

func currentTheme() apptheme.Info {
	app := application.Get()
	info := apptheme.Snapshot()

	// Enrich with Wails-provided fallbacks when native checks are incomplete.
	if !info.DarkMode {
		if dark, ok := apptheme.DetectDarkMode(); !ok && app != nil && app.Env != nil {
			_ = dark
			info.DarkMode = app.Env.IsDarkMode()
		}
	}

	if info.AccentColor == "" && app != nil && app.Env != nil {
		if wailsAccent := app.Env.GetAccentColor(); wailsAccent != "" {
			info.AccentColor = apptheme.NormalizeToHex(wailsAccent)
			info.AccentWatchSupported = info.AccentColor != ""
		}
	}

	return info
}
