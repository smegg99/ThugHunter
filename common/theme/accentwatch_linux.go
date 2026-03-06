//go:build linux

package theme

import (
	"github.com/godbus/dbus/v5"
)

func watchAccentChanges(onChange func()) (stop func()) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return func() {}
	}

	rule := "type='signal',interface='org.freedesktop.portal.Settings',member='SettingChanged',path='/org/freedesktop/portal/desktop'"
	if err := conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, rule).Err; err != nil {
		return func() {}
	}

	ch := make(chan *dbus.Signal, 4)
	conn.Signal(ch)

	done := make(chan struct{})
	go func() {
		for {
			select {
			case sig, ok := <-ch:
				if !ok {
					return
				}
				if sig == nil || sig.Name != "org.freedesktop.portal.Settings.SettingChanged" {
					continue
				}
				if len(sig.Body) >= 2 {
					ns, _ := sig.Body[0].(string)
					if ns == "org.freedesktop.appearance" {
						onChange()
					}
				}
			case <-done:
				return
			}
		}
	}()

	return func() {
		close(done)
		conn.RemoveSignal(ch)
		_ = conn.BusObject().Call("org.freedesktop.DBus.RemoveMatch", 0, rule).Err
	}
}
