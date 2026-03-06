//go:build linux

package theme

import "github.com/godbus/dbus/v5"

func detectDarkMode() (bool, bool) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return false, false
	}

	obj := conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")
	var v dbus.Variant
	call := obj.Call("org.freedesktop.portal.Settings.Read", 0, "org.freedesktop.appearance", "color-scheme")
	if call.Err != nil {
		return false, false
	}
	if err := call.Store(&v); err != nil {
		return false, false
	}

	val := v.Value()
	if vv, ok := val.(dbus.Variant); ok {
		val = vv.Value()
	}

	switch n := val.(type) {
	case uint32:
		return n == 1, true
	case int32:
		return n == 1, true
	case int64:
		return n == 1, true
	case uint64:
		return n == 1, true
	}
	return false, false
}
