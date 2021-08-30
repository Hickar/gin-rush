package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

var _logger *Logger

type Logger struct {
	logger     *log.Logger
	flags      []string
	logFormat  string
	timeFormat string
}

var (
	logger     *log.Logger
	flags      = []string{"INFO", "WARNING", "ERROR"}
	logFormat  string
	timeFormat string
)

type Level int

const (
	INFO Level = iota
	WARNING
	ERROR
)

func NewLogger(logFilePath string, logFormat string, timeFormat string) (*Logger, error) {
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return nil, err
	}

	logger = log.New(logFile, "", 0)
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
	l.logger.Printf(logFormat, time.Now().Format(timeFormat), message)
}

func (l *Logger) Warning(message string) {
	l.setLogPrefix(WARNING)
	l.logger.Printf(logFormat, time.Now().Format(timeFormat), message)
}

func (l *Logger) Error(message string) {
	l.setLogPrefix(ERROR)
	l.logger.Printf(logFormat, time.Now().Format(timeFormat), message)
}

func (l Logger) setLogPrefix(level Level) {
	prefix := flags[level]
	l.logger.SetPrefix(fmt.Sprintf("[%s]", prefix))
}