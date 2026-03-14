package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var (
	currentLevel  = INFO
	logger        *log.Logger
	fileLogger    *log.Logger
	logFile       *os.File
	enableFileLog = false
)

func Init(level LogLevel, logDir string) error {
	currentLevel = level

	logger = log.New(os.Stdout, "", 0)

	if logDir != "" {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		logPath := filepath.Join(logDir, "launcher.log")
		var err error
		logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			logger.Printf("Warning: failed to open log file: %v (continuing without file logging)", err)
			return nil
		}
		fileLogger = log.New(logFile, "", log.LstdFlags)
		enableFileLog = true
	}

	return nil
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

func SetLevel(level LogLevel) {
	currentLevel = level
}

func logMessage(level LogLevel, format string, args ...interface{}) {
	if level < currentLevel {
		return
	}

	var prefix string
	switch level {
	case DEBUG:
		prefix = "\033[36m[DEBUG]\033[0m "
	case INFO:
		prefix = "\033[32m[INFO]\033[0m "
	case WARN:
		prefix = "\033[33m[WARN]\033[0m "
	case ERROR:
		prefix = "\033[31m[ERROR]\033[0m "
	}

	message := fmt.Sprintf(format, args...)
	logger.Printf("%s%s\n", prefix, message)

	if enableFileLog && fileLogger != nil {
		fileLogger.Printf("[%s] %s\n", levelToString(level), message)
	}
}

func levelToString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func Debug(format string, args ...interface{}) {
	logMessage(DEBUG, format, args...)
}

func Info(format string, args ...interface{}) {
	logMessage(INFO, format, args...)
}

func Warn(format string, args ...interface{}) {
	logMessage(WARN, format, args...)
}

func Error(format string, args ...interface{}) {
	logMessage(ERROR, format, args...)
}

func Fatal(format string, args ...interface{}) {
	logMessage(ERROR, format, args...)
	os.Exit(1)
}
