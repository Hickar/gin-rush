package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	_logger *Logger
	flags      = []string{"INFO", "WARNING", "ERROR"}
)

type Level int

const (
	INFO Level = iota
	WARNING
	ERROR
)

type Logger struct {
	logger     *log.Logger
	flags      []string
	logFormat  string
	timeFormat string
}

func NewLogger(logFilePath string, logFormat string, timeFormat string) (*Logger, error) {
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return nil, err
	}

	logger := log.New(logFile, "", 0)
	logger.SetFlags(0)

	_logger = &Logger{
		logger:     logger,
		flags:      []string{"INFO", "WARNING", "ERROR"},
		logFormat:  logFormat,
		timeFormat: timeFormat,
	}

	return _logger, nil
}

func GetLogger() *Logger {
	return _logger
}

func (l *Logger) Info(message string) {
	l.setLogPrefix(INFO)
	l.logger.Printf(l.logFormat, time.Now().Format(l.timeFormat), message)
}

func (l *Logger) Warning(message string) {
	l.setLogPrefix(WARNING)
	l.logger.Printf(l.logFormat, time.Now().Format(l.timeFormat), message)
}

func (l *Logger) Error(message error) {
	l.setLogPrefix(ERROR)
	l.logger.Printf(l.logFormat, time.Now().Format(l.timeFormat), message.Error())
}

func (l Logger) setLogPrefix(level Level) {
	prefix := flags[level]
	l.logger.SetPrefix(fmt.Sprintf("[%s] ", prefix))
}