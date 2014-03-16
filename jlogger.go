package jantar

import (
	"io"
	"os"
	"fmt"
	"log"
)

const (
	level_debug 	= iota
	level_info 		= iota
	level_warning = iota
	level_error   = iota
	level_fatal 	= iota

	nocolor = 0
	red 		= 31
	green 	= 32
	yellow 	= 33
	blue 		= 34
)

// JLogger is a simple logger with some handy functions. Largely inspired by https://github.com/Sirupsen/logrus
type JLogger struct {
	log 			*log.Logger
	minlevel 	uint
	ansiMode	bool
}

// JLData is a type describing map that can be passed to Data<Level>f functions for debug output
type JLData map[string]interface{}

func levelToColor(level uint) int {
	switch level {
	case level_info:
		return blue
	case level_debug:
		return green
	case level_warning:
		return yellow
	case level_error:
		return red
	case level_fatal:
		return red
	default:
		return nocolor
	}
}

func levelToString(level uint) string {
	switch level {
	case level_info:
		return "INFO"
	case level_debug:
		return "DEBU"
	case level_warning:
		return "WARN"
	case level_error:
		return "ERRO"
	case level_fatal:
		return "FATA"
	default:
		return "Unknown"
	}
}

func NewJLogger(out io.Writer, prefix string, minlevel uint) *JLogger {
	ansiMode := false
	if isTerminal(os.Stderr.Fd()) {
		ansiMode = true
	}

	return &JLogger{log.New(out, prefix, 0), minlevel, ansiMode}
}

func (l *JLogger) DataInfof(data JLData, format string, v ...interface{}) {
	l.printData(fmt.Sprintf(format, v...), level_info, data)
}

func (l *JLogger) DataDebugf(data JLData, format string, v ...interface{}) {
	l.printData(fmt.Sprintf(format, v...), level_debug, data)
}

func (l *JLogger) DataWarningf(data JLData, format string, v ...interface{}) {
	l.printData(fmt.Sprintf(format, v...), level_warning, data)
}

func (l *JLogger) DataErrorf(data JLData, format string, v ...interface{}) {
	l.printData(fmt.Sprintf(format, v...), level_error, data)
}

func (l *JLogger) DataFatalf(data JLData, format string, v ...interface{}) {
	l.printData(fmt.Sprintf(format, v...), level_fatal, data)
}

func (l *JLogger) Infof(format string, v ...interface{}) {
	l.printLevel(fmt.Sprintf(format, v...), level_info)
}

func (l *JLogger) Debugf(format string, v ...interface{}) {
	l.printLevel(fmt.Sprintf(format, v...), level_debug)
}

func (l *JLogger) Warningf(format string, v ...interface{}) {
	l.printLevel(fmt.Sprintf(format, v...), level_warning)
}

func (l *JLogger) Errorf(format string, v ...interface{}) {
	l.printLevel(fmt.Sprintf(format, v...), level_fatal)
}

func (l *JLogger) Fatalf(format string, v ...interface{}) {
	l.printLevel(fmt.Sprintf(format, v...), level_fatal)
}

func (l *JLogger) Info(v interface{}) {
	l.printLevel(v, level_info)
}

func (l *JLogger) Debug(v interface{}) {
	l.printLevel(v, level_debug)
}

func (l *JLogger) Warning(v interface{}) {
	l.printLevel(v, level_warning)
}

func (l *JLogger) Error(v interface{}) {
	l.printLevel(v, level_fatal)
}

func (l *JLogger) Fatal(v interface{}) {
	l.printLevel(v, level_fatal)
}

func (l *JLogger) printData(v interface{}, level uint, data JLData) {
	if level >= l.minlevel {
		if l.ansiMode {
			color := levelToColor(level)
			msg := fmt.Sprintf("\x1b[%dm[%s]\x1b[0m %v\n       \x1b[%[1]dm→ ", color, levelToString(level), v)

			for key, val := range data {
				msg += fmt.Sprintf("\x1b[%dm%v\x1b[0m=%v ", color, key, val)
			}

			l.log.Output(2, msg)
		} else {
			msg := fmt.Sprintf("[%s] %v\n       ", levelToString(level), v)
			
			for key, val := range data {
				msg += fmt.Sprintf("%v=%v ", key, val)
			}

			l.log.Output(2, msg)
		}
	}
}

func (l *JLogger) printLevel(v interface{}, level uint) {
	if level >= l.minlevel {
		if l.ansiMode {
			l.log.Output(2, fmt.Sprintf("\x1b[%dm[%s]\x1b[0m %v\n", levelToColor(level), levelToString(level), v))
		} else {
			l.log.Output(2, fmt.Sprintf("[%s] %v", levelToString(level), v))
		}
	}
}