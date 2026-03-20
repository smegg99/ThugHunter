// services/logger/logger.go
package logger

import (
	"strings"

	"github.com/rs/zerolog"

	appconfig "smegg.me/thughunter/common/config"
	applogger "smegg.me/thughunter/common/logger"
)

const (
	frontendPrefix = "[frontend] "
	frontendSuffix = ""
)

type Service struct{}

type FrontendLoggerConfig struct {
	ConsoleLog bool `json:"console_log"`
}

// GetFrontendLoggerConfig returns the current frontend logger config.
func (s *Service) GetFrontendLoggerConfig() FrontendLoggerConfig {
	cfg := appconfig.Get()
	return FrontendLoggerConfig{
		ConsoleLog: cfg.Logger.FrontendConsoleLog,
	}
}

// LogTrace logs a trace-level message from the frontend.
func (s *Service) LogTrace(msg string, fields map[string]any) {
	emit(applogger.Trace(), msg, fields)
}

// LogDebug logs a debug-level message from the frontend.
func (s *Service) LogDebug(msg string, fields map[string]any) {
	emit(applogger.Debug(), msg, fields)
}

// LogInfo logs an info-level message from the frontend.
func (s *Service) LogInfo(msg string, fields map[string]any) {
	emit(applogger.Info(), msg, fields)
}

// LogWarn logs a warn-level message from the frontend.
func (s *Service) LogWarn(msg string, fields map[string]any) {
	emit(applogger.Warn(), msg, fields)
}

// LogError logs an error-level message from the frontend.
func (s *Service) LogError(msg string, fields map[string]any) {
	emit(applogger.Error(), msg, fields)
}

func emit(ev *zerolog.Event, msg string, fields map[string]any) {
	ev.Str("source", "frontend")
	for k, v := range fields {
		ev.Interface(k, v)
	}
	ev.Msg(formatMessage(msg))
}

func formatMessage(msg string) string {
	var b strings.Builder
	b.Grow(len(frontendPrefix) + len(msg) + len(frontendSuffix))
	b.WriteString(frontendPrefix)
	b.WriteString(msg)
	b.WriteString(frontendSuffix)
	return b.String()
}
