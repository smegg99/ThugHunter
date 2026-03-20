// core/scraper/agent_browser.go
package scraper

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"syscall"
	"time"
	"unsafe"

	"github.com/smegg99/unrevealed"
	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
)

// browserCloseTimeout is the maximum time to wait for a browser to close
// gracefully before force-killing the Chrome process.
const browserCloseTimeout = 2 * time.Second

// ensureBrowser lazily launches the browser if not already running.
func (a *ScraperAgent) ensureBrowser(ctx context.Context) error {
	if a.browser != nil {
		return nil
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	logger.Debug().Str("agent", a.Name).Msg("launching browser")

	cfg := config.Get()
	browser, err := unrevealed.New(ctx, unrevealed.Config{
		BlockFilenames: []string{"pendo.js"},
		ChromePath:     cfg.Scraper.BrowserBinaryPath,
		Headless:       false,
		Minimal:        cfg.Scraper.MinimalBrowser,
		VirtualDisplay: cfg.Scraper.VirtualDisplay,
	})
	if err != nil {
		return fmt.Errorf("agent %s: %w: %w", a.Name, ErrBrowserLaunchFailed, err)
	}

	a.browser = browser
	logger.Debug().Str("agent", a.Name).Msg("browser launched")
	return nil
}

// RelaunchBrowser closes the current browser and launches a fresh one.
func (a *ScraperAgent) RelaunchBrowser(ctx context.Context) error {
	logger.Info().Str("agent", a.Name).Msg("relaunching browser")

	a.page = nil
	a.pageReady = false
	a.account = nil
	a.estimatedCredits = 0

	if a.browser != nil {
		closeBrowserWithTimeout(a.Name, a.browser)
		a.browser = nil
	}

	return a.ensureBrowser(ctx)
}

// Close releases the agent's browser and resets its state.
func (a *ScraperAgent) Close() error {
	logger.Debug().Str("agent", a.Name).Msg("closing agent")

	a.page = nil
	a.pageReady = false

	if a.browser != nil {
		b := a.browser
		a.browser = nil

		pid, hasPID := browserPID(b)

		done := make(chan error, 1)
		go func() { done <- b.Close() }()

		select {
		case err := <-done:
			if err != nil {
				logger.Error().Err(err).Str("agent", a.Name).Msg("error closing browser")
				if hasPID {
					killBrowserProcess(pid)
				}
				return err
			}
		case <-time.After(browserCloseTimeout):
			logger.Warn().Str("agent", a.Name).Msg("browser close timed out, force-killing")
			if hasPID {
				killBrowserProcess(pid)
			}
		}
	}

	logger.Info().Str("agent", a.Name).Msg("agent closed")
	return nil
}

// ForceClose immediately kills the browser process without graceful shutdown.
func (a *ScraperAgent) ForceClose() {
	logger.Debug().Str("agent", a.Name).Msg("force-closing agent")

	a.page = nil
	a.pageReady = false

	if a.browser != nil {
		b := a.browser
		a.browser = nil

		pid, hasPID := browserPID(b)
		if hasPID {
			killBrowserProcess(pid)
		} else {
			done := make(chan error, 1)
			go func() { done <- b.Close() }()
			select {
			case <-done:
			case <-time.After(1 * time.Second):
			}
		}
	}

	logger.Info().Str("agent", a.Name).Msg("agent force-closed")
}

// closeBrowserWithTimeout attempts graceful close, then force-kills on timeout.
func closeBrowserWithTimeout(agentName string, b *unrevealed.Browser) {
	pid, hasPID := browserPID(b)
	done := make(chan error, 1)
	go func() { done <- b.Close() }()
	select {
	case <-done:
	case <-time.After(browserCloseTimeout):
		logger.Warn().Str("agent", agentName).Msg("browser close timed out during relaunch, force-killing")
		if hasPID {
			killBrowserProcess(pid)
		}
	}
}

// browserPID extracts the Chrome process PID from an unrevealed.Browser
// via reflection on the unexported cmd field.
func browserPID(b *unrevealed.Browser) (pid int, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			pid, ok = 0, false
		}
	}()
	v := reflect.ValueOf(b).Elem()
	cmdAddr := unsafe.Pointer(v.FieldByName("cmd").UnsafeAddr())
	cmd := *(**exec.Cmd)(cmdAddr)
	if cmd == nil || cmd.Process == nil {
		return 0, false
	}
	return cmd.Process.Pid, true
}

// killBrowserProcess force-kills the Chrome process tree.
func killBrowserProcess(pid int) {
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
		logger.Debug().Err(err).Int("pid", pid).Msg("SIGTERM to process group failed")
	}
	time.Sleep(200 * time.Millisecond)
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
		logger.Debug().Err(err).Int("pid", pid).Msg("process group kill failed, trying single process")
		if proc, err := os.FindProcess(pid); err == nil {
			_ = proc.Kill()
		}
	}
}
