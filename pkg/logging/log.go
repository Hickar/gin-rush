package logging

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	logger *log.Logger
	flags = []string{"INFO", "WARNING", "ERROR"}
	logFormat string
	timeFormat string
)

type Level int

const (
	INFO Level = iota
	WARNING
	ERROR
)

func Setup(logFilePath string, lFormat string, tFormat string) error {
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		//logFile, err = os.Create(logFilePath)
		//if err != nil {
		//	return err
		//}
		return err
	}

	logger = log.New(logFile, "", 0)
	logger.SetFlags(0)

	logFormat = lFormat
	timeFormat = tFormat

	return nil
}

func Info(message string) {
	setLogPrefix(INFO)
	logger.Printf(logFormat, time.Now().Format(timeFormat), message)
}

func Warning(message string) {
	setLogPrefix(WARNING)
	logger.Printf(logFormat, time.Now().Format(timeFormat), message)
}

func Error(message string) {
	setLogPrefix(ERROR)
	logger.Printf(logFormat, time.Now().Format(timeFormat), message)
}

func setLogPrefix(level Level) {
	prefix := flags[level]
	logger.SetPrefix(fmt.Sprintf("[%s]", prefix))
}