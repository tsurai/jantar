package jantar

import (
	"fmt"
	"io"
	"log"
	"os"
)

// Defines the various logging level
const (
	LogLevelDebug   = iota
	LogLevelInfo    = iota
	LogLevelWarning = iota
	LogLevelError   = iota
	LogLevelFatal   = iota
	LogLevelPanic   = iota
)

const (
	nocolor = 0
	red     = 31
	green   = 32
	yellow  = 33
	blue    = 34
)

// JLogger is a simple Log with some handy functions. Largely inspired by https://github.com/Sirupsen/logrus
type JLogger struct {
	log      *log.Logger
	minLevel uint
	ansiMode bool
}

// JLData is a type describing map that can be passed to Data<Level>f functions for debug output
type JLData map[string]interface{}

func levelToColor(level uint) int {
	switch level {
	case LogLevelInfo:
		return blue
	case LogLevelDebug:
		return green
	case LogLevelWarning:
		return yellow
	case LogLevelError:
		fallthrough
	case LogLevelFatal:
		fallthrough
	case LogLevelPanic:
		return red
	default:
		return nocolor
	}
}

func levelToString(level uint) string {
	switch level {
	case LogLevelInfo:
		return "INFO"
	case LogLevelDebug:
		return "DEBU"
	case LogLevelWarning:
		return "WARN"
	case LogLevelError:
		return "ERRO"
	case LogLevelFatal:
		return "FATA"
	case LogLevelPanic:
		return "PANI"
	default:
		return "Unknown"
	}
}

// NewJLogger creates a new JLogger instance
func NewJLogger(out io.Writer, prefix string, minLevel uint) *JLogger {
	ansiMode := false
	if isTerminal(os.Stderr.Fd()) {
		ansiMode = true
	}

	return &JLogger{log.New(out, prefix, 0), minLevel, ansiMode}
}

// SetMinLevel changes the loggers current minimal logging level
func (l *JLogger) SetMinLevel(minLevel uint) {
	if minLevel < LogLevelInfo || minLevel > LogLevelPanic {
		l.minLevel = LogLevelInfo
	} else {
		l.minLevel = minLevel
	}
}

// Infodf is equivalent to Infof but takes additional data to display
func (l *JLogger) Infodf(data JLData, format string, v ...interface{}) {
	l.printData(LogLevelInfo, data, format, v...)
}

// Debugdf is equivalent to Infodf but uses a different logging level
func (l *JLogger) Debugdf(data JLData, format string, v ...interface{}) {
	l.printData(LogLevelDebug, data, format, v...)
}

// Warningdf is equivalent to Infodf but uses a different logging level
func (l *JLogger) Warningdf(data JLData, format string, v ...interface{}) {
	l.printData(LogLevelWarning, data, format, v...)
}

// Errordf is equivalent to Infodf but uses a different logging level
func (l *JLogger) Errordf(data JLData, format string, v ...interface{}) {
	l.printData(LogLevelError, data, format, v...)
}

// Fataldf is equivalent to Infodf but uses a different logging level and calls fatal afterwards
func (l *JLogger) Fataldf(data JLData, format string, v ...interface{}) {
	l.printData(LogLevelFatal, data, format, v...)
}

// Panicdf is equivalent to Infodf but uses a different logging level and calls fatal afterwards
func (l *JLogger) Panicdf(data JLData, format string, v ...interface{}) {
	l.printData(LogLevelPanic, data, format, v...)
}

// Infof is similar to log.Printf but uses a special markup indicating the message severity
func (l *JLogger) Infof(format string, v ...interface{}) {
	l.print(LogLevelInfo, format, v...)
}

// Debugf is equivalent to Infof but uses a different logging level
func (l *JLogger) Debugf(format string, v ...interface{}) {
	l.print(LogLevelDebug, format, v...)
}

// Warningf is equivalent to Infof but uses a different logging level
func (l *JLogger) Warningf(format string, v ...interface{}) {
	l.print(LogLevelWarning, format, v...)
}

// Errorf is equivalent to Infof but uses a different logging level
func (l *JLogger) Errorf(format string, v ...interface{}) {
	l.print(LogLevelFatal, format, v...)
}

// Fatalf is equivalent to Infof but uses a different logging level and calls fatal afterwards
func (l *JLogger) Fatalf(format string, v ...interface{}) {
	l.print(LogLevelFatal, format, v...)
}

// Panicf is equivalent to Infof but uses a different logging level and calls panic afterwards
func (l *JLogger) Panicf(format string, v ...interface{}) {
	l.print(LogLevelPanic, format, v...)
}

// Infod is equivalent to Info but takes additional data to display
func (l *JLogger) Infod(data JLData, v ...interface{}) {
	l.printData(LogLevelInfo, data, "", v...)
}

// Debugd is equivalent to Infod but uses a different logging level
func (l *JLogger) Debugd(data JLData, v ...interface{}) {
	l.printData(LogLevelDebug, data, "", v...)
}

// Warningd is equivalent to Infod but uses a different logging level
func (l *JLogger) Warningd(data JLData, v ...interface{}) {
	l.printData(LogLevelWarning, data, "", v...)
}

// Errord is equivalent to Infod but uses a different logging level
func (l *JLogger) Errord(data JLData, v ...interface{}) {
	l.printData(LogLevelError, data, "", v...)
}

// Fatald is equivalent to Infod but uses a different logging level and calls fatal afterwards
func (l *JLogger) Fatald(data JLData, v ...interface{}) {
	l.printData(LogLevelFatal, data, "", v...)
}

// Panicd is equivalent to Infod but uses a different logging level and calls panic afterwards
func (l *JLogger) Panicd(data JLData, v ...interface{}) {
	l.printData(LogLevelPanic, data, "", v...)
}

// Info is similar to log.Print but uses a special markup indicating the message severity
func (l *JLogger) Info(v ...interface{}) {
	l.print(LogLevelInfo, "", v...)
}

// Debug is equivalent to Info but uses a different logging level
func (l *JLogger) Debug(v ...interface{}) {
	l.print(LogLevelDebug, "", v...)
}

// Warning is equivalent to Info but uses a different logging level
func (l *JLogger) Warning(v ...interface{}) {
	l.print(LogLevelWarning, "", v...)
}

// Error is equivalent to Info but uses a different logging level
func (l *JLogger) Error(v ...interface{}) {
	l.print(LogLevelError, "", v...)
}

// Fatal is equivalent to Info but uses a different logging level and calls fatal afterwards
func (l *JLogger) Fatal(v ...interface{}) {
	l.print(LogLevelFatal, "", v...)
}

// Panic is equivalent to Info but uses a different logging level and calls panic afterwards
func (l *JLogger) Panic(v ...interface{}) {
	l.print(LogLevelPanic, "", v...)
}

func (l *JLogger) printData(level uint, data JLData, format string, v ...interface{}) {
	if level >= l.minLevel {
		var msg string

		if format != "" {
			msg = fmt.Sprintf(format, v...)
		} else {
			msg = fmt.Sprint(v...)
		}

		if l.ansiMode {
			color := levelToColor(level)
			out := fmt.Sprintf("\x1b[%dm[%s]\x1b[0m %s\n     \x1b[%[1]dm→ ", color, levelToString(level), msg)

			for key, val := range data {
				out += fmt.Sprintf("\x1b[%dm%v\x1b[0m=%v ", color, key, val)
			}

			l.log.Output(2, out)
		} else {
			out := fmt.Sprintf("[%s] %s\n     → ", levelToString(level), msg)

			for key, val := range data {
				out += fmt.Sprintf("%v=%v ", key, val)
			}

			l.log.Output(2, out)
		}

		if level == LogLevelFatal {
			os.Exit(1)
		} else if level == LogLevelPanic {
			panic(msg)
		}
	}
}

func (l *JLogger) print(level uint, format string, v ...interface{}) {
	if level >= l.minLevel {
		var msg string

		if format != "" {
			msg = fmt.Sprintf(format, v...)
		} else {
			msg = fmt.Sprint(v...)
		}

		if l.ansiMode {
			l.log.Output(2, fmt.Sprintf("\x1b[%dm[%s]\x1b[0m %s", levelToColor(level), levelToString(level), msg))
		} else {
			l.log.Output(2, fmt.Sprintf("[%s] %s", levelToString(level), msg))
		}

		if level == LogLevelFatal {
			os.Exit(1)
		} else if level == LogLevelPanic {
			panic(msg)
		}
	}
}
