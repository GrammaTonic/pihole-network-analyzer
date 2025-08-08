package ssh

import (
	"testing"
	"time"

	"pihole-analyzer/internal/types"
)

func TestSSHErrorClassification(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantType  ErrorType
		retryable bool
	}{
		{
			name:      "connection refused",
			err:       &mockNetError{msg: "connection refused", temporary: true},
			wantType:  ErrorTypeConnection,
			retryable: true,
		},
		{
			name:      "timeout error",
			err:       &mockNetError{msg: "timeout", timeout: true},
			wantType:  ErrorTypeTimeout,
			retryable: true,
		},
		{
			name:      "permission denied",
			err:       &mockError{msg: "permission denied"},
			wantType:  ErrorTypePermission,
			retryable: false,
		},
		{
			name:      "no such file",
			err:       &mockError{msg: "no such file"},
			wantType:  ErrorTypeFile,
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sshErr := NewSSHError("test", "localhost", 22, tt.err)

			if sshErr.Type != tt.wantType {
				t.Errorf("Expected error type %v, got %v", tt.wantType, sshErr.Type)
			}

			if sshErr.IsRetryable() != tt.retryable {
				t.Errorf("Expected retryable %v, got %v", tt.retryable, sshErr.IsRetryable())
			}
		})
	}
}

func TestRetryConfigCalculateDelay(t *testing.T) {
	config := &RetryConfig{
		InitialDelay:  time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
	}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 0},
		{1, time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{4, 8 * time.Second},
		{10, 30 * time.Second}, // Should be capped at MaxDelay
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := config.CalculateDelay(tt.attempt)
			if got != tt.want {
				t.Errorf("CalculateDelay(%d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}
}

func TestConnectionManagerOptions(t *testing.T) {
	config := &types.PiholeConfig{
		Host:     "test.example.com",
		Port:     22,
		Username: "testuser",
	}

	customRetry := &RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 2 * time.Second,
	}

	logger := &mockLogger{}

	cm := NewConnectionManager(config,
		WithRetryConfig(customRetry),
		WithLogger(logger),
	)

	if cm.retryConfig.MaxAttempts != 5 {
		t.Errorf("Expected MaxAttempts 5, got %d", cm.retryConfig.MaxAttempts)
	}

	if cm.retryConfig.InitialDelay != 2*time.Second {
		t.Errorf("Expected InitialDelay 2s, got %v", cm.retryConfig.InitialDelay)
	}

	if cm.logger != logger {
		t.Error("Expected custom logger to be set")
	}
}

func TestFileTransferOptions(t *testing.T) {
	options := DefaultTransferOptions()

	if !options.CreateDirs {
		t.Error("Expected CreateDirs to be true by default")
	}

	if !options.Overwrite {
		t.Error("Expected Overwrite to be true by default")
	}

	if options.BufferSize != 32*1024 {
		t.Errorf("Expected BufferSize 32KB, got %d", options.BufferSize)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := formatBytes(tt.bytes)
			if got != tt.want {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestShellEscape(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"simple", "'simple'"},
		{"with spaces", "'with spaces'"},
		{"with'quote", "'with'\"'\"'quote'"},
		{"multiple'quotes'here", "'multiple'\"'\"'quotes'\"'\"'here'"},
		{"", "''"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := shellEscape(tt.input)
			if got != tt.want {
				t.Errorf("shellEscape(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// Mock implementations for testing

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

type mockNetError struct {
	msg       string
	timeout   bool
	temporary bool
}

func (e *mockNetError) Error() string {
	return e.msg
}

func (e *mockNetError) Timeout() bool {
	return e.timeout
}

func (e *mockNetError) Temporary() bool {
	return e.temporary
}

type mockLogger struct {
	logs []string
}

func (l *mockLogger) Info(msg string, args ...interface{}) {
	l.logs = append(l.logs, "INFO: "+msg)
}

func (l *mockLogger) Warn(msg string, args ...interface{}) {
	l.logs = append(l.logs, "WARN: "+msg)
}

func (l *mockLogger) Error(msg string, args ...interface{}) {
	l.logs = append(l.logs, "ERROR: "+msg)
}

// Benchmark tests for performance validation

func BenchmarkErrorClassification(b *testing.B) {
	err := &mockNetError{msg: "connection refused", temporary: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		classifyError(err)
	}
}

func BenchmarkRetryDelayCalculation(b *testing.B) {
	config := DefaultRetryConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.CalculateDelay(5)
	}
}

func BenchmarkShellEscape(b *testing.B) {
	input := "test'with'quotes and spaces"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shellEscape(input)
	}
}
