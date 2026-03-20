// assets/tray/embed.go
package trayicons

import "embed"

//go:embed dark/windows/*.png dark/linux-png-fallback/*.png
//go:embed light/windows/*.png light/linux-png-fallback/*.png
var FS embed.FS
