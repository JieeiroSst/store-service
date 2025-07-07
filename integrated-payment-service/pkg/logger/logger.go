package logger

import (
	"github.com/sirupsen/logrus"
)

type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

type logrusLogger struct {
	logger *logrus.Logger
}

func New(level string) Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	switch level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	return &logrusLogger{logger: logger}
}

func (l *logrusLogger) Info(msg string, fields ...interface{}) {
	l.logger.WithFields(l.parseFields(fields...)).Info(msg)
}

func (l *logrusLogger) Error(msg string, fields ...interface{}) {
	l.logger.WithFields(l.parseFields(fields...)).Error(msg)
}

func (l *logrusLogger) Debug(msg string, fields ...interface{}) {
	l.logger.WithFields(l.parseFields(fields...)).Debug(msg)
}

func (l *logrusLogger) Warn(msg string, fields ...interface{}) {
	l.logger.WithFields(l.parseFields(fields...)).Warn(msg)
}

func (l *logrusLogger) parseFields(fields ...interface{}) logrus.Fields {
	result := logrus.Fields{}
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key := fields[i].(string)
			value := fields[i+1]
			result[key] = value
		}
	}
	return result
}
