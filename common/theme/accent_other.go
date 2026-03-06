//go:build !(linux || windows)

package theme

func getAccentColor() Color {
	return Color{OK: false}
}
