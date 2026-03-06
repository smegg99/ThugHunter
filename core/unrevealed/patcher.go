// patcher.go
package unrevealed

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
)

const legacyMaxVersion = 114

const (
	legacyVersionURL  = "https://chromedriver.storage.googleapis.com/LATEST_RELEASE_%d"
	modernVersionURL  = "https://googlechromelabs.github.io/chrome-for-testing/latest-versions-per-milestone-with-downloads.json"
	legacyDownloadURL = "https://chromedriver.storage.googleapis.com/%s/chromedriver_%s.zip"
	modernDownloadURL = "https://storage.googleapis.com/chrome-for-testing-public/%s/%s/chromedriver-%s.zip"

	defaultMaxDownloadSize int64 = 100 << 20 // 100 MB
)

const (
	cdcPattern       = `\{window\.cdc.*?;\}`
	patchReplacement = `{console.log("unrevealed chromedriver 1337!")}`
	patchMarker      = "unrevealed chromedriver"
)

const (
	platformWin32    = "win32"
	platformWin64    = "win64"
	platformMac64    = "mac64"
	platformMacArm64 = "mac-arm64"
	platformMacX64   = "mac-x64"
	platformLinux64  = "linux64"

	exeNameWindows = "chromedriver.exe"
	exeNameUnix    = "chromedriver"
)

const dataDirName = "unrevealed"

type milestonesResponse struct {
	Milestones map[string]milestoneInfo `json:"milestones"`
}

type milestoneInfo struct {
	Version string `json:"version"`
}

// Patcher downloads and patches a ChromeDriver binary to remove
// automation detection markers injected by ChromeDriver into the browser.
type Patcher struct {
	MajorVersion int
	DriverPath   string
	DataDir      string
	platform     string
	exeName      string

	// MaxDownloadSize is the maximum permitted archive size in bytes.
	// Default: 100 MB.
	MaxDownloadSize int64

	// ExpectedSHA256 optionally pins the download to a specific hash.
	// When non-empty, Run verifies the downloaded archive matches this
	// hex-encoded SHA256 digest.
	ExpectedSHA256 string

	// DownloadSHA256 is populated after Run with the hex-encoded SHA256
	// of the downloaded archive. Can be used for auditing or pinning.
	DownloadSHA256 string
}

// NewPatcher creates a patcher for the given Chrome major version.
func NewPatcher(majorVersion int) *Patcher {
	p := &Patcher{
		MajorVersion: majorVersion,
		DataDir:      defaultDataDir(),
	}
	p.initPlatform()
	return p
}

// Run downloads the matching ChromeDriver and patches it.
// Returns the path to the patched binary.
func (p *Patcher) Run(ctx context.Context) (string, error) {
	if p.MaxDownloadSize == 0 {
		p.MaxDownloadSize = defaultMaxDownloadSize
	}

	if err := os.MkdirAll(p.DataDir, 0o755); err != nil {
		return "", fmt.Errorf("prepare: %w", err)
	}

	hex, err := randomHex(8)
	if err != nil {
		return "", fmt.Errorf("prepare: %w", err)
	}
	p.DriverPath = filepath.Join(p.DataDir, hex+"_"+p.exeName)

	version, err := p.fetchVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("fetch version: %w", err)
	}

	data, err := p.download(ctx, version)
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}

	if err := p.extract(data); err != nil {
		return "", fmt.Errorf("extract: %w", err)
	}

	if err := p.patch(); err != nil {
		return "", fmt.Errorf("patch: %w", err)
	}

	return p.DriverPath, nil
}

// IsPatched reports whether the binary at DriverPath has been patched.
func (p *Patcher) IsPatched() bool {
	data, err := os.ReadFile(p.DriverPath)
	if err != nil {
		return false
	}
	return bytes.Contains(data, []byte(patchMarker))
}

// Cleanup removes the patched binary from disk.
func (p *Patcher) Cleanup() {
	if p.DriverPath != "" {
		os.Remove(p.DriverPath)
	}
}
