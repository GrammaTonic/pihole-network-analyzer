package loki

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"pihole-analyzer/internal/integrations/interfaces"
	"pihole-analyzer/internal/types"
)

// Client implements Loki integration for structured log shipping
type Client struct {
	config    *types.LokiConfig
	logger    *slog.Logger
	httpClient *http.Client
	buffer    []interfaces.LogEntry
	bufferMu  sync.Mutex
	batchTicker *time.Ticker
	stopChan  chan struct{}
	enabled   bool
	status    interfaces.IntegrationStatus
	mu        sync.RWMutex
}

// NewClient creates a new Loki integration client
func NewClient(config *types.LokiConfig, logger *slog.Logger) *Client {
	client := &Client{
		config:   config,
		logger:   logger,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		buffer:   make([]interfaces.LogEntry, 0, config.BufferSize),
		stopChan: make(chan struct{}),
		enabled:  config.Enabled,
		status: interfaces.IntegrationStatus{
			Name:    "loki",
			Enabled: config.Enabled,
		},
	}

	if config.Enabled {
		// Parse batch timeout
		batchTimeout, err := time.ParseDuration(config.BatchTimeout)
		if err != nil {
			logger.Warn("⚠️ Invalid batch timeout, using default 10s",
				slog.String("component", "loki"),
				slog.String("value", config.BatchTimeout))
			batchTimeout = 10 * time.Second
		}

		client.batchTicker = time.NewTicker(batchTimeout)
		go client.batchProcessor()
	}

	return client
}

// Initialize sets up the Loki integration
func (c *Client) Initialize(ctx context.Context, config interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if lokiConfig, ok := config.(*types.LokiConfig); ok {
		c.config = lokiConfig
		c.enabled = lokiConfig.Enabled
	}

	if !c.enabled {
		c.logger.Debug("⏭️ Loki integration disabled",
			slog.String("component", "loki"))
		return nil
	}

	c.logger.Info("🚀 Initializing Loki integration",
		slog.String("component", "loki"),
		slog.String("url", c.config.URL))

	// Test connection
	if err := c.TestConnection(ctx); err != nil {
		c.status.LastError = err.Error()
		return fmt.Errorf("failed to connect to Loki: %w", err)
	}

	c.status.Connected = true
	c.status.LastConnect = time.Now()

	c.logger.Info("✅ Loki integration initialized successfully",
		slog.String("component", "loki"))

	return nil
}

// IsEnabled returns whether the integration is enabled
func (c *Client) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.enabled
}

// GetName returns the name of the integration
func (c *Client) GetName() string {
	return "loki"
}

// GetStatus returns the current status of the integration
func (c *Client) GetStatus() interfaces.IntegrationStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

// SendMetrics sends metrics data (not implemented for Loki)
func (c *Client) SendMetrics(ctx context.Context, data *types.AnalysisResult) error {
	// Loki is for logs, not metrics - this is a no-op
	return nil
}

// SendLogs sends log data to Loki
func (c *Client) SendLogs(ctx context.Context, logs []interfaces.LogEntry) error {
	if !c.enabled {
		return nil
	}

	if len(logs) == 0 {
		return nil
	}

	c.logger.Debug("📝 Sending logs to Loki",
		slog.String("component", "loki"),
		slog.Int("log_count", len(logs)))

	// Convert logs to Loki format and send
	streams := c.convertToLokiStreams(logs)
	return c.pushStreams(ctx, streams)
}

// WriteLogs writes a batch of logs
func (c *Client) WriteLogs(ctx context.Context, logs []interfaces.LogEntry) error {
	if !c.enabled {
		return nil
	}

	// Add to buffer for batch processing
	c.bufferMu.Lock()
	defer c.bufferMu.Unlock()

	for _, log := range logs {
		if len(c.buffer) >= c.config.BufferSize {
			// Buffer is full, send immediately
			if err := c.flushBuffer(ctx); err != nil {
				c.logger.Error("❌ Failed to flush log buffer",
					slog.String("component", "loki"),
					slog.String("error", err.Error()))
				return err
			}
		}
		c.buffer = append(c.buffer, log)
	}

	return nil
}

// SetLogLevel sets the minimum log level to forward
func (c *Client) SetLogLevel(level slog.Level) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Store log level in status metadata
	if c.status.Metadata == nil {
		c.status.Metadata = make(map[string]string)
	}
	c.status.Metadata["min_log_level"] = level.String()
	
	c.logger.Debug("📊 Set minimum log level for Loki",
		slog.String("component", "loki"),
		slog.String("level", level.String()))
}

// AddStaticLabels adds static labels to all log entries
func (c *Client) AddStaticLabels(labels map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Merge with existing static labels
	if c.config.StaticLabels == nil {
		c.config.StaticLabels = make(map[string]string)
	}
	
	for k, v := range labels {
		c.config.StaticLabels[k] = v
	}
	
	c.logger.Debug("🏷️ Added static labels to Loki",
		slog.String("component", "loki"),
		slog.Int("label_count", len(labels)))
}

// StreamLogs streams logs to Loki
func (c *Client) StreamLogs(ctx context.Context, logStream <-chan interfaces.LogEntry) error {
	if !c.enabled {
		return nil
	}

	c.logger.Info("🔄 Starting log streaming to Loki",
		slog.String("component", "loki"))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case log, ok := <-logStream:
			if !ok {
				return nil
			}
			
			if err := c.WriteLogs(ctx, []interfaces.LogEntry{log}); err != nil {
				c.logger.Error("❌ Failed to write log to Loki stream",
					slog.String("component", "loki"),
					slog.String("error", err.Error()))
				return err
			}
		}
	}
}

// TestConnection tests the connection to Loki
func (c *Client) TestConnection(ctx context.Context) error {
	if !c.enabled {
		return nil
	}

	// Test with a simple health check
	url := c.config.URL + "/ready"
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.addAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Loki: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Loki health check failed: %d %s", resp.StatusCode, string(body))
	}

	c.logger.Debug("✅ Loki connection test successful",
		slog.String("component", "loki"))

	return nil
}

// Close cleans up resources and closes connections
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.enabled {
		return nil
	}

	c.logger.Info("🔌 Closing Loki integration",
		slog.String("component", "loki"))

	// Stop batch processor
	if c.batchTicker != nil {
		c.batchTicker.Stop()
	}
	close(c.stopChan)

	// Flush any remaining logs
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := c.flushBuffer(ctx); err != nil {
		c.logger.Error("❌ Failed to flush logs on close",
			slog.String("component", "loki"),
			slog.String("error", err.Error()))
	}

	c.enabled = false
	c.status.Connected = false

	c.logger.Info("✅ Loki integration closed successfully",
		slog.String("component", "loki"))

	return nil
}

// batchProcessor processes logs in batches
func (c *Client) batchProcessor() {
	for {
		select {
		case <-c.batchTicker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := c.flushBuffer(ctx); err != nil {
				c.logger.Error("❌ Failed to flush log buffer periodically",
					slog.String("component", "loki"),
					slog.String("error", err.Error()))
			}
			cancel()
		case <-c.stopChan:
			return
		}
	}
}

// flushBuffer sends all buffered logs to Loki
func (c *Client) flushBuffer(ctx context.Context) error {
	c.bufferMu.Lock()
	defer c.bufferMu.Unlock()

	if len(c.buffer) == 0 {
		return nil
	}

	logs := make([]interfaces.LogEntry, len(c.buffer))
	copy(logs, c.buffer)
	c.buffer = c.buffer[:0] // Clear buffer

	return c.SendLogs(ctx, logs)
}

// convertToLokiStreams converts log entries to Loki streams format
func (c *Client) convertToLokiStreams(logs []interfaces.LogEntry) map[string][]lokiLogEntry {
	streams := make(map[string][]lokiLogEntry)

	for _, log := range logs {
		// Build labels for this log entry
		labels := make(map[string]string)

		// Add static labels
		for k, v := range c.config.StaticLabels {
			labels[k] = v
		}

		// Add log-specific labels
		for k, v := range log.Labels {
			labels[k] = v
		}

		// Add dynamic labels
		for _, labelName := range c.config.DynamicLabels {
			switch labelName {
			case "level":
				labels["level"] = log.Level.String()
			case "component":
				if log.Component != "" {
					labels["component"] = log.Component
				}
			}
		}

		// Create label string (sorted for consistency)
		labelString := c.formatLabels(labels)

		// Add to appropriate stream
		entry := lokiLogEntry{
			Timestamp: strconv.FormatInt(log.Timestamp.UnixNano(), 10),
			Line:      c.formatLogLine(log),
		}

		streams[labelString] = append(streams[labelString], entry)
	}

	return streams
}

// formatLabels formats labels into Loki's label string format
func (c *Client) formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return "{}"
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	buf.WriteString("{")
	for i, k := range keys {
		if i > 0 {
			buf.WriteString(",")
		}
		buf.WriteString(k)
		buf.WriteString("=\"")
		buf.WriteString(labels[k])
		buf.WriteString("\"")
	}
	buf.WriteString("}")

	return buf.String()
}

// formatLogLine formats a log entry into a single line
func (c *Client) formatLogLine(log interfaces.LogEntry) string {
	line := log.Message

	// Add fields if present
	if len(log.Fields) > 0 {
		fieldsJSON, _ := json.Marshal(log.Fields)
		line += " " + string(fieldsJSON)
	}

	return line
}

// pushStreams sends streams to Loki
func (c *Client) pushStreams(ctx context.Context, streams map[string][]lokiLogEntry) error {
	if len(streams) == 0 {
		return nil
	}

	// Build Loki push request
	pushReq := lokiPushRequest{
		Streams: make([]lokiStream, 0, len(streams)),
	}

	for labels, entries := range streams {
		stream := lokiStream{
			Stream: json.RawMessage(labels),
			Values: make([][]string, len(entries)),
		}

		for i, entry := range entries {
			stream.Values[i] = []string{entry.Timestamp, entry.Line}
		}

		pushReq.Streams = append(pushReq.Streams, stream)
	}

	// Marshal request
	payload, err := json.Marshal(pushReq)
	if err != nil {
		return fmt.Errorf("failed to marshal Loki request: %w", err)
	}

	// Send to Loki
	url := c.config.URL + "/loki/api/v1/push"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	c.addAuth(req)

	if c.config.TenantID != "" {
		req.Header.Set("X-Scope-OrgID", c.config.TenantID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send logs to Loki: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Loki push failed: %d %s", resp.StatusCode, string(body))
	}

	c.logger.Debug("✅ Successfully sent logs to Loki",
		slog.String("component", "loki"),
		slog.Int("stream_count", len(streams)))

	return nil
}

// addAuth adds authentication to the request
func (c *Client) addAuth(req *http.Request) {
	if c.config.Username != "" && c.config.Password != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}
}

// Loki API types
type lokiPushRequest struct {
	Streams []lokiStream `json:"streams"`
}

type lokiStream struct {
	Stream json.RawMessage `json:"stream"`
	Values [][]string      `json:"values"`
}

type lokiLogEntry struct {
	Timestamp string
	Line      string
}

// Compile-time interface checks
var (
	_ interfaces.MonitoringIntegration = (*Client)(nil)
	_ interfaces.LogsIntegration      = (*Client)(nil)
)