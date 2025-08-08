package ssh

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// ConnectionManager handles SSH connections with retry logic and error handling
type ConnectionManager struct {
	config      *types.PiholeConfig
	retryConfig *RetryConfig
	logger      Logger
}

// Logger interface for connection logging
type Logger interface {
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// DefaultLogger provides basic console logging
type DefaultLogger struct{}

func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	logger.Info(msg, args...)
}

func (l *DefaultLogger) Warn(msg string, args ...interface{}) {
	logger.Warn(msg, args...)
}

func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	logger.Error(msg, args...)
}

// NewConnectionManager creates a new SSH connection manager
func NewConnectionManager(config *types.PiholeConfig, options ...func(*ConnectionManager)) *ConnectionManager {
	cm := &ConnectionManager{
		config:      config,
		retryConfig: DefaultRetryConfig(),
		logger:      &DefaultLogger{},
	}

	for _, option := range options {
		option(cm)
	}

	return cm
}

// WithRetryConfig sets custom retry configuration
func WithRetryConfig(retryConfig *RetryConfig) func(*ConnectionManager) {
	return func(cm *ConnectionManager) {
		cm.retryConfig = retryConfig
	}
}

// WithLogger sets a custom logger
func WithLogger(logger Logger) func(*ConnectionManager) {
	return func(cm *ConnectionManager) {
		cm.logger = logger
	}
}

// Connect establishes an SSH connection with retry logic
func (cm *ConnectionManager) Connect(ctx context.Context) (*ssh.Client, error) {
	cm.logger.Info("Connecting to Pi-hole server %s:%d", cm.config.Host, cm.config.Port)

	var lastErr error
	startTime := time.Now()

	for attempt := 1; attempt <= cm.retryConfig.MaxAttempts; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Check overall timeout
		if time.Since(startTime) > cm.retryConfig.Timeout {
			return nil, NewSSHError("connect", cm.config.Host, cm.config.Port,
				fmt.Errorf("overall timeout exceeded after %v", cm.retryConfig.Timeout))
		}

		if attempt > 1 {
			delay := cm.retryConfig.CalculateDelay(attempt - 1)
			cm.logger.Warn("Connection attempt %d/%d failed, retrying in %v",
				attempt-1, cm.retryConfig.MaxAttempts, delay)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		cm.logger.Info("SSH connection attempt %d/%d", attempt, cm.retryConfig.MaxAttempts)

		client, err := cm.attemptConnection(ctx)
		if err == nil {
			cm.logger.Info("SSH connection established successfully")
			return client, nil
		}

		lastErr = err
		sshErr := NewSSHError("connect", cm.config.Host, cm.config.Port, err)

		// Log the error with context
		cm.logger.Error("Connection attempt %d failed: %v", attempt, sshErr)

		// Don't retry if the error is not retryable
		if !sshErr.IsRetryable() {
			cm.logger.Error("Error is not retryable, aborting")
			return nil, sshErr
		}

		// Don't retry on the last attempt
		if attempt == cm.retryConfig.MaxAttempts {
			break
		}
	}

	return nil, NewSSHError("connect", cm.config.Host, cm.config.Port,
		fmt.Errorf("all %d connection attempts failed, last error: %v",
			cm.retryConfig.MaxAttempts, lastErr))
}

// attemptConnection performs a single connection attempt
func (cm *ConnectionManager) attemptConnection(ctx context.Context) (*ssh.Client, error) {
	// Prepare authentication methods
	authMethods, err := cm.prepareAuthMethods()
	if err != nil {
		return nil, fmt.Errorf("failed to prepare authentication: %w", err)
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication methods available")
	}

	// Create SSH client configuration with timeout
	sshConfig := &ssh.ClientConfig{
		User:            cm.config.Username,
		Auth:            authMethods,
		HostKeyCallback: cm.getHostKeyCallback(),
		Timeout:         30 * time.Second,
		ClientVersion:   "SSH-2.0-PiHole-Analyzer",
	}

	// Create a cancellable dialer
	dialer := &net.Dialer{
		Timeout: 30 * time.Second,
	}

	// Dial with context support
	address := fmt.Sprintf("%s:%d", cm.config.Host, cm.config.Port)
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", address, err)
	}

	// Create SSH client connection
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, address, sshConfig)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("SSH handshake failed: %w", err)
	}

	return ssh.NewClient(sshConn, chans, reqs), nil
}

// prepareAuthMethods prepares all available authentication methods
func (cm *ConnectionManager) prepareAuthMethods() ([]ssh.AuthMethod, error) {
	var authMethods []ssh.AuthMethod
	var authErrors []error

	// Try SSH agent first
	if agentAuth, err := cm.trySSHAgent(); err != nil {
		authErrors = append(authErrors, fmt.Errorf("SSH agent: %w", err))
	} else if agentAuth != nil {
		authMethods = append(authMethods, agentAuth)
		cm.logger.Info("SSH agent authentication available")
	}

	// Try key-based authentication
	if cm.config.KeyFile != "" {
		if keyAuth, err := cm.tryKeyFile(); err != nil {
			authErrors = append(authErrors, fmt.Errorf("key file: %w", err))
		} else if keyAuth != nil {
			authMethods = append(authMethods, keyAuth)
			cm.logger.Info("Key file authentication available: %s", cm.config.KeyFile)
		}
	}

	// Try password authentication
	if cm.config.Password != "" {
		authMethods = append(authMethods, ssh.Password(cm.config.Password))
		cm.logger.Info("Password authentication available")
	}

	// Log authentication preparation results
	if len(authMethods) == 0 {
		cm.logger.Error("No authentication methods could be prepared")
		for _, err := range authErrors {
			cm.logger.Error("Auth preparation error: %v", err)
		}
		return nil, fmt.Errorf("no authentication methods available")
	}

	cm.logger.Info("Prepared %d authentication methods", len(authMethods))
	return authMethods, nil
}

// trySSHAgent attempts to connect to SSH agent
func (cm *ConnectionManager) trySSHAgent() (ssh.AuthMethod, error) {
	sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
	if sshAuthSock == "" {
		return nil, fmt.Errorf("SSH_AUTH_SOCK not set")
	}

	agentConn, err := net.Dial("unix", sshAuthSock)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH agent: %w", err)
	}

	agentClient := agent.NewClient(agentConn)
	signers, err := agentClient.Signers()
	if err != nil {
		agentConn.Close()
		return nil, fmt.Errorf("failed to get signers from SSH agent: %w", err)
	}

	if len(signers) == 0 {
		agentConn.Close()
		return nil, fmt.Errorf("no keys available in SSH agent")
	}

	// Note: We don't close agentConn here as it will be used by the auth method
	return ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
		return signers, nil
	}), nil
}

// tryKeyFile attempts to load SSH key from file
func (cm *ConnectionManager) tryKeyFile() (ssh.AuthMethod, error) {
	keyData, err := os.ReadFile(cm.config.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file %s: %w", cm.config.KeyFile, err)
	}

	signer, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return ssh.PublicKeys(signer), nil
}

// getHostKeyCallback returns appropriate host key callback
func (cm *ConnectionManager) getHostKeyCallback() ssh.HostKeyCallback {
	// In production, you would want to implement proper host key verification
	// For now, we'll use InsecureIgnoreHostKey but log a warning
	cm.logger.Warn("Using insecure host key verification - consider implementing proper host key checking")
	return ssh.InsecureIgnoreHostKey()
}

// TestConnection performs a simple connectivity test
func (cm *ConnectionManager) TestConnection(ctx context.Context) error {
	client, err := cm.Connect(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	// Test by running a simple command
	session, err := client.NewSession()
	if err != nil {
		return NewSSHError("test", cm.config.Host, cm.config.Port,
			fmt.Errorf("failed to create session: %w", err))
	}
	defer session.Close()

	// Run a simple echo command
	output, err := session.Output("echo 'connection_test'")
	if err != nil {
		return NewSSHError("test", cm.config.Host, cm.config.Port,
			fmt.Errorf("test command failed: %w", err))
	}

	expected := "connection_test\n"
	if string(output) != expected {
		return NewSSHError("test", cm.config.Host, cm.config.Port,
			fmt.Errorf("unexpected test output: got %q, expected %q", string(output), expected))
	}

	cm.logger.Info("Connection test successful")
	return nil
}
