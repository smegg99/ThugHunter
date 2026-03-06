// common/logger/formatter.go
package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
)

var (
	traceStyle = color.New(color.FgHiBlue)
	debugStyle = color.New(color.FgHiGreen)
	infoStyle  = color.New(color.FgHiCyan)
	warnStyle  = color.New(color.FgHiYellow)
	errorStyle = color.New(color.FgHiRed)
	fatalStyle = color.New(color.FgHiRed, color.Bold)
	panicStyle = color.New(color.FgHiRed, color.ReverseVideo)
	fieldStyle = color.New(color.Faint)
)

func newConsoleWriter(noColor bool) zerolog.ConsoleWriter {
	color.NoColor = noColor

	w := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.DateTime,
		NoColor:    noColor,
	}

	w.FormatLevel = func(i interface{}) string {
		lvl := strings.ToUpper(fmt.Sprintf("%s", i))
		switch lvl {
		case "TRACE":
			return traceStyle.Sprint("TRACE")
		case "DEBUG":
			return debugStyle.Sprint("DEBUG")
		case "INFO":
			return infoStyle.Sprint("INFO ")
		case "WARN":
			return warnStyle.Sprint("WARN ")
		case "ERROR":
			return errorStyle.Sprint("ERROR")
		case "FATAL":
			return fatalStyle.Sprint("FATAL")
		case "PANIC":
			return panicStyle.Sprint("PANIC")
		default:
			return lvl
		}
	}

	w.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}

	w.FormatFieldName = func(i interface{}) string {
		return fieldStyle.Sprintf("%s=", i)
	}

	w.FormatFieldValue = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}

	return w
}
