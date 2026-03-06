// common/config/format.go
package config

import (
	"os"
	"path/filepath"
	"runtime"
)

type Format string

const (
	FormatBinary   Format = "binary"   // Plain binary, portable / development
	FormatAppImage Format = "appimage" // Linux AppImage (read-only squashfs)
	FormatFlatpak  Format = "flatpak"  // Linux Flatpak sandbox
	FormatSnap     Format = "snap"     // Linux Snap package
	FormatDeb      Format = "deb"      // Debian .deb package
	FormatRPM      Format = "rpm"      // RPM package
	FormatAUR      Format = "aur"      // Arch Linux package
	FormatNSIS     Format = "nsis"     // Windows NSIS installer
	FormatMSIX     Format = "msix"     // Windows MSIX package
	FormatExe      Format = "exe"      // Windows portable executable
	FormatMacApp   Format = "macapp"   // macOS .app bundle
	FormatDocker   Format = "docker"   // Docker container
)

const appName = "thughunter"

// Set of compile-time formats to build for. Used by build scripts and ldflags.
var buildFormat string

// DetectFormat returns the active build format.
// It prefers the compile-time value set via ldflags, then falls back to
// runtime environment variable detection.
func DetectFormat() Format {
	if buildFormat != "" {
		return Format(buildFormat)
	}
	return detectFromEnv()
}

func detectFromEnv() Format {
	if LaunchedViaAppImage() {
		return FormatAppImage
	}
	if os.Getenv("FLATPAK_ID") != "" {
		return FormatFlatpak
	}
	if os.Getenv("SNAP") != "" {
		return FormatSnap
	}
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return FormatDocker
	}
	return FormatBinary
}

// IsInstalled reports whether the format uses a system-managed or read-only
// installation directory that cannot (or should not) be written to at runtime.
func (f Format) IsInstalled() bool {
	switch f {
	case FormatAppImage, FormatFlatpak, FormatSnap, FormatMacApp,
		FormatDeb, FormatRPM, FormatAUR, FormatNSIS, FormatMSIX:
		return true
	default:
		return false
	}
}

// BundledConfigDir returns the directory that may contain a shipped default
// config.json. Returns "" when no bundled directory is known.
func (f Format) BundledConfigDir() string {
	switch f {
	case FormatAppImage:
		return os.Getenv("APPDIR")
	case FormatFlatpak:
		return "/app/share/" + appName
	case FormatSnap:
		return bundledSnapDir()
	case FormatDeb, FormatRPM, FormatAUR:
		return "/etc/" + appName
	case FormatMacApp:
		return bundledMacDir()
	case FormatNSIS, FormatMSIX:
		return bundledWindowsDir()
	default:
		return ""
	}
}

func bundledSnapDir() string {
	if d := os.Getenv("SNAP"); d != "" {
		return d
	}
	return ""
}

func bundledMacDir() string {
	exe, err := os.Executable()
	if err == nil {
		return filepath.Join(filepath.Dir(filepath.Dir(exe)), "Resources")
	}
	return ""
}

func bundledWindowsDir() string {
	exe, err := os.Executable()
	if err == nil {
		return filepath.Dir(exe)
	}
	return ""
}

// userConfigDir returns the platform-appropriate user config directory:
//   - Linux:   $XDG_CONFIG_HOME/thughunter  (default ~/.config/thughunter)
//   - macOS:   ~/Library/Application Support/thughunter
//   - Windows: %AppData%\thughunter
func userConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appName), nil
}

// userDataDir returns the platform-appropriate user data directory:
//   - Linux:   $XDG_DATA_HOME/thughunter  (default ~/.local/share/thughunter)
//   - macOS:   ~/Library/Application Support/thughunter  (same as config on macOS)
//   - Windows: %LocalAppData%\thughunter
func userDataDir() (string, error) {
	switch runtime.GOOS {
	case "linux":
		return resolveLinuxDataDir()
	case "windows":
		if d := os.Getenv("LOCALAPPDATA"); d != "" {
			return filepath.Join(d, appName), nil
		}
		return userConfigDir()
	default:
		return userConfigDir()
	}
}

// resolveLinuxDataDir returns the Linux-specific data directory.
func resolveLinuxDataDir() (string, error) {
	if d := os.Getenv("XDG_DATA_HOME"); d != "" {
		return filepath.Join(d, appName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", appName), nil
}
