package theme

import (
	"fmt"
	"strings"
)

// Info holds a point-in-time snapshot of device-level theme settings.
type Info struct {
	AccentColor          string `json:"accent_color"`
	DarkMode             bool   `json:"dark_mode"`
	AccentWatchSupported bool   `json:"accent_watch_supported"`
}

// Snapshot returns the current device theme info using only native OS APIs.
// It does not depend on Wails or any GUI framework.
func Snapshot() Info {
	info := Info{}

	if dark, ok := DetectDarkMode(); ok {
		info.DarkMode = dark
	}

	if hex := GetAccentColor().Hex(); hex != "" {
		info.AccentColor = hex
	}

	info.AccentWatchSupported = info.AccentColor != ""
	return info
}

// NormalizeToHex converts rgb(R,G,B) or rgba(R,G,B,A) strings into #RRGGBB
// hex. If the input is already hex it's returned as-is.
func NormalizeToHex(color string) string {
	color = strings.TrimSpace(color)
	if strings.HasPrefix(color, "#") {
		return color
	}
	var r, g, b int

	if n, _ := fmt.Sscanf(color, "rgb(%d,%d,%d", &r, &g, &b); n == 3 {
		return fmt.Sprintf("#%02x%02x%02x", clamp(r), clamp(g), clamp(b))
	}
	if n, _ := fmt.Sscanf(color, "rgba(%d,%d,%d", &r, &g, &b); n == 3 {
		return fmt.Sprintf("#%02x%02x%02x", clamp(r), clamp(g), clamp(b))
	}
	if n, _ := fmt.Sscanf(color, "rgb(%d, %d, %d", &r, &g, &b); n == 3 {
		return fmt.Sprintf("#%02x%02x%02x", clamp(r), clamp(g), clamp(b))
	}
	if n, _ := fmt.Sscanf(color, "rgba(%d, %d, %d", &r, &g, &b); n == 3 {
		return fmt.Sprintf("#%02x%02x%02x", clamp(r), clamp(g), clamp(b))
	}
	return ""
}

func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}
