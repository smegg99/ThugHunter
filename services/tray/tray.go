// services/tray/tray.go
package tray

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"time"

	trayicons "smegg.me/thughunter/assets/tray"
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/common/theme"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// ScraperState represents the current scraper lifecycle phase for icon selection.
type ScraperState int

const (
	StateIdle     ScraperState = iota // Not running - no badge
	StateStarting                     // Initializing - badge color1 (blue)
	StateRunning                      // In progress - badge color2 (green)
	StateStopping                     // Stopping - badge color3 (orange)
)

func iconFilename(themeName string, state ScraperState) string {
	base := fmt.Sprintf("thughunter-tray-%s", themeName)
	switch state {
	case StateStarting:
		return base + "-badge-color1.png"
	case StateRunning:
		return base + "-badge-color2.png"
	case StateStopping:
		return base + "-badge-color3.png"
	default:
		return base + ".png"
	}
}

func loadIcon(themeName string, state ScraperState) ([]byte, error) {
	filename := iconFilename(themeName, state)

	var dir string
	switch runtime.GOOS {
	case "windows":
		dir = fmt.Sprintf("%s/windows", themeName)
	default:
		dir = fmt.Sprintf("%s/linux-png-fallback", themeName)
	}

	return trayicons.FS.ReadFile(dir + "/" + filename)
}

// Manager drives the system tray icon based on theme and scraper state.
type Manager struct {
	tray      *application.SystemTray
	state     ScraperState
	dark      bool
	startedAt time.Time // when StateStarting began
	graceMs   int       // random startup grace period in ms
}

// NewManager creates a Manager and sets the initial tray icon.
func NewManager(tray *application.SystemTray) *Manager {
	dark := false
	if d, ok := theme.DetectDarkMode(); ok {
		dark = d
	}
	m := &Manager{tray: tray, state: StateIdle, dark: dark}
	m.apply()
	return m
}

// UpdateState sets the scraper state and refreshes the tray icon.
func (m *Manager) UpdateState(state ScraperState) {
	if m.state == state {
		return
	}
	logger.Debug().Int("old", int(m.state)).Int("new", int(state)).Msg("tray: state transition")
	m.state = state
	m.apply()
}

// UpdateTheme sets the dark-mode flag and refreshes the tray icon.
func (m *Manager) UpdateTheme(dark bool) {
	if m.dark == dark {
		return
	}
	m.dark = dark
	m.apply()
}

func (m *Manager) themeName() string {
	if m.dark {
		return "dark"
	}
	return "light"
}

func (m *Manager) apply() {
	icon, err := loadIcon(m.themeName(), m.state)
	if err != nil {
		logger.Error().Err(err).
			Str("theme", m.themeName()).
			Int("state", int(m.state)).
			Msg("failed to load tray icon")
		return
	}
	m.tray.SetIcon(icon)

	// Keep the frontend badge colour in sync with the tray icon.
	if app := application.Get(); app != nil {
		app.Event.Emit("tray:state_changed", int(m.state))
	}
}

// RegisterEvents wires the Manager to Wails application events so the tray
// icon updates automatically when the scraper state or OS theme changes.
func (m *Manager) RegisterEvents(app *application.App) {
	app.Event.On("scraper:service:run_state_changed", func(e *application.CustomEvent) {
		running, _ := e.Data.(bool)
		if running {
			m.startedAt = time.Now()
			m.graceMs = 2000 + rand.IntN(2001) // 2–4 s
			m.UpdateState(StateStarting)
		} else {
			m.state = StateIdle
			m.apply()
		}
	})

	app.Event.On("scraper:progress", func(_ *application.CustomEvent) {
		if m.state == StateIdle || m.state == StateStopping {
			return
		}
		// Keep the blue "starting" badge visible for a short grace period.
		if m.state == StateStarting && time.Since(m.startedAt) < time.Duration(m.graceMs)*time.Millisecond {
			return
		}
		m.UpdateState(StateRunning)
	})

	app.Event.On("scraper:service:stopping", func(_ *application.CustomEvent) {
		if m.state == StateStarting || m.state == StateRunning {
			m.UpdateState(StateStopping)
		}
	})

	// Re-detect dark mode directly from the OS when theme changes,
	// rather than relying on type-asserting the event payload.
	detectAndUpdate := func() {
		dark := false
		if d, ok := theme.DetectDarkMode(); ok {
			dark = d
		} else if a := application.Get(); a != nil && a.Env != nil {
			dark = a.Env.IsDarkMode()
		}
		m.UpdateTheme(dark)
	}

	// Wails native theme-changed application event.
	app.Event.OnApplicationEvent(events.Common.ThemeChanged, func(_ *application.ApplicationEvent) {
		detectAndUpdate()
	})

	// Custom theme:changed event (fired by accent/DBus watcher).
	app.Event.On("theme:changed", func(_ *application.CustomEvent) {
		detectAndUpdate()
	})
}
