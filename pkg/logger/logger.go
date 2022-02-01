package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	flags   = []string{"INFO", "WARNING", "ERROR"}
)

type Level int

const (
	INFO Level = iota
	WARNING
	ERROR
)

type Logger interface {
	Info(string)
	Warning(string)
	Error(error)
}

type logger struct {
	logger     *log.Logger
	flags      []string
	logFormat  string
	timeFormat string
}

func NewLogger(logFilePath string, logFormat string, timeFormat string) (Logger, error) {
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return nil, err
	}

	newLogger := log.New(logFile, "", 0)
	newLogger.SetFlags(0)

	logger := &logger{
		logger:     newLogger,
		flags:      []string{"INFO", "WARNING", "ERROR"},
		logFormat:  logFormat,
		timeFormat: timeFormat,
	}

	return logger, nil
}

func (l *logger) Info(message string) {
	l.setLogPrefix(INFO)
	l.logger.Printf(l.logFormat, time.Now().Format(l.timeFormat), message)
}

func (l *logger) Warning(message string) {
	l.setLogPrefix(WARNING)
	l.logger.Printf(l.logFormat, time.Now().Format(l.timeFormat), message)
}

func (l *logger) Error(message error) {
	l.setLogPrefix(ERROR)
	l.logger.Printf(l.logFormat, time.Now().Format(l.timeFormat), message.Error())
}

func (l logger) setLogPrefix(level Level) {
	prefix := flags[level]
	l.logger.SetPrefix(fmt.Sprintf("[%s] ", prefix))
}
