//go:build windows

package theme

import "golang.org/x/sys/windows/registry"

func detectDarkMode() (bool, bool) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if err != nil {
		return false, false
	}
	defer k.Close()

	v, _, err := k.GetIntegerValue("AppsUseLightTheme")
	if err != nil {
		return false, false
	}
	return v == 0, true
}
