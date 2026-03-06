//go:build linux

package theme

import (
	"image/color"
	"math"

	"github.com/godbus/dbus/v5"
)

// Gets the accent color from the XDG portal.
func getAccentColor() Color {
	conn, err := dbus.SessionBus()
	if err != nil {
		return Color{OK: false, Source: XDGPortalSource}
	}

	obj := conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")

	var v dbus.Variant
	call := obj.Call("org.freedesktop.portal.Settings.Read", 0, "org.freedesktop.appearance", "accent-color")
	if call.Err != nil {
		return Color{OK: false, Source: XDGPortalSource}
	}
	if err := call.Store(&v); err != nil {
		return Color{OK: false, Source: XDGPortalSource}
	}

	val := v.Value()
	if vv, ok := val.(dbus.Variant); ok {
		val = vv.Value()
	}

	t, ok := val.([]interface{})
	if !ok || len(t) != 3 {
		return Color{OK: false, Source: XDGPortalSource}
	}

	r, ok1 := t[0].(float64)
	g, ok2 := t[1].(float64)
	b, ok3 := t[2].(float64)
	if !(ok1 && ok2 && ok3) {
		return Color{OK: false, Source: XDGPortalSource}
	}

	if r < 0 || r > 1 || g < 0 || g > 1 || b < 0 || b > 1 {
		return Color{OK: false, Source: XDGPortalSource}
	}

	toByte := func(x float64) uint8 {
		return uint8(math.Round(x * 255))
	}

	return Color{
		RGBA:   color.RGBA{R: toByte(r), G: toByte(g), B: toByte(b), A: 0xFF},
		OK:     true,
		Source: XDGPortalSource,
	}
}
