//go:build windows

package theme

import (
	"encoding/binary"
	"image/color"

	"golang.org/x/sys/windows/registry"
)

// Gets the accent color from the Windows registry.
func getAccentColor() Color {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\DWM`, registry.QUERY_VALUE)
	if err != nil {
		return Color{OK: false, Source: WindowsRegistrySource}
	}
	defer k.Close()

	if v, _, err := k.GetIntegerValue("ColorizationColor"); err == nil {
		argb := uint32(v)
		return Color{
			RGBA:   color.RGBA{R: uint8(argb >> 16), G: uint8(argb >> 8), B: uint8(argb), A: 0xFF},
			OK:     true,
			Source: WindowsRegistrySource,
		}
	}

	k2, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Explorer\Accent`, registry.QUERY_VALUE)
	if err != nil {
		return Color{OK: false, Source: WindowsRegistrySource}
	}
	defer k2.Close()

	if b, _, err := k2.GetBinaryValue("AccentPalette"); err == nil && len(b) >= 4 {
		u := binary.LittleEndian.Uint32(b[0:4])
		bb := uint8(u & 0xFF)
		gg := uint8((u >> 8) & 0xFF)
		rr := uint8((u >> 16) & 0xFF)
		return Color{
			RGBA:   color.RGBA{R: rr, G: gg, B: bb, A: 0xFF},
			OK:     true,
			Source: WindowsRegistrySource,
		}
	}

	return Color{OK: false, Source: WindowsRegistrySource}
}
