package logger

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestLogger_BasicLogging(t *testing.T) {
	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create logger with colors disabled for testing
	config := &Config{
		Level:        LevelDebug,
		EnableColors: false,
		EnableEmojis: false,
	}
	logger := New(config)

	// Test different log levels
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warning message")
	logger.Error("Error message")

	// Close writer and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	buf := new(bytes.Buffer)
	io.Copy(buf, r)
	output := buf.String()

	// Verify log messages are present
	if !strings.Contains(output, "Debug message") {
		t.Error("Debug message not found in output")
	}
	if !strings.Contains(output, "Info message") {
		t.Error("Info message not found in output")
	}
	if !strings.Contains(output, "Warning message") {
		t.Error("Warning message not found in output")
	}
	if !strings.Contains(output, "Error message") {
		t.Error("Error message not found in output")
	}
}

func TestLogger_WithComponent(t *testing.T) {
	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	config := &Config{
		Level:        LevelInfo,
		EnableColors: false,
		EnableEmojis: false,
	}
	logger := New(config)
	componentLogger := logger.Component("SSH")

	componentLogger.Info("Connection established")

	// Close writer and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	buf := new(bytes.Buffer)
	io.Copy(buf, r)
	output := buf.String()

	// Verify component prefix is present
	if !strings.Contains(output, "[SSH]") {
		t.Error("Component prefix not found in output")
	}
	if !strings.Contains(output, "Connection established") {
		t.Error("Log message not found in output")
	}
}

func TestLogger_WithFields(t *testing.T) {
	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	config := &Config{
		Level:        LevelInfo,
		EnableColors: false,
		EnableEmojis: false,
	}
	logger := New(config)
	fieldsLogger := logger.WithFields(map[string]any{
		"host": "192.168.1.100",
		"port": 22,
	})

	fieldsLogger.Info("Connection attempt")

	// Close writer and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	buf := new(bytes.Buffer)
	io.Copy(buf, r)
	output := buf.String()

	// Verify fields are present
	if !strings.Contains(output, "host=192.168.1.100") {
		t.Error("Host field not found in output")
	}
	if !strings.Contains(output, "port=22") {
		t.Error("Port field not found in output")
	}
}

func TestLogger_SpecialMethods(t *testing.T) {
	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	config := &Config{
		Level:        LevelInfo,
		EnableColors: false,
		EnableEmojis: true, // Enable emojis for this test
	}
	logger := New(config)

	logger.Success("Operation completed")
	logger.Progress("Processing data")

	// Close writer and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	buf := new(bytes.Buffer)
	io.Copy(buf, r)
	output := buf.String()

	// Verify special log types
	if !strings.Contains(output, "✅") {
		t.Error("Success emoji not found in output")
	}
	if !strings.Contains(output, "🔄") {
		t.Error("Progress emoji not found in output")
	}
}

func TestGlobalLogger(t *testing.T) {
	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test global logger functions
	Info("Global info message")
	Warn("Global warning message")

	// Close writer and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	buf := new(bytes.Buffer)
	io.Copy(buf, r)
	output := buf.String()

	// Verify global logger messages
	if !strings.Contains(output, "Global info message") {
		t.Error("Global info message not found in output")
	}
	if !strings.Contains(output, "Global warning message") {
		t.Error("Global warning message not found in output")
	}
}

func TestLogLevel_Filtering(t *testing.T) {
	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create logger with INFO level (should filter out DEBUG)
	config := &Config{
		Level:        LevelInfo,
		EnableColors: false,
		EnableEmojis: false,
	}
	logger := New(config)

	logger.Debug("This should not appear")
	logger.Info("This should appear")

	// Close writer and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	buf := new(bytes.Buffer)
	io.Copy(buf, r)
	output := buf.String()

	// Verify filtering works
	if strings.Contains(output, "This should not appear") {
		t.Error("Debug message should have been filtered out")
	}
	if !strings.Contains(output, "This should appear") {
		t.Error("Info message should have appeared")
	}
}

// Benchmark tests
func BenchmarkLogger_Info(b *testing.B) {
	config := &Config{
		Level:        LevelInfo,
		EnableColors: false,
		EnableEmojis: false,
	}
	logger := New(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark message %d", i)
	}
}

func BenchmarkLogger_WithFields(b *testing.B) {
	config := &Config{
		Level:        LevelInfo,
		EnableColors: false,
		EnableEmojis: false,
	}
	logger := New(config).WithFields(map[string]any{
		"host": "192.168.1.100",
		"port": 22,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark message %d", i)
	}
}

func BenchmarkGlobalLogger(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("Global benchmark message %d", i)
	}
}
