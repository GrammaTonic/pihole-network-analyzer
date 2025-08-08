package ssh

import (
	"errors"
	"fmt"
	"net"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
)

// ErrorType represents different categories of SSH-related errors
type ErrorType int

const (
	ErrorTypeUnknown ErrorType = iota
	ErrorTypeNetwork
	ErrorTypeAuth
	ErrorTypeTimeout
	ErrorTypeFile
	ErrorTypeDatabase
	ErrorTypePermission
	ErrorTypeConnection
)

// SSHError provides detailed context for SSH operation failures
type SSHError struct {
	Type      ErrorType
	Operation string
	Host      string
	Port      int
	Err       error
	Retryable bool
	Context   map[string]interface{}
}

func (e *SSHError) Error() string {
	baseMsg := fmt.Sprintf("SSH %s failed for %s:%d: %v", e.Operation, e.Host, e.Port, e.Err)

	if len(e.Context) > 0 {
		baseMsg += " (context:"
		for k, v := range e.Context {
			baseMsg += fmt.Sprintf(" %s=%v", k, v)
		}
		baseMsg += ")"
	}

	return baseMsg
}

func (e *SSHError) Is(target error) bool {
	if t, ok := target.(*SSHError); ok {
		return e.Type == t.Type
	}
	return false
}

func (e *SSHError) Unwrap() error {
	return e.Err
}

// IsRetryable returns true if the error might be resolved by retrying
func (e *SSHError) IsRetryable() bool {
	return e.Retryable
}

// NewSSHError creates a new SSH error with automatic type classification
func NewSSHError(operation, host string, port int, err error) *SSHError {
	sshErr := &SSHError{
		Operation: operation,
		Host:      host,
		Port:      port,
		Err:       err,
		Context:   make(map[string]interface{}),
	}

	sshErr.Type, sshErr.Retryable = classifyError(err)
	return sshErr
}

// classifyError determines the error type and retry eligibility
func classifyError(err error) (ErrorType, bool) {
	if err == nil {
		return ErrorTypeUnknown, false
	}

	// First check error message for specific patterns (higher priority)
	errMsg := err.Error()
	switch {
	case contains(errMsg, "connection refused"):
		return ErrorTypeConnection, true
	case contains(errMsg, "no route to host"):
		return ErrorTypeNetwork, true
	case contains(errMsg, "network unreachable"):
		return ErrorTypeNetwork, true
	case contains(errMsg, "timeout"):
		return ErrorTypeTimeout, true
	case contains(errMsg, "permission denied"):
		return ErrorTypePermission, false
	case contains(errMsg, "authentication"):
		return ErrorTypeAuth, false
	case contains(errMsg, "no such file"):
		return ErrorTypeFile, false
	case contains(errMsg, "database"):
		return ErrorTypeDatabase, false
	}

	// Check for connection refused via OpError
	if opErr, ok := err.(*net.OpError); ok {
		if sysErr, ok := opErr.Err.(*syscall.Errno); ok {
			if *sysErr == syscall.ECONNREFUSED {
				return ErrorTypeConnection, true
			}
		}
	}

	// Check for SSH-specific errors
	switch {
	case errors.Is(err, ssh.ErrNoAuth):
		return ErrorTypeAuth, false
	}

	// Check error message for SSH auth patterns
	if contains(errMsg, "ssh: handshake failed") || contains(errMsg, "authentication failed") {
		return ErrorTypeAuth, false
	}

	// Check for timeout and temporary errors (lower priority)
	if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() {
			return ErrorTypeTimeout, true
		}
		if netErr.Temporary() {
			return ErrorTypeNetwork, true
		}
	}

	return ErrorTypeUnknown, false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsHelper(s, substr))))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// RetryConfig defines retry behavior for SSH operations
type RetryConfig struct {
	MaxAttempts   int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
	Timeout       time.Duration
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		Timeout:       5 * time.Minute,
	}
}

// CalculateDelay calculates the delay for a given attempt number
func (rc *RetryConfig) CalculateDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	delay := time.Duration(float64(rc.InitialDelay) *
		powFloat64(rc.BackoffFactor, float64(attempt-1)))

	if delay > rc.MaxDelay {
		return rc.MaxDelay
	}

	return delay
}

// powFloat64 calculates base^exp for float64
func powFloat64(base, exp float64) float64 {
	if exp == 0 {
		return 1
	}
	result := base
	for i := 1; i < int(exp); i++ {
		result *= base
	}
	return result
}
