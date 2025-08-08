package ssh

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// FileTransfer handles secure file transfers with error recovery
type FileTransfer struct {
	client      *ssh.Client
	logger      Logger
	retryConfig *RetryConfig
}

// NewFileTransfer creates a new file transfer handler
func NewFileTransfer(client *ssh.Client, logger Logger) *FileTransfer {
	return &FileTransfer{
		client:      client,
		logger:      logger,
		retryConfig: DefaultRetryConfig(),
	}
}

// TransferOptions configures file transfer behavior
type TransferOptions struct {
	PreservePermissions bool
	CreateDirs          bool
	Overwrite           bool
	BufferSize          int
	ProgressCallback    func(transferred, total int64)
}

// DefaultTransferOptions returns sensible defaults
func DefaultTransferOptions() *TransferOptions {
	return &TransferOptions{
		PreservePermissions: false,
		CreateDirs:          true,
		Overwrite:           true,
		BufferSize:          32 * 1024, // 32KB buffer
	}
}

// DownloadFile downloads a file from remote server with retry logic
func (ft *FileTransfer) DownloadFile(ctx context.Context, remotePath, localPath string, options *TransferOptions) error {
	if options == nil {
		options = DefaultTransferOptions()
	}

	ft.logger.Info("Starting download: %s → %s", remotePath, localPath)

	// Create local directory if needed
	if options.CreateDirs {
		localDir := filepath.Dir(localPath)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return fmt.Errorf("failed to create local directory %s: %w", localDir, err)
		}
	}

	// Check if local file exists
	if !options.Overwrite {
		if _, err := os.Stat(localPath); err == nil {
			return fmt.Errorf("local file %s already exists and overwrite is disabled", localPath)
		}
	}

	var lastErr error
	for attempt := 1; attempt <= ft.retryConfig.MaxAttempts; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if attempt > 1 {
			delay := ft.retryConfig.CalculateDelay(attempt - 1)
			ft.logger.Warn("Download attempt %d/%d failed, retrying in %v",
				attempt-1, ft.retryConfig.MaxAttempts, delay)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		ft.logger.Info("Download attempt %d/%d", attempt, ft.retryConfig.MaxAttempts)

		err := ft.attemptDownload(ctx, remotePath, localPath, options)
		if err == nil {
			ft.logger.Info("Download completed successfully")
			return nil
		}

		lastErr = err
		ft.logger.Error("Download attempt %d failed: %v", attempt, err)

		// Check if error is retryable
		if sshErr, ok := err.(*SSHError); ok && !sshErr.IsRetryable() {
			ft.logger.Error("Error is not retryable, aborting")
			return sshErr
		}

		// Clean up partial file on error
		if _, statErr := os.Stat(localPath); statErr == nil {
			if removeErr := os.Remove(localPath); removeErr != nil {
				ft.logger.Warn("Failed to clean up partial file %s: %v", localPath, removeErr)
			}
		}
	}

	return fmt.Errorf("all %d download attempts failed, last error: %w",
		ft.retryConfig.MaxAttempts, lastErr)
}

// attemptDownload performs a single download attempt
func (ft *FileTransfer) attemptDownload(ctx context.Context, remotePath, localPath string, options *TransferOptions) error {
	// First, get remote file info
	fileInfo, err := ft.getRemoteFileInfo(remotePath)
	if err != nil {
		return NewSSHError("stat", "", 0, fmt.Errorf("failed to get remote file info: %w", err))
	}

	ft.logger.Info("Remote file size: %d bytes", fileInfo.Size)

	// Create local file
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file %s: %w", localPath, err)
	}
	defer func() {
		if closeErr := localFile.Close(); closeErr != nil {
			ft.logger.Warn("Failed to close local file: %v", closeErr)
		}
	}()

	// Use SCP protocol for reliable transfer
	return ft.scpDownload(ctx, remotePath, localFile, fileInfo, options)
}

// RemoteFileInfo contains information about a remote file
type RemoteFileInfo struct {
	Size        int64
	Mode        os.FileMode
	ModTime     time.Time
	IsDirectory bool
}

// getRemoteFileInfo retrieves information about a remote file
func (ft *FileTransfer) getRemoteFileInfo(remotePath string) (*RemoteFileInfo, error) {
	session, err := ft.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Use stat command to get file info
	cmd := fmt.Sprintf("stat -c '%%s %%f %%Y' %s 2>/dev/null || ls -ld %s",
		shellEscape(remotePath), shellEscape(remotePath))

	output, err := session.Output(cmd)
	if err != nil {
		return nil, fmt.Errorf("stat command failed: %w", err)
	}

	return ft.parseFileInfo(string(output))
}

// parseFileInfo parses the output of stat command
func (ft *FileTransfer) parseFileInfo(output string) (*RemoteFileInfo, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty stat output")
	}

	line := strings.TrimSpace(lines[0])

	// Try to parse stat output first (size mode timestamp)
	fields := strings.Fields(line)
	if len(fields) >= 3 {
		size, err1 := strconv.ParseInt(fields[0], 10, 64)
		mode, err2 := strconv.ParseUint(fields[1], 16, 32)
		timestamp, err3 := strconv.ParseInt(fields[2], 10, 64)

		if err1 == nil && err2 == nil && err3 == nil {
			return &RemoteFileInfo{
				Size:        size,
				Mode:        os.FileMode(mode),
				ModTime:     time.Unix(timestamp, 0),
				IsDirectory: os.FileMode(mode).IsDir(),
			}, nil
		}
	}

	// Fallback to parsing ls output
	if len(fields) >= 5 {
		sizeStr := fields[4]
		if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
			isDir := strings.HasPrefix(line, "d")
			return &RemoteFileInfo{
				Size:        size,
				Mode:        0644,
				ModTime:     time.Now(),
				IsDirectory: isDir,
			}, nil
		}
	}

	return nil, fmt.Errorf("failed to parse file info from: %s", output)
}

// scpDownload performs SCP-style download
func (ft *FileTransfer) scpDownload(ctx context.Context, remotePath string, localFile *os.File, fileInfo *RemoteFileInfo, options *TransferOptions) error {
	session, err := ft.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Create pipe for reading data
	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Start SCP command
	cmd := fmt.Sprintf("cat %s", shellEscape(remotePath))
	if err := session.Start(cmd); err != nil {
		return fmt.Errorf("failed to start SCP command: %w", err)
	}

	// Copy data with progress tracking
	return ft.copyWithProgress(ctx, stdout, localFile, fileInfo.Size, options.ProgressCallback)
}

// copyWithProgress copies data while tracking progress and handling cancellation
func (ft *FileTransfer) copyWithProgress(ctx context.Context, src io.Reader, dst io.Writer, totalSize int64, progressCallback func(int64, int64)) error {
	buffer := make([]byte, 32*1024) // 32KB buffer
	var transferred int64

	startTime := time.Now()
	lastProgress := time.Now()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := src.Read(buffer)
		if n > 0 {
			if _, writeErr := dst.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("write failed: %w", writeErr)
			}

			transferred += int64(n)

			// Call progress callback and log progress
			if progressCallback != nil {
				progressCallback(transferred, totalSize)
			}

			// Log progress every 5 seconds
			if time.Since(lastProgress) > 5*time.Second {
				ft.logProgress(transferred, totalSize, time.Since(startTime))
				lastProgress = time.Now()
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read failed: %w", err)
		}
	}

	// Final progress log
	ft.logProgress(transferred, totalSize, time.Since(startTime))

	// Verify transfer size
	if totalSize > 0 && transferred != totalSize {
		return fmt.Errorf("transfer size mismatch: expected %d bytes, got %d bytes", totalSize, transferred)
	}

	return nil
}

// logProgress logs transfer progress
func (ft *FileTransfer) logProgress(transferred, total int64, elapsed time.Duration) {
	if total > 0 {
		percent := float64(transferred) / float64(total) * 100
		rate := float64(transferred) / elapsed.Seconds() / 1024 // KB/s
		ft.logger.Info("Transfer progress: %.1f%% (%s/%s) at %.1f KB/s",
			percent, formatBytes(transferred), formatBytes(total), rate)
	} else {
		rate := float64(transferred) / elapsed.Seconds() / 1024 // KB/s
		ft.logger.Info("Transfer progress: %s at %.1f KB/s",
			formatBytes(transferred), rate)
	}
}

// formatBytes formats byte count as human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// shellEscape escapes a string for safe use in shell commands
func shellEscape(s string) string {
	// Simple escaping - wrap in single quotes and escape any single quotes
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

// VerifyFile verifies the integrity of a downloaded file
func (ft *FileTransfer) VerifyFile(localPath string, expectedSize int64) error {
	stat, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("failed to stat local file: %w", err)
	}

	if stat.Size() != expectedSize {
		return fmt.Errorf("file size mismatch: expected %d bytes, got %d bytes",
			expectedSize, stat.Size())
	}

	// Check if file is readable
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file for verification: %w", err)
	}
	defer file.Close()

	ft.logger.Info("File verification successful: %s (%s)", localPath, formatBytes(stat.Size()))
	return nil
}
