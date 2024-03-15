package logging

import (
	"fmt"
	"github.com/logrusorgru/aurora/v4"
	"time"
)

type LoggerLevel int

const (
	DEBUG LoggerLevel = 0
	INFO  LoggerLevel = 1
	WARN  LoggerLevel = 2
	ERROR LoggerLevel = 3
)

type Logger interface {
	Info(format string, a ...any)
	Error(format string, a ...any)
	Debug(format string, a ...any)
	Warn(format string, a ...any)
}

type SimpleLogger struct {
	Tag   string
	Level LoggerLevel
}

func NewLogger(tag string, level LoggerLevel) *SimpleLogger {
	return &SimpleLogger{
		Tag:   tag,
		Level: level,
	}
}

var timePattern = "2006-01-02 15:04:05.000"
var defaultPattern = "[%-23s] [%10s] [%-5s] - %s\n"

func (logger *SimpleLogger) Info(format string, a ...any) {
	if logger.Level <= INFO {
		msg := fmt.Sprintf(format, a...)
		fmt.Printf(defaultPattern, getLoggerTime(), getLoggerTag(logger.Tag), aurora.Green("INFO"), aurora.Green(msg))
	}
}

func (logger *SimpleLogger) Debug(format string, a ...any) {
	if logger.Level <= DEBUG {
		msg := fmt.Sprintf(format, a...)
		fmt.Printf(defaultPattern, getLoggerTime(), getLoggerTag(logger.Tag), aurora.Gray(10, "DEBUG"), aurora.Gray(10, msg))
	}
}
func (logger *SimpleLogger) Warn(format string, a ...any) {
	if logger.Level <= WARN {
		msg := fmt.Sprintf(format, a...)
		fmt.Printf(defaultPattern, getLoggerTime(), getLoggerTag(logger.Tag), aurora.Yellow("WARN"), aurora.Yellow(msg))
	}
}
func (logger *SimpleLogger) Error(format string, a ...any) {
	if logger.Level <= ERROR {
		msg := fmt.Sprintf(format, a...)
		fmt.Printf(defaultPattern, getLoggerTime(), getLoggerTag(logger.Tag), aurora.Red("Error"), aurora.Red(msg))
	}
}

func getLoggerTime() aurora.Value {
	return aurora.Cyan(time.Now().Format(timePattern))
}

func getLoggerTag(tag string) aurora.Value {
	return aurora.Cyan(tag)
}
