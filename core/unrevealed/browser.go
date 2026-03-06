// browser.go
package unrevealed

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// FindChrome locates a Chrome or Chromium executable on the system.
func FindChrome() (string, error) {
	searchByName := func() string {
		for _, name := range chromeNames() {
			if p, err := exec.LookPath(name); err == nil {
				return p
			}
		}
		return ""
	}

	searchByPath := func() string {
		for _, p := range chromePaths() {
			if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
				return p
			}
		}
		return ""
	}

	if p := searchByName(); p != "" {
		return p, nil
	}
	if p := searchByPath(); p != "" {
		return p, nil
	}
	return "", fmt.Errorf("chrome executable not found")
}

// ChromeVersion returns the full version string (e.g., "120.0.6099.109").
func ChromeVersion(path string) (string, error) {
	runVersionCmd := func() ([]byte, error) {
		switch runtime.GOOS {
		case "windows":
			return exec.Command("powershell", "-Command",
				fmt.Sprintf(`(Get-Item '%s').VersionInfo.FileVersion`, path)).Output()
		default:
			return exec.Command(path, "--version").Output()
		}
	}

	out, err := runVersionCmd()
	if err != nil {
		return "", fmt.Errorf("get chrome version: %w", err)
	}

	m := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`).FindString(string(out))
	if m == "" {
		return "", fmt.Errorf("could not parse chrome version from: %q", strings.TrimSpace(string(out)))
	}
	return m, nil
}

// ChromeMajorVersion returns the major version number of Chrome at the given path.
func ChromeMajorVersion(path string) (int, error) {
	v, err := ChromeVersion(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.SplitN(v, ".", 2)[0])
}

func chromeNames() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{"chrome", "chrome.exe"}
	case "darwin":
		return []string{"Google Chrome", "Chromium"}
	default:
		return []string{
			"google-chrome",
			"google-chrome-stable",
			"chrome",
			"chromium",
			"chromium-browser",
		}
	}
}

func chromePaths() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}
	case "windows":
		var paths []string
		for _, env := range []string{"PROGRAMFILES", "PROGRAMFILES(X86)", "LOCALAPPDATA", "PROGRAMW6432"} {
			root := os.Getenv(env)
			if root == "" {
				continue
			}
			paths = append(paths,
				filepath.Join(root, "Google", "Chrome", "Application", "chrome.exe"),
				filepath.Join(root, "Chromium", "Application", "chrome.exe"),
			)
		}
		return paths
	default:
		return []string{
			"/snap/bin/chromium",
			"/var/lib/flatpak/exports/bin/org.chromium.Chromium",
			"/var/lib/flatpak/exports/bin/com.google.Chrome",
		}
	}
}
