package firegorm

import (
	"log"
	"os"
	"strings"
)

// LogLevel represents the severity level of the logger.
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// logger is the global logger instance.
var logger *log.Logger
var logLevel LogLevel

// SetLogLevel configures the logging level for Firegorm.
func SetLogLevel(level string) {
	level = strings.ToUpper(level)
	switch level {
	case "DEBUG":
		logLevel = DEBUG
	case "INFO":
		logLevel = INFO
	case "WARN":
		logLevel = WARN
	case "ERROR":
		logLevel = ERROR
	default:
		logLevel = INFO // Default level
	}
}

// Log logs messages based on the current logging level.
func Log(level LogLevel, format string, v ...interface{}) {
	if level >= logLevel {
		logger.Printf(format, v...)
	}
}

// InitializeLogger sets up the logger.
func InitializeLogger() {
	logger = log.New(os.Stdout, "[Firegorm] ", log.LstdFlags)
	SetLogLevel(os.Getenv("FIREGORM_LOG_LEVEL"))
}
