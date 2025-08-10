package grafana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"pihole-analyzer/internal/integrations/interfaces"
	"pihole-analyzer/internal/types"
)

// Client implements Grafana integration for dashboards, data sources, and alerts
type Client struct {
	config     *types.GrafanaConfig
	logger     *slog.Logger
	httpClient *http.Client
	enabled    bool
	status     interfaces.IntegrationStatus
	mu         sync.RWMutex
}

// NewClient creates a new Grafana integration client
func NewClient(config *types.GrafanaConfig, logger *slog.Logger) *Client {
	return &Client{
		config:     config,
		logger:     logger,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		enabled: config.Enabled,
		status: interfaces.IntegrationStatus{
			Name:    "grafana",
			Enabled: config.Enabled,
		},
	}
}

// Initialize sets up the Grafana integration
func (c *Client) Initialize(ctx context.Context, config interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if grafanaConfig, ok := config.(*types.GrafanaConfig); ok {
		c.config = grafanaConfig
		c.enabled = grafanaConfig.Enabled
	}

	if !c.enabled {
		c.logger.Debug("⏭️ Grafana integration disabled",
			slog.String("component", "grafana"))
		return nil
	}

	c.logger.Info("🚀 Initializing Grafana integration",
		slog.String("component", "grafana"),
		slog.String("url", c.config.URL))

	// Test connection
	if err := c.TestConnection(ctx); err != nil {
		c.status.LastError = err.Error()
		return fmt.Errorf("failed to connect to Grafana: %w", err)
	}

	// Set up data source if configured
	if c.config.DataSource.CreateIfNotExists {
		if err := c.setupDataSource(ctx); err != nil {
			c.logger.Warn("⚠️ Failed to setup data source",
				slog.String("component", "grafana"),
				slog.String("error", err.Error()))
		}
	}

	c.status.Connected = true
	c.status.LastConnect = time.Now()

	c.logger.Info("✅ Grafana integration initialized successfully",
		slog.String("component", "grafana"))

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
	return "grafana"
}

// GetStatus returns the current status of the integration
func (c *Client) GetStatus() interfaces.IntegrationStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

// SendMetrics sends metrics data to Grafana (creates/updates dashboards)
func (c *Client) SendMetrics(ctx context.Context, data *types.AnalysisResult) error {
	if !c.enabled {
		return nil
	}

	c.logger.Debug("📊 Processing metrics for Grafana dashboards",
		slog.String("component", "grafana"),
		slog.Int("total_queries", data.TotalQueries),
		slog.Int("unique_clients", data.UniqueClients))

	// Auto-provision dashboards if enabled
	if c.config.Dashboards.AutoProvision {
		if err := c.provisionDashboards(ctx, data); err != nil {
			c.logger.Error("❌ Failed to provision dashboards",
				slog.String("component", "grafana"),
				slog.String("error", err.Error()))
			return err
		}
	}

	return nil
}

// SendLogs sends log data (not implemented for Grafana)
func (c *Client) SendLogs(ctx context.Context, logs []interfaces.LogEntry) error {
	// Grafana doesn't directly receive logs - this is a no-op
	return nil
}

// TestConnection tests the connection to Grafana
func (c *Client) TestConnection(ctx context.Context) error {
	if !c.enabled {
		return nil
	}

	// Test with health endpoint
	url := c.config.URL + "/api/health"
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.addAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Grafana: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Grafana health check failed: %d %s", resp.StatusCode, string(body))
	}

	c.logger.Debug("✅ Grafana connection test successful",
		slog.String("component", "grafana"))

	return nil
}

// Close cleans up resources and closes connections
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.enabled {
		return nil
	}

	c.logger.Info("🔌 Closing Grafana integration",
		slog.String("component", "grafana"))

	c.enabled = false
	c.status.Connected = false

	c.logger.Info("✅ Grafana integration closed successfully",
		slog.String("component", "grafana"))

	return nil
}

// CreateDashboard creates a new dashboard
func (c *Client) CreateDashboard(ctx context.Context, dashboard interfaces.Dashboard) error {
	if !c.enabled {
		return fmt.Errorf("Grafana integration not enabled")
	}

	c.logger.Info("📋 Creating Grafana dashboard",
		slog.String("component", "grafana"),
		slog.String("title", dashboard.Title))

	// Convert to Grafana dashboard format
	grafanaDashboard := c.convertToGrafanaDashboard(dashboard)

	// Create dashboard payload
	payload := map[string]interface{}{
		"dashboard": grafanaDashboard,
		"overwrite": c.config.Dashboards.OverwriteExisting,
	}

	if dashboard.FolderID != "" {
		payload["folderId"] = dashboard.FolderID
	}

	return c.sendDashboardRequest(ctx, "POST", "/api/dashboards/db", payload)
}

// UpdateDashboard updates an existing dashboard
func (c *Client) UpdateDashboard(ctx context.Context, dashboard interfaces.Dashboard) error {
	if !c.enabled {
		return fmt.Errorf("Grafana integration not enabled")
	}

	c.logger.Info("📝 Updating Grafana dashboard",
		slog.String("component", "grafana"),
		slog.String("title", dashboard.Title))

	return c.CreateDashboard(ctx, dashboard) // Same endpoint with overwrite=true
}

// DeleteDashboard deletes a dashboard
func (c *Client) DeleteDashboard(ctx context.Context, uid string) error {
	if !c.enabled {
		return fmt.Errorf("Grafana integration not enabled")
	}

	c.logger.Info("🗑️ Deleting Grafana dashboard",
		slog.String("component", "grafana"),
		slog.String("uid", uid))

	url := fmt.Sprintf("/api/dashboards/uid/%s", uid)
	
	req, err := c.createRequest(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete dashboard: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete dashboard: %d %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListDashboards lists all dashboards
func (c *Client) ListDashboards(ctx context.Context) ([]interfaces.Dashboard, error) {
	if !c.enabled {
		return nil, fmt.Errorf("Grafana integration not enabled")
	}

	req, err := c.createRequest(ctx, "GET", "/api/search?type=dash-db", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list dashboards: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list dashboards: %d %s", resp.StatusCode, string(body))
	}

	var searchResults []grafanaSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&searchResults); err != nil {
		return nil, fmt.Errorf("failed to decode dashboard list: %w", err)
	}

	// Convert to interface format
	dashboards := make([]interfaces.Dashboard, len(searchResults))
	for i, result := range searchResults {
		dashboards[i] = interfaces.Dashboard{
			ID:          result.UID,
			Title:       result.Title,
			FolderID:    fmt.Sprintf("%d", result.FolderID),
			Tags:        result.Tags,
		}
	}

	return dashboards, nil
}

// GetDashboard retrieves a specific dashboard
func (c *Client) GetDashboard(ctx context.Context, uid string) (*interfaces.Dashboard, error) {
	if !c.enabled {
		return nil, fmt.Errorf("Grafana integration not enabled")
	}

	url := fmt.Sprintf("/api/dashboards/uid/%s", uid)
	
	req, err := c.createRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get dashboard: %d %s", resp.StatusCode, string(body))
	}

	var result grafanaDashboardResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode dashboard: %w", err)
	}

	// Convert to interface format
	dashboard := &interfaces.Dashboard{
		ID:         getStringFromMap(result.Dashboard, "uid"),
		Title:      getStringFromMap(result.Dashboard, "title"),
		Tags:       getStringSliceFromMap(result.Dashboard, "tags"),
		Definition: result.Dashboard,
	}

	return dashboard, nil
}

// setupDataSource creates or updates the configured data source
func (c *Client) setupDataSource(ctx context.Context) error {
	c.logger.Info("🔧 Setting up Grafana data source",
		slog.String("component", "grafana"),
		slog.String("name", c.config.DataSource.Name))

	// Check if data source exists
	existing, err := c.getDataSourceByName(ctx, c.config.DataSource.Name)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("failed to check existing data source: %w", err)
	}

	dataSource := map[string]interface{}{
		"name":   c.config.DataSource.Name,
		"type":   c.config.DataSource.Type,
		"url":    c.config.DataSource.URL,
		"access": c.config.DataSource.Access,
	}

	if c.config.DataSource.BasicAuth {
		dataSource["basicAuth"] = true
		dataSource["basicAuthUser"] = c.config.DataSource.Username
		dataSource["basicAuthPassword"] = c.config.DataSource.Password
	}

	var endpoint string
	var method string

	if existing != nil {
		// Update existing
		endpoint = fmt.Sprintf("/api/datasources/%d", existing.ID)
		method = "PUT"
		dataSource["id"] = existing.ID
	} else {
		// Create new
		endpoint = "/api/datasources"
		method = "POST"
	}

	return c.sendDataSourceRequest(ctx, method, endpoint, dataSource)
}

// provisionDashboards creates or updates dashboards based on analysis data
func (c *Client) provisionDashboards(ctx context.Context, data *types.AnalysisResult) error {
	c.logger.Debug("📊 Provisioning Grafana dashboards",
		slog.String("component", "grafana"))

	// Create Pi-hole Network Analyzer dashboard
	dashboard := c.createMainDashboard(data)
	
	if err := c.CreateDashboard(ctx, dashboard); err != nil {
		return fmt.Errorf("failed to provision main dashboard: %w", err)
	}

	c.logger.Info("✅ Successfully provisioned Grafana dashboards",
		slog.String("component", "grafana"))

	return nil
}

// createMainDashboard creates the main Pi-hole Network Analyzer dashboard
func (c *Client) createMainDashboard(data *types.AnalysisResult) interfaces.Dashboard {
	dashboard := interfaces.Dashboard{
		Title:       "Pi-hole Network Analyzer",
		Description: "Comprehensive network analysis and DNS monitoring dashboard",
		Tags:        c.config.Dashboards.Tags,
		Panels: []interfaces.Panel{
			{
				ID:    1,
				Title: "Total DNS Queries",
				Type:  "stat",
				Query: "sum(pihole_analyzer_total_queries)",
			},
			{
				ID:    2,
				Title: "Active Clients",
				Type:  "stat",
				Query: "pihole_analyzer_active_clients",
			},
			{
				ID:    3,
				Title: "Queries by Type",
				Type:  "piechart",
				Query: "sum by (query_type) (pihole_analyzer_queries_by_type_total)",
			},
			{
				ID:    4,
				Title: "Query Response Time",
				Type:  "timeseries",
				Query: "histogram_quantile(0.95, pihole_analyzer_query_response_time_seconds)",
			},
			{
				ID:    5,
				Title: "Top Domains",
				Type:  "table",
				Query: "topk(10, sum by (domain) (pihole_analyzer_top_domains_total))",
			},
			{
				ID:    6,
				Title: "Blocked vs Allowed",
				Type:  "timeseries",
				Query: "rate(pihole_analyzer_blocked_domains_total[5m]) or rate(pihole_analyzer_allowed_domains_total[5m])",
			},
		},
	}

	return dashboard
}

// Helper methods

func (c *Client) createRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reader = bytes.NewReader(jsonBody)
	}

	url := c.config.URL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	c.addAuth(req)
	return req, nil
}

func (c *Client) addAuth(req *http.Request) {
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}
}

func (c *Client) sendDashboardRequest(ctx context.Context, method, endpoint string, payload interface{}) error {
	req, err := c.createRequest(ctx, method, endpoint, payload)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send dashboard request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("dashboard request failed: %d %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) sendDataSourceRequest(ctx context.Context, method, endpoint string, payload interface{}) error {
	req, err := c.createRequest(ctx, method, endpoint, payload)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send data source request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("data source request failed: %d %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) getDataSourceByName(ctx context.Context, name string) (*grafanaDataSource, error) {
	req, err := c.createRequest(ctx, "GET", "/api/datasources", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get data sources: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get data sources: %d", resp.StatusCode)
	}

	var dataSources []grafanaDataSource
	if err := json.NewDecoder(resp.Body).Decode(&dataSources); err != nil {
		return nil, fmt.Errorf("failed to decode data sources: %w", err)
	}

	for _, ds := range dataSources {
		if ds.Name == name {
			return &ds, nil
		}
	}

	return nil, fmt.Errorf("data source %s not found", name)
}

func (c *Client) convertToGrafanaDashboard(dashboard interfaces.Dashboard) map[string]interface{} {
	// Convert interface Dashboard to Grafana format
	grafanaDashboard := map[string]interface{}{
		"title":       dashboard.Title,
		"description": dashboard.Description,
		"tags":        dashboard.Tags,
		"panels":      c.convertPanels(dashboard.Panels),
		"time": map[string]interface{}{
			"from": "now-1h",
			"to":   "now",
		},
		"refresh": "30s",
	}

	return grafanaDashboard
}

func (c *Client) convertPanels(panels []interfaces.Panel) []map[string]interface{} {
	grafanaPanels := make([]map[string]interface{}, len(panels))
	
	for i, panel := range panels {
		grafanaPanels[i] = map[string]interface{}{
			"id":    panel.ID,
			"title": panel.Title,
			"type":  panel.Type,
			"targets": []map[string]interface{}{
				{
					"expr": panel.Query,
				},
			},
			"gridPos": map[string]interface{}{
				"h": 8,
				"w": 12,
				"x": (i % 2) * 12,
				"y": (i / 2) * 8,
			},
		}
	}

	return grafanaPanels
}

// Grafana API types
type grafanaSearchResult struct {
	ID       int      `json:"id"`
	UID      string   `json:"uid"`
	Title    string   `json:"title"`
	Tags     []string `json:"tags"`
	FolderID int      `json:"folderId"`
}

type grafanaDashboardResponse struct {
	Dashboard map[string]interface{} `json:"dashboard"`
	Meta      map[string]interface{} `json:"meta"`
}

type grafanaDataSource struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// Helper functions to extract values from map[string]interface{}
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getStringSliceFromMap(m map[string]interface{}, key string) []string {
	if val, ok := m[key]; ok {
		if slice, ok := val.([]interface{}); ok {
			result := make([]string, len(slice))
			for i, v := range slice {
				if str, ok := v.(string); ok {
					result[i] = str
				}
			}
			return result
		}
	}
	return []string{}
}

// Compile-time interface checks
var (
	_ interfaces.MonitoringIntegration = (*Client)(nil)
	_ interfaces.DashboardIntegration  = (*Client)(nil)
)