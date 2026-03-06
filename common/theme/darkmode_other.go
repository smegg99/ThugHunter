//go:build !(linux || windows)

package theme

func detectDarkMode() (bool, bool) {
	return false, false
}
