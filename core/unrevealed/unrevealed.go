package unrevealed

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// Config controls the behavior of the undetected browser.
type Config struct {
	ChromePath        string // Optional path to the Chrome executable. If empty, the package will attempt to find it automatically.
	Headless          bool   // Whether to run Chrome in headless mode. Ignored if VirtualDisplay is true.
	VirtualDisplay    bool
	UserDataDir       string
	ExtraArgs         []string
	WindowWidth       int
	WindowHeight      int
	Language          string
	ConnectTimeout    time.Duration
	NoSandbox         bool
	ChromeDriverPath  string
	PatchChromeDriver bool
}

func (c *Config) withDefaults() {
	if c.WindowWidth == 0 {
		c.WindowWidth = 1920
	}
	if c.WindowHeight == 0 {
		c.WindowHeight = 1080
	}
	if c.Language == "" {
		c.Language = "en-US"
	}
	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = 15 * time.Second
	}
}

// Browser wraps a [rod.Browser] with the underlying Chrome process,
// providing safe cleanup via [Browser.Close].
type Browser struct {
	*rod.Browser
	cmd     *exec.Cmd
	tmpDir  string
	patcher *Patcher
	xvfb    *Xvfb
	mu      sync.Mutex
	closed  bool
}

// Close shuts down the browser, kills Chrome, and removes temporary files.
// Safe to call multiple times.
func (b *Browser) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil
	}
	b.closed = true

	var errs []error
	if b.Browser != nil {
		if err := b.Browser.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close browser: %w", err))
		}
	}
	if b.cmd != nil && b.cmd.Process != nil {
		_ = b.cmd.Process.Kill()
		_, _ = b.cmd.Process.Wait()
	}
	if b.patcher != nil {
		b.patcher.Cleanup()
	}
	if b.xvfb != nil {
		_ = b.xvfb.Close()
	}
	cleanupTmpDir(b.tmpDir)
	return errors.Join(errs...)
}

// New launches a stealth-configured Chrome and connects go-rod to it.
// Set [Config.PatchChromeDriver] or [Config.ChromeDriverPath] to launch
// through a patched ChromeDriver instead of directly.
func New(ctx context.Context, cfg Config) (*Browser, error) {
	cfg.withDefaults()

	// When VirtualDisplay is requested we force Headless off
	// Chrome will run headful inside the invisible Xvfb framebuffer.
	var xvfb *Xvfb
	if cfg.VirtualDisplay {
		cfg.Headless = false
		var err error
		xvfb, err = StartXvfb(cfg.WindowWidth, cfg.WindowHeight)
		if err != nil {
			return nil, fmt.Errorf("start xvfb: %w", err)
		}
	}

	chromePath, err := resolveChromePath(cfg.ChromePath)
	if err != nil {
		if xvfb != nil {
			_ = xvfb.Close()
		}
		return nil, fmt.Errorf("find chrome: %w", err)
	}

	var browser *Browser
	if cfg.ChromeDriverPath != "" || cfg.PatchChromeDriver {
		browser, err = newViaChromeDriver(ctx, chromePath, cfg)
	} else {
		browser, err = newDirect(ctx, chromePath, cfg, xvfb)
	}
	if err != nil {
		if xvfb != nil {
			_ = xvfb.Close()
		}
		return nil, err
	}

	browser.xvfb = xvfb
	return browser, nil
}

// Stealth injects anti-detection scripts into a rod page.
// Call before navigating. Scripts persist across navigations.
func Stealth(page *rod.Page) error {
	// _ = proto.EmulationClearDeviceMetricsOverride{}.Call(page)

	for _, script := range StealthScripts() {
		_, err := proto.PageAddScriptToEvaluateOnNewDocument{
			Source: script,
		}.Call(page)
		if err != nil {
			return fmt.Errorf("inject stealth script: %w", err)
		}
	}
	return nil
}

func resolveChromePath(path string) (string, error) {
	if path != "" {
		return path, nil
	}
	return FindChrome()
}

func setupDataDir(userDir string) (userDataDir, tmpDir string, err error) {
	if userDir != "" {
		return userDir, "", nil
	}
	tmp, err := os.MkdirTemp("", "unrevealed-*")
	if err != nil {
		return "", "", err
	}
	return tmp, tmp, nil
}

func freePort() (int, net.Listener, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, nil, err
	}
	return l.Addr().(*net.TCPAddr).Port, l, nil
}

func cleanupTmpDir(dir string) {
	if dir != "" {
		os.RemoveAll(dir)
	}
}
