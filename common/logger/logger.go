// common/logger/logger.go
package logger

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"smegg.me/thughunter/common/config"
)

func Initialize() {
	console := newConsoleWriter(false)
	log.Logger = zerolog.New(console).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
}

func ReInitialize(cfg config.LoggerConfig) error {
	if cfg.Verbose {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	console := newConsoleWriter(cfg.NoColor)

	writers := []io.Writer{console}

	if cfg.Dir != "" {
		fw, err := newFileWriter(&cfg)
		if err != nil {
			return err
		}
		writers = append(writers, fw)
	}

	multi := zerolog.MultiLevelWriter(writers...)
	ctx := zerolog.New(multi).With().Timestamp()

	if cfg.ShowCaller {
		ctx = ctx.Caller()
	}

	log.Logger = ctx.Logger()
	return nil
}

func Get() zerolog.Logger {
	return log.Logger
}

func Writer() io.Writer {
	return log.Logger.With().Logger().Output(os.Stderr)
}

func Trace() *zerolog.Event        { return log.Trace() }
func Debug() *zerolog.Event        { return log.Debug() }
func Info() *zerolog.Event         { return log.Info() }
func Warn() *zerolog.Event         { return log.Warn() }
func Error() *zerolog.Event        { return log.Error() }
func Err(err error) *zerolog.Event { return log.Err(err) }
func Fatal() *zerolog.Event        { return log.Fatal() }
func Panic() *zerolog.Event        { return log.Panic() }
