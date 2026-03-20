// core/screenshot/capture.go
package screenshot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"smegg.me/thughunter/common/logger"
)

// screenshotSem limits concurrent capture goroutines. Resized by InitSem.
var screenshotSem = make(chan struct{}, 32)

// InitSem replaces the global semaphore with one of the given size.
func InitSem(size int) {
	if size <= 0 {
		size = 32
	}
	screenshotSem = make(chan struct{}, size)
}

// ErrBlank is returned when the captured image is near-single-color.
var ErrBlank = errors.New("blank screenshot")

// Config holds resolved screenshot configuration values.
type Config struct {
	Template     string
	DelaySeconds int
	PauseSeconds int
	RejectBlank  bool
}

// Result reports the capture outcome.
type Result struct {
	Data   []byte
	Method string // "native", "external", or "" on failure
	Err    error
}

// Capture takes a VNC screenshot. Tries native Go RFB first, falls back
// to the external tool if configured. The semaphore is acquired once for
// the entire attempt (native + fallback).
func Capture(ctx context.Context, ip string, port, timeoutSec int, cfg Config) Result {
	if err := acquireSem(ctx); err != nil {
		return Result{Err: err}
	}
	defer releaseSem()

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	start := time.Now()
	logger.Info().Str("ip", ip).Int("port", port).Int("timeout", timeoutSec).Msg("screenshot: starting native capture")

	// Try native Go capture first.
	imgData, nativeErr := CaptureNative(ctx, ip, port)
	nativeBlank := nativeErr == nil && cfg.RejectBlank && !Validate(imgData)

	if nativeErr == nil && !nativeBlank {
		logger.Info().Str("ip", ip).Int("port", port).Int("bytes", len(imgData)).Dur("elapsed", time.Since(start)).Msg("screenshot: native capture succeeded")
		return Result{Data: imgData, Method: "native"}
	}

	logNativeFailure(ip, port, start, nativeErr, nativeBlank, ctx)

	return tryExternalFallback(ctx, ip, port, timeoutSec, cfg, nativeBlank, nativeErr)
}

// tryExternalFallback attempts the external tool capture, returning an
// appropriate Result whether or not a template is configured.
func tryExternalFallback(ctx context.Context, ip string, port, timeoutSec int, cfg Config, nativeBlank bool, nativeErr error) Result {
	if cfg.Template == "" {
		if nativeBlank {
			return Result{Err: ErrBlank}
		}
		return Result{Err: fmt.Errorf("native capture failed and no external tool configured: %w", nativeErr)}
	}

	logger.Info().Str("ip", ip).Int("port", port).Msg("screenshot: falling back to external tool")
	extData, extErr := captureExternal(ctx, ip, port, timeoutSec, cfg)
	if extErr != nil {
		if nativeBlank {
			return Result{Err: ErrBlank}
		}
		return Result{Err: extErr}
	}
	return Result{Data: extData, Method: "external"}
}

// acquireSem blocks until a semaphore slot is available or ctx is cancelled.
func acquireSem(ctx context.Context) error {
	select {
	case screenshotSem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func releaseSem() { <-screenshotSem }

// logNativeFailure emits a log for why native capture didn't produce a result.
func logNativeFailure(ip string, port int, start time.Time, err error, blank bool, ctx context.Context) {
	switch {
	case blank:
		logger.Info().Str("ip", ip).Int("port", port).Dur("elapsed", time.Since(start)).Msg("screenshot: native capture blank, will try external fallback")
	case ctx.Err() != nil:
		logger.Warn().Str("ip", ip).Int("port", port).Dur("elapsed", time.Since(start)).Msg("screenshot: native capture timed out")
	default:
		logger.Warn().Err(err).Str("ip", ip).Int("port", port).Dur("elapsed", time.Since(start)).Msg("screenshot: native capture failed")
	}
}
