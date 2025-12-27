package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger wraps logrus for consistent logging across the application
type Logger struct {
	log *logrus.Logger
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.InfoLevel)

	return &Logger{log: log}
}

// Info logs an informational message
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	if fields != nil {
		l.log.WithFields(logrus.Fields(fields)).Info(msg)
	} else {
		l.log.Info(msg)
	}
}

// Error logs an error message
func (l *Logger) Error(msg string, err error, fields map[string]interface{}) {
	logFields := logrus.Fields{}
	if fields != nil {
		logFields = logrus.Fields(fields)
	}
	if err != nil {
		logFields["error"] = err.Error()
	}

	if len(logFields) > 0 {
		l.log.WithFields(logFields).Error(msg)
	} else {
		l.log.Error(msg)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	if fields != nil {
		l.log.WithFields(logrus.Fields(fields)).Debug(msg)
	} else {
		l.log.Debug(msg)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	if fields != nil {
		l.log.WithFields(logrus.Fields(fields)).Warn(msg)
	} else {
		l.log.Warn(msg)
	}
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level string) {
	switch level {
	case "debug":
		l.log.SetLevel(logrus.DebugLevel)
	case "info":
		l.log.SetLevel(logrus.InfoLevel)
	case "warn":
		l.log.SetLevel(logrus.WarnLevel)
	case "error":
		l.log.SetLevel(logrus.ErrorLevel)
	default:
		l.log.SetLevel(logrus.InfoLevel)
	}
}
