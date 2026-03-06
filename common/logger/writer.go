// common/logger/writer.go
package logger

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/DeRuina/timberjack"

	"smegg.me/thughunter/common/config"
)

const defaultRotationInterval = 24 * time.Hour

func newFileWriter(cfg *config.LoggerConfig) (io.Writer, error) {
	if err := os.MkdirAll(cfg.Dir, 0o750); err != nil {
		return nil, err
	}

	logName := cfg.LogName
	if logName == "" {
		logName = "app.log"
	}

	return &timberjack.Logger{
		Filename:         filepath.Join(cfg.Dir, logName),
		MaxSize:          int(cfg.MaxSizeMb),
		MaxBackups:       int(cfg.MaxBackups),
		MaxAge:           int(cfg.MaxAgeDays),
		RotationInterval: defaultRotationInterval,
		Compress:         true,
		LocalTime:        true,
	}, nil
}
