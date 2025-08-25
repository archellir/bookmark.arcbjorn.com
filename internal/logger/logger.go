package logger

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Logger struct {
	*log.Logger
	level LogLevel
}

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

type LogEntry struct {
	Timestamp string      `json:"timestamp"`
	Level     string      `json:"level"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
}

var std *Logger

func init() {
	std = New()
}

func New() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", 0),
		level:  INFO,
	}
}

func (l *Logger) logJSON(level LogLevel, message string, data interface{}) {
	if level < l.level {
		return
	}

	levelNames := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     levelNames[level],
		Message:   message,
		Data:      data,
	}

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		l.Printf("Failed to marshal log entry: %v", err)
		return
	}

	l.Print(string(jsonBytes))
}

func (l *Logger) Debug(message string, data ...interface{}) {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	l.logJSON(DEBUG, message, d)
}

func (l *Logger) Info(message string, data ...interface{}) {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	l.logJSON(INFO, message, d)
}

func (l *Logger) Warn(message string, data ...interface{}) {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	l.logJSON(WARN, message, d)
}

func (l *Logger) Error(message string, data ...interface{}) {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	l.logJSON(ERROR, message, d)
}

func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
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