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
	level_panic		= iota

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
		fallthrough
	case level_fatal:
		fallthrough
	case level_panic:
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
	case level_panic:
		return "PANI"
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

func (l *JLogger) Infodf(data JLData, format string, v ...interface{}) {
	l.printData(level_info, data, fmt.Sprintf(format, v...))
}

func (l *JLogger) Debugdf(data JLData, format string, v ...interface{}) {
	l.printData(level_debug, data, fmt.Sprintf(format, v...))
}

func (l *JLogger) Warningdf(data JLData, format string, v ...interface{}) {
	l.printData(level_warning, data, fmt.Sprintf(format, v...))
}

func (l *JLogger) Errordf(data JLData, format string, v ...interface{}) {
	l.printData(level_error, data, fmt.Sprintf(format, v...))
}

func (l *JLogger) Fataldf(data JLData, format string, v ...interface{}) {
	l.printData(level_fatal, data, fmt.Sprintf(format, v...))
}

func (l *JLogger) Panicdf(data JLData, format string, v ...interface{}) {
	l.printData(level_panic, data, fmt.Sprintf(format, v...))	
}

func (l *JLogger) Infof(format string, v ...interface{}) {
	l.print(level_info, fmt.Sprintf(format, v...))
}

func (l *JLogger) Debugf(format string, v ...interface{}) {
	l.print(level_debug, fmt.Sprintf(format, v...))
}

func (l *JLogger) Warningf(format string, v ...interface{}) {
	l.print(level_warning, fmt.Sprintf(format, v...))
}

func (l *JLogger) Errorf(format string, v ...interface{}) {
	l.print(level_fatal, fmt.Sprintf(format, v...))
}

func (l *JLogger) Fatalf(format string, v ...interface{}) {
	l.print(level_fatal, fmt.Sprintf(format, v...))
}

func (l *JLogger) Panicf(format string, v ...interface{}) {
	l.print(level_panic, fmt.Sprintf(format, v...))
}

func (l *JLogger) Infod(data JLData, v ...interface{}) {
	l.printData(level_info, data, fmt.Sprint(v...))
}

func (l *JLogger) Debugd(data JLData, v ...interface{}) {
	l.printData(level_debug, data, fmt.Sprint(v...))
}

func (l *JLogger) Warningd(data JLData, v ...interface{}) {
	l.printData(level_warning, data, fmt.Sprint(v...))
}

func (l *JLogger) Errord(data JLData, v ...interface{}) {
	l.printData(level_error, data, fmt.Sprint(v...))
}

func (l *JLogger) Fatald(data JLData, v ...interface{}) {
	l.printData(level_fatal, data, fmt.Sprint(v...))
}

func (l *JLogger) Panicd(data JLData, v ...interface{}) {
	l.printData(level_panic, data, fmt.Sprint(v...))	
}

func (l *JLogger) Info(v ...interface{}) {
	l.print(level_info, fmt.Sprint(v...))
}

func (l *JLogger) Debug(v ...interface{}) {
	l.print(level_debug, fmt.Sprint(v...))
}

func (l *JLogger) Warning(v ...interface{}) {
	l.print(level_warning, fmt.Sprint(v...))
}

func (l *JLogger) Error(v ...interface{}) {
	l.print(level_error, fmt.Sprint(v...))
}

func (l *JLogger) Fatal(v ...interface{}) {
	l.print(level_fatal, fmt.Sprint(v...))
}

func (l *JLogger) Panic(v ...interface{}) {
	l.print(level_panic, fmt.Sprint(v...))
}

func (l *JLogger) printData(level uint, data JLData, msg string) {
	if level >= l.minlevel {
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

		if level == level_fatal {
			os.Exit(1)
		} else if level == level_panic {
			panic(msg)
		}
	}
}

func (l *JLogger) print(level uint, msg string) {
	if level >= l.minlevel {
		if l.ansiMode {
			l.log.Output(2, fmt.Sprintf("\x1b[%dm[%s]\x1b[0m %s", levelToColor(level), levelToString(level), msg))
		} else {
			l.log.Output(2, fmt.Sprintf("[%s] %s", levelToString(level), msg))
		}

		if level == level_fatal {
			os.Exit(1)
		} else if level == level_panic {
			panic(msg)
		}
	}
}