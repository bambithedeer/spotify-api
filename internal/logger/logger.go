package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

type Level int

const (
	TraceLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

var levelNames = map[Level]string{
	TraceLevel: "TRACE",
	DebugLevel: "DEBUG",
	InfoLevel:  "INFO",
	WarnLevel:  "WARN",
	ErrorLevel: "ERROR",
	FatalLevel: "FATAL",
	PanicLevel: "PANIC",
}

type Logger struct {
	level  Level
	format string
	logger *log.Logger
}

type Fields map[string]interface{}

// New creates a new logger with the specified configuration
func New(level, format, output string) (*Logger, error) {
	l := &Logger{
		format: strings.ToLower(format),
	}

	// Parse level
	switch strings.ToLower(level) {
	case "trace":
		l.level = TraceLevel
	case "debug":
		l.level = DebugLevel
	case "info":
		l.level = InfoLevel
	case "warn":
		l.level = WarnLevel
	case "error":
		l.level = ErrorLevel
	case "fatal":
		l.level = FatalLevel
	case "panic":
		l.level = PanicLevel
	default:
		return nil, fmt.Errorf("invalid log level: %s", level)
	}

	// Set output
	var writer io.Writer
	switch strings.ToLower(output) {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		// Treat as file path
		file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		writer = file
	}

	l.logger = log.New(writer, "", 0)
	return l, nil
}

// log writes a log entry with the specified level and message
func (l *Logger) log(level Level, msg string, fields Fields) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format(time.RFC3339)
	levelName := levelNames[level]

	if l.format == "json" {
		entry := map[string]interface{}{
			"timestamp": timestamp,
			"level":     levelName,
			"message":   msg,
		}

		// Add fields
		for k, v := range fields {
			entry[k] = v
		}

		data, _ := json.Marshal(entry)
		l.logger.Println(string(data))
	} else {
		// Text format
		var fieldStr string
		if len(fields) > 0 {
			var parts []string
			for k, v := range fields {
				parts = append(parts, fmt.Sprintf("%s=%v", k, v))
			}
			fieldStr = " " + strings.Join(parts, " ")
		}

		l.logger.Printf("%s [%s] %s%s", timestamp, levelName, msg, fieldStr)
	}

	// Handle fatal and panic levels
	if level == FatalLevel {
		os.Exit(1)
	} else if level == PanicLevel {
		panic(msg)
	}
}

func (l *Logger) Trace(msg string) {
	l.log(TraceLevel, msg, nil)
}

func (l *Logger) TraceWithFields(msg string, fields Fields) {
	l.log(TraceLevel, msg, fields)
}

func (l *Logger) Debug(msg string) {
	l.log(DebugLevel, msg, nil)
}

func (l *Logger) DebugWithFields(msg string, fields Fields) {
	l.log(DebugLevel, msg, fields)
}

func (l *Logger) Info(msg string) {
	l.log(InfoLevel, msg, nil)
}

func (l *Logger) InfoWithFields(msg string, fields Fields) {
	l.log(InfoLevel, msg, fields)
}

func (l *Logger) Warn(msg string) {
	l.log(WarnLevel, msg, nil)
}

func (l *Logger) WarnWithFields(msg string, fields Fields) {
	l.log(WarnLevel, msg, fields)
}

func (l *Logger) Error(msg string) {
	l.log(ErrorLevel, msg, nil)
}

func (l *Logger) ErrorWithFields(msg string, fields Fields) {
	l.log(ErrorLevel, msg, fields)
}

func (l *Logger) Fatal(msg string) {
	l.log(FatalLevel, msg, nil)
}

func (l *Logger) FatalWithFields(msg string, fields Fields) {
	l.log(FatalLevel, msg, fields)
}

func (l *Logger) Panic(msg string) {
	l.log(PanicLevel, msg, nil)
}

func (l *Logger) PanicWithFields(msg string, fields Fields) {
	l.log(PanicLevel, msg, fields)
}