package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"pihole-analyzer/internal/colors"
)

// LogLevel represents different log levels
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

// Logger provides structured logging with color support
type Logger struct {
	slogger *slog.Logger
	config  *Config
}

// Config holds logger configuration
type Config struct {
	Level         LogLevel
	EnableColors  bool
	EnableEmojis  bool
	OutputFile    string
	ShowTimestamp bool
	ShowCaller    bool
	Component     string
}

// DefaultConfig returns a default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:         LevelInfo,
		EnableColors:  true,
		EnableEmojis:  true,
		ShowTimestamp: true,
		ShowCaller:    false,
	}
}

// New creates a new structured logger with the given configuration
func New(config *Config) *Logger {
	if config == nil {
		config = DefaultConfig()
	}

	var writer io.Writer = os.Stdout
	if config.OutputFile != "" {
		// Ensure log directory exists
		if err := os.MkdirAll(filepath.Dir(config.OutputFile), 0755); err == nil {
			if file, err := os.OpenFile(config.OutputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
				writer = io.MultiWriter(os.Stdout, file)
			}
		}
	}

	opts := &slog.HandlerOptions{
		Level: slogLevelFromString(string(config.Level)),
	}

	var handler slog.Handler
	if config.EnableColors && isTerminal() {
		handler = newColorHandler(writer, opts, config)
	} else {
		handler = slog.NewTextHandler(writer, opts)
	}

	return &Logger{
		slogger: slog.New(handler),
		config:  config,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...any) {
	l.log(LevelDebug, "🔍", msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...any) {
	l.log(LevelInfo, "ℹ️", msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...any) {
	l.log(LevelWarn, "⚠️", msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...any) {
	l.log(LevelError, "❌", msg, args...)
}

// Success logs a success message (info level with success styling)
func (l *Logger) Success(msg string, args ...any) {
	l.log(LevelInfo, "✅", msg, args...)
}

// Progress logs a progress message (info level with progress styling)
func (l *Logger) Progress(msg string, args ...any) {
	l.log(LevelInfo, "🔄", msg, args...)
}

// Component creates a new logger with a component prefix
func (l *Logger) Component(name string) *Logger {
	newConfig := *l.config
	newConfig.Component = name
	return &Logger{
		slogger: l.slogger.With("component", name),
		config:  &newConfig,
	}
}

// WithFields creates a new logger with additional fields
func (l *Logger) WithFields(fields map[string]any) *Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{
		slogger: l.slogger.With(args...),
		config:  l.config,
	}
}

// InfoFields logs an info message with structured fields
func (l *Logger) InfoFields(msg string, fields map[string]any) {
	l.WithFields(fields).slogger.Info(msg)
}

// DebugFields logs a debug message with structured fields
func (l *Logger) DebugFields(msg string, fields map[string]any) {
	l.WithFields(fields).slogger.Debug(msg)
}

// ErrorFields logs an error message with structured fields
func (l *Logger) ErrorFields(msg string, fields map[string]any) {
	l.WithFields(fields).slogger.Error(msg)
}

// WarnFields logs a warning message with structured fields
func (l *Logger) WarnFields(msg string, fields map[string]any) {
	l.WithFields(fields).slogger.Warn(msg)
}

// GetSlogger returns the underlying slog.Logger
func (l *Logger) GetSlogger() *slog.Logger {
	return l.slogger
}

// log performs the actual logging with emoji and color support
func (l *Logger) log(level LogLevel, emoji string, msg string, args ...any) {
	// Format the message
	formattedMsg := msg
	if len(args) > 0 {
		formattedMsg = fmt.Sprintf(msg, args...)
	}

	// Add emoji if enabled
	if l.config.EnableEmojis {
		formattedMsg = emoji + " " + formattedMsg
	}

	// Add component prefix if set
	if l.config.Component != "" {
		formattedMsg = fmt.Sprintf("[%s] %s", l.config.Component, formattedMsg)
	}

	// Log using slog
	switch level {
	case LevelDebug:
		l.slogger.Debug(formattedMsg)
	case LevelInfo:
		l.slogger.Info(formattedMsg)
	case LevelWarn:
		l.slogger.Warn(formattedMsg)
	case LevelError:
		l.slogger.Error(formattedMsg)
	}
}

// colorHandler implements a custom slog handler with color support
type colorHandler struct {
	*slog.TextHandler
	config *Config
}

func newColorHandler(w io.Writer, opts *slog.HandlerOptions, config *Config) *colorHandler {
	return &colorHandler{
		TextHandler: slog.NewTextHandler(w, opts),
		config:      config,
	}
}

func (h *colorHandler) Handle(ctx context.Context, r slog.Record) error {
	if !h.config.EnableColors {
		return h.TextHandler.Handle(ctx, r)
	}

	// Build colored output
	var builder strings.Builder

	// Timestamp
	if h.config.ShowTimestamp {
		timestamp := r.Time.Format("15:04:05")
		builder.WriteString(colors.Gray(timestamp))
		builder.WriteString(" ")
	}

	// Level with color
	level := r.Level.String()
	switch r.Level {
	case slog.LevelDebug:
		builder.WriteString(colors.Gray("[" + level + "]"))
	case slog.LevelInfo:
		builder.WriteString(colors.Info("[" + level + "]"))
	case slog.LevelWarn:
		builder.WriteString(colors.Warning("[" + level + "]"))
	case slog.LevelError:
		builder.WriteString(colors.Error("[" + level + "]"))
	default:
		builder.WriteString("[" + level + "]")
	}
	builder.WriteString(" ")

	// Message
	builder.WriteString(r.Message)

	// Attributes
	if r.NumAttrs() > 0 {
		builder.WriteString(" ")
		r.Attrs(func(a slog.Attr) bool {
			builder.WriteString(colors.Gray(fmt.Sprintf("%s=%v", a.Key, a.Value)))
			builder.WriteString(" ")
			return true
		})
	}

	builder.WriteString("\n")
	fmt.Print(builder.String())
	return nil
}

// Helper functions

func slogLevelFromString(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func isTerminal() bool {
	if os.Getenv("TERM") == "" {
		return false
	}
	return true
}

// Global logger instance
var defaultLogger *Logger

// Initialize the default logger
func init() {
	defaultLogger = New(DefaultConfig())
}

// Package-level functions for convenience

// Debug logs a debug message using the default logger
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Info logs an info message using the default logger
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Warn logs a warning message using the default logger
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Error logs an error message using the default logger
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// Success logs a success message using the default logger
func Success(msg string, args ...any) {
	defaultLogger.Success(msg, args...)
}

// Progress logs a progress message using the default logger
func Progress(msg string, args ...any) {
	defaultLogger.Progress(msg, args...)
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	defaultLogger = logger
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	return defaultLogger
}

// Component creates a component logger using the default logger
func Component(name string) *Logger {
	return defaultLogger.Component(name)
}

// InfoFields logs an info message with structured fields using the default logger
func InfoFields(msg string, fields map[string]any) {
	defaultLogger.InfoFields(msg, fields)
}

// DebugFields logs a debug message with structured fields using the default logger
func DebugFields(msg string, fields map[string]any) {
	defaultLogger.DebugFields(msg, fields)
}

// ErrorFields logs an error message with structured fields using the default logger
func ErrorFields(msg string, fields map[string]any) {
	defaultLogger.ErrorFields(msg, fields)
}
