package logger

import (
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

type LogLevel slog.Level

const (
	DEBUG = LogLevel(slog.LevelDebug)
	INFO  = LogLevel(slog.LevelInfo)
	WARN  = LogLevel(slog.LevelWarn)
	ERROR = LogLevel(slog.LevelError)
)

var std *Logger

func init() {
	std = New()
}

func New() *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return &Logger{
		Logger: slog.New(handler),
	}
}

func (l *Logger) Debug(message string, data ...interface{}) {
	if len(data) > 0 {
		attrs := l.convertToAttrs(data[0])
		l.Logger.Debug(message, slog.GroupAttrs("data", attrs...))
	} else {
		l.Logger.Debug(message)
	}
}

func (l *Logger) Info(message string, data ...interface{}) {
	if len(data) > 0 {
		attrs := l.convertToAttrs(data[0])
		l.Logger.Info(message, slog.GroupAttrs("data", attrs...))
	} else {
		l.Logger.Info(message)
	}
}

func (l *Logger) Warn(message string, data ...interface{}) {
	if len(data) > 0 {
		attrs := l.convertToAttrs(data[0])
		l.Logger.Warn(message, slog.GroupAttrs("data", attrs...))
	} else {
		l.Logger.Warn(message)
	}
}

func (l *Logger) Error(message string, data ...interface{}) {
	if len(data) > 0 {
		attrs := l.convertToAttrs(data[0])
		l.Logger.Error(message, slog.GroupAttrs("data", attrs...))
	} else {
		l.Logger.Error(message)
	}
}

func (l *Logger) SetLevel(level LogLevel) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.Level(level),
	})
	l.Logger = slog.New(handler)
}

// convertToAttrs converts arbitrary data to slog attributes
func (l *Logger) convertToAttrs(data interface{}) []slog.Attr {
	switch v := data.(type) {
	case map[string]interface{}:
		var attrs []slog.Attr
		for key, value := range v {
			attrs = append(attrs, slog.Any(key, value))
		}
		return attrs
	default:
		return []slog.Attr{slog.Any("value", v)}
	}
}

// Package level functions
func Debug(message string, data ...interface{}) {
	std.Debug(message, data...)
}

func Info(message string, data ...interface{}) {
	std.Info(message, data...)
}

func Warn(message string, data ...interface{}) {
	std.Warn(message, data...)
}

func Error(message string, data ...interface{}) {
	std.Error(message, data...)
}

func SetLevel(level LogLevel) {
	std.SetLevel(level)
}