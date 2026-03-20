// core/screenshot/external.go
package screenshot

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/common/templating"

	"context"
)

const maxExternalAttempts = 2

type templateData struct {
	IP      string
	PORT    string
	OUTPUT  string
	TIMEOUT string
	DELAY   string
	PAUSE   string
}

// captureExternal resolves the command template, runs it, and reads the
// output image. Retries once on non-timeout failures.
func captureExternal(ctx context.Context, ip string, port, timeoutSec int, cfg Config) ([]byte, error) {
	if cfg.Template == "" {
		return nil, fmt.Errorf("screenshot command template is not configured")
	}

	tmpDir, err := os.MkdirTemp("", "thughunter-screenshot-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	outFile := filepath.Join(tmpDir, "screenshot.png")

	resolved, err := resolveTemplate(cfg, ip, port, timeoutSec, outFile)
	if err != nil {
		return nil, err
	}

	logger.Info().Str("command", resolved).Str("ip", ip).Int("port", port).Msg("screenshot: running external command")

	extStart := time.Now()
	if err := runWithRetries(ctx, resolved, ip, port, extStart); err != nil {
		return nil, err
	}

	imgData, outFile, err := readOutputFile(tmpDir, outFile)
	if err != nil {
		return nil, err
	}

	if err := rejectIfBlank(cfg, outFile, ip, port, extStart); err != nil {
		return nil, err
	}

	logger.Info().Str("ip", ip).Int("port", port).Int("bytes", len(imgData)).Dur("elapsed", time.Since(extStart)).Msg("screenshot: external tool succeeded")
	return imgData, nil
}

// resolveTemplate fills in the command template with capture parameters.
func resolveTemplate(cfg Config, ip string, port, timeoutSec int, outFile string) (string, error) {
	resolved, err := templating.Resolve(cfg.Template, templateData{
		IP:      ip,
		PORT:    strconv.Itoa(port),
		OUTPUT:  outFile,
		TIMEOUT: strconv.Itoa(timeoutSec),
		DELAY:   strconv.Itoa(cfg.DelaySeconds),
		PAUSE:   strconv.Itoa(cfg.PauseSeconds),
	})
	if err != nil {
		return "", fmt.Errorf("resolve screenshot template: %w", err)
	}
	return resolved, nil
}

// runWithRetries executes the shell command up to maxExternalAttempts times.
func runWithRetries(ctx context.Context, resolved, ip string, port int, start time.Time) error {
	var lastErr error
	for attempt := 1; attempt <= maxExternalAttempts; attempt++ {
		if ctx.Err() != nil {
			logger.Warn().Str("ip", ip).Int("port", port).Msg("screenshot: external tool cancelled (context done)")
			return ctx.Err()
		}

		if err := runOnce(ctx, resolved); err != nil {
			logAttemptFailure(ip, port, attempt, start, err, ctx)
			lastErr = fmt.Errorf("screenshot command failed (attempt %d/%d): %w", attempt, maxExternalAttempts, err)
			if !shouldRetry(ctx, attempt) {
				continue
			}
			select {
			case <-time.After(500 * time.Millisecond):
			case <-ctx.Done():
				return ctx.Err()
			}
			continue
		}
		return nil
	}
	return lastErr
}

// runOnce executes the resolved command in its own process group so the
// entire tree can be killed on cancellation.
func runOnce(ctx context.Context, resolved string) error {
	var stderrBuf bytes.Buffer
	cmd := exec.CommandContext(ctx, "sh", "-c", resolved)
	cmd.Stdout = nil
	cmd.Stderr = &stderrBuf
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	cmd.WaitDelay = 2 * time.Second

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, stderrBuf.String())
	}
	return nil
}

// shouldRetry returns true if another attempt is worthwhile.
func shouldRetry(ctx context.Context, attempt int) bool {
	return attempt < maxExternalAttempts && ctx.Err() == nil
}

// logAttemptFailure logs why an external tool attempt failed.
func logAttemptFailure(ip string, port, attempt int, start time.Time, err error, ctx context.Context) {
	if ctx.Err() != nil {
		logger.Warn().Str("ip", ip).Int("port", port).Int("attempt", attempt).Dur("elapsed", time.Since(start)).Msg("screenshot: external tool timed out")
	} else {
		logger.Warn().Err(err).Str("ip", ip).Int("port", port).Int("attempt", attempt).Msg("screenshot: external tool failed")
	}
}

// readOutputFile reads the screenshot from tmpDir. Falls back to globbing
// if the exact path doesn't exist (some tools append their own extension).
func readOutputFile(tmpDir, outFile string) ([]byte, string, error) {
	imgData, err := os.ReadFile(outFile)
	if err != nil {
		matches, _ := filepath.Glob(filepath.Join(tmpDir, "screenshot*"))
		if len(matches) == 0 {
			return nil, "", fmt.Errorf("read screenshot output: %w", err)
		}
		imgData, err = os.ReadFile(matches[0])
		if err != nil {
			return nil, "", fmt.Errorf("read screenshot output: %w", err)
		}
		outFile = matches[0]
	}
	if len(imgData) == 0 {
		return nil, "", fmt.Errorf("screenshot output file is empty")
	}
	return imgData, outFile, nil
}

// rejectIfBlank checks the output image for blankness when configured.
func rejectIfBlank(cfg Config, outFile, ip string, port int, start time.Time) error {
	if !cfg.RejectBlank {
		return nil
	}
	isBlank, err := isBlankImageFile(outFile)
	if err != nil {
		logger.Debug().Err(err).Msg("could not decode screenshot for blank check, keeping it")
		return nil
	}
	if isBlank {
		logger.Info().Str("ip", ip).Int("port", port).Dur("elapsed", time.Since(start)).Msg("screenshot: external tool produced blank image")
		return ErrBlank
	}
	return nil
}
