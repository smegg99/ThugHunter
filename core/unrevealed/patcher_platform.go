// patcher_platform.go
package unrevealed

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func (p *Patcher) initPlatform() {
	switch runtime.GOOS {
	case "windows":
		p.platform = platformWin32
		if runtime.GOARCH == "amd64" && p.MajorVersion > legacyMaxVersion {
			p.platform = platformWin64
		}
		p.exeName = exeNameWindows
	case "darwin":
		p.platform = p.darwinPlatform()
		p.exeName = exeNameUnix
	default:
		p.platform = platformLinux64
		p.exeName = exeNameUnix
	}
}

func (p *Patcher) darwinPlatform() string {
	if p.MajorVersion <= legacyMaxVersion {
		return platformMac64
	}
	if runtime.GOARCH == "arm64" {
		return platformMacArm64
	}
	return platformMacX64
}

func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), dataDirName)
	}
	switch runtime.GOOS {
	case "windows":
		return windowsDataDir(home)
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", dataDirName)
	default:
		return filepath.Join(home, ".local", "share", dataDirName)
	}
}

func windowsDataDir(home string) string {
	if appdata := os.Getenv("APPDATA"); appdata != "" {
		return filepath.Join(appdata, dataDirName)
	}
	return filepath.Join(home, "AppData", "Roaming", dataDirName)
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random hex: %w", err)
	}
	return hex.EncodeToString(b), nil
}
