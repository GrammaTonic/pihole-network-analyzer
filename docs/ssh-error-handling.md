# SSH Error Handling Improvements

This document describes the enhanced SSH error handling implementation for the Pi-hole Network Analyzer.

## Overview

The SSH operations have been significantly improved with comprehensive error handling, retry logic, and better user feedback. The improvements focus on reliability, robustness, and user experience.

## Key Features

### 1. Enhanced Error Classification

Errors are now classified into specific types for better handling:

- **ErrorTypeNetwork**: Network connectivity issues (retryable)
- **ErrorTypeAuth**: Authentication failures (not retryable)
- **ErrorTypeTimeout**: Operation timeouts (retryable)
- **ErrorTypeFile**: File system errors (not retryable)
- **ErrorTypeDatabase**: Database access errors (context-dependent)
- **ErrorTypePermission**: Permission issues (not retryable)
- **ErrorTypeConnection**: Connection establishment failures (retryable)

### 2. Retry Logic with Exponential Backoff

```go
type RetryConfig struct {
    MaxAttempts   int           // Maximum number of retry attempts
    InitialDelay  time.Duration // Initial delay between retries
    MaxDelay      time.Duration // Maximum delay cap
    BackoffFactor float64       // Multiplier for exponential backoff
    Timeout       time.Duration // Overall operation timeout
}
```

**Default Configuration:**
- Max Attempts: 3
- Initial Delay: 1 second
- Max Delay: 30 seconds
- Backoff Factor: 2.0 (doubles each attempt)
- Timeout: 5 minutes

### 3. Connection Manager

The `ConnectionManager` provides centralized SSH connection handling:

```go
// Create connection manager with default settings
connMgr := NewConnectionManager(config)

// Or with custom options
connMgr := NewConnectionManager(config,
    WithRetryConfig(customRetryConfig),
    WithLogger(customLogger),
)

// Connect with context support
ctx := context.WithTimeout(context.Background(), 2*time.Minute)
client, err := connMgr.Connect(ctx)
```

**Features:**
- Automatic retry on transient failures
- Support for multiple authentication methods
- Connection health testing
- Context-aware cancellation
- Comprehensive logging

### 4. Enhanced File Transfer

The `FileTransfer` component provides robust file operations:

```go
// Create file transfer handler
ft := NewFileTransfer(client, logger)

// Download with options
options := &TransferOptions{
    CreateDirs:         true,
    Overwrite:          true,
    BufferSize:         32 * 1024,
    ProgressCallback:   func(transferred, total int64) { /* track progress */ },
}

err := ft.DownloadFile(ctx, remotePath, localPath, options)
```

**Features:**
- Progress tracking and logging
- Retry logic for failed transfers
- File integrity verification
- Context-aware cancellation
- Automatic cleanup on failure

### 5. Authentication Methods

Enhanced authentication with better error handling:

1. **SSH Agent**: Automatic detection and key enumeration
2. **Key Files**: Improved key loading with detailed error messages
3. **Password**: Fallback authentication method
4. **Host Key Verification**: Configurable (currently insecure for development)

## Error Handling Examples

### Basic Error Handling

```go
client, err := connMgr.Connect(ctx)
if err != nil {
    if sshErr, ok := err.(*SSHError); ok {
        switch sshErr.Type {
        case ErrorTypeAuth:
            log.Error("Authentication failed - check credentials")
            return err
        case ErrorTypeNetwork:
            log.Warn("Network issue - may be temporary")
            // Could implement additional retry logic here
        case ErrorTypeTimeout:
            log.Warn("Connection timeout - try again later")
        default:
            log.Error("SSH connection failed: %v", sshErr)
        }
        
        if sshErr.IsRetryable() {
            log.Info("Error may be temporary, retry recommended")
        }
    }
    return err
}
```

### File Transfer Error Handling

```go
err := ft.DownloadFile(ctx, remotePath, localPath, options)
if err != nil {
    if sshErr, ok := err.(*SSHError); ok {
        switch sshErr.Type {
        case ErrorTypeFile:
            log.Error("Remote file not found or inaccessible: %s", remotePath)
        case ErrorTypePermission:
            log.Error("Permission denied accessing: %s", remotePath)
        case ErrorTypeNetwork:
            log.Warn("Network interruption during transfer")
        }
    }
    return err
}
```

## Configuration Examples

### Custom Retry Configuration

```go
customRetry := &RetryConfig{
    MaxAttempts:   5,                // Try up to 5 times
    InitialDelay:  2 * time.Second,  // Start with 2 second delay
    MaxDelay:      60 * time.Second, // Cap at 1 minute
    BackoffFactor: 1.5,              // 50% increase each attempt
    Timeout:       10 * time.Minute, // 10 minute overall timeout
}

connMgr := NewConnectionManager(config, WithRetryConfig(customRetry))
```

### Custom Logger

```go
type MyLogger struct {
    logger *log.Logger
}

func (l *MyLogger) Info(msg string, args ...interface{}) {
    l.logger.Printf("INFO: "+msg, args...)
}

func (l *MyLogger) Warn(msg string, args ...interface{}) {
    l.logger.Printf("WARN: "+msg, args...)
}

func (l *MyLogger) Error(msg string, args ...interface{}) {
    l.logger.Printf("ERROR: "+msg, args...)
}

customLogger := &MyLogger{logger: log.New(os.Stdout, "", log.LstdFlags)}
connMgr := NewConnectionManager(config, WithLogger(customLogger))
```

## Testing

The implementation includes comprehensive tests:

- Error classification validation
- Retry logic verification
- Configuration option testing
- Performance benchmarks
- Mock implementations for unit testing

Run tests with:

```bash
go test ./internal/ssh/...
go test -bench=. ./internal/ssh/...
```

## Migration Notes

### From Old Implementation

The old `connectSSH` and `downloadFile` functions are deprecated. Update existing code:

**Old:**
```go
client, err := connectSSH(config)
if err != nil {
    return fmt.Errorf("SSH connection failed: %v", err)
}
defer client.Close()

err = downloadFile(client, remotePath, localPath)
```

**New:**
```go
connMgr := NewConnectionManager(config)
client, err := connMgr.Connect(ctx)
if err != nil {
    return fmt.Errorf("SSH connection failed: %v", err)
}
defer client.Close()

ft := NewFileTransfer(client, &DefaultLogger{})
err = ft.DownloadFile(ctx, remotePath, localPath, DefaultTransferOptions())
```

## Performance Considerations

- Connection pooling is not implemented (single connection per operation)
- File transfers use 32KB buffer size by default (configurable)
- Progress logging occurs every 5 seconds to avoid spam
- Memory usage is optimized for large file transfers
- Context cancellation allows early termination

## Security Considerations

- Host key verification is currently disabled (development mode)
- Multiple authentication methods are attempted in order
- SSH agent integration provides secure key handling
- No credentials are logged or stored permanently
- Connection timeouts prevent hanging operations

## Future Improvements

1. **Host Key Verification**: Implement proper known_hosts handling
2. **Connection Pooling**: Reuse connections for multiple operations
3. **Compression**: Add support for SSH compression
4. **Multiplexing**: Support multiple concurrent file transfers
5. **Metrics**: Add detailed performance and reliability metrics
6. **Configuration**: More granular timeout and retry settings

## Troubleshooting

### Common Issues

1. **Authentication Failures**:
   - Check SSH key permissions (600 for private keys)
   - Verify SSH agent is running (`ssh-add -l`)
   - Test manual SSH connection first

2. **Network Timeouts**:
   - Check network connectivity
   - Verify firewall settings
   - Consider increasing timeout values

3. **File Transfer Failures**:
   - Check disk space on local system
   - Verify remote file permissions
   - Ensure remote file exists and is readable

### Debug Logging

Enable detailed logging by implementing a custom logger with debug level output:

```go
type DebugLogger struct{}

func (l *DebugLogger) Info(msg string, args ...interface{}) {
    log.Printf("SSH-INFO: "+msg, args...)
}

func (l *DebugLogger) Warn(msg string, args ...interface{}) {
    log.Printf("SSH-WARN: "+msg, args...)
}

func (l *DebugLogger) Error(msg string, args ...interface{}) {
    log.Printf("SSH-ERROR: "+msg, args...)
}
```

This will provide detailed information about connection attempts, authentication methods, and transfer progress.
