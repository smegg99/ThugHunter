package theme

import (
	"fmt"
	"image/color"
)

// ColorSource identifies the platform mechanism used to retrieve the accent color.
type ColorSource string

const (
	XDGPortalSource       ColorSource = "xdg-portal"
	WindowsRegistrySource ColorSource = "windows-registry"
)

// Color holds a resolved platform accent color.
type Color struct {
	RGBA   color.RGBA
	OK     bool
	Source ColorSource
}

// Hex returns the color as a #rrggbb string, or "" if not available.
func (c Color) Hex() string {
	if !c.OK {
		return ""
	}
	return fmt.Sprintf("#%02x%02x%02x", c.RGBA.R, c.RGBA.G, c.RGBA.B)
}

// GetAccentColor returns the platform accent color when available.
func GetAccentColor() Color {
	return getAccentColor()
}

// DetectDarkMode returns whether dark mode is active according to native
// OS APIs. The second return value is false when the check is inconclusive
// (unsupported platform).
func DetectDarkMode() (dark bool, ok bool) {
	return detectDarkMode()
}

// WatchAccentChanges starts a platform-specific listener for accent/color-scheme
// changes. The onChange callback fires whenever a change is detected. Returns a
// stop function to tear down the watcher.
func WatchAccentChanges(onChange func()) (stop func()) {
	return watchAccentChanges(onChange)
}
