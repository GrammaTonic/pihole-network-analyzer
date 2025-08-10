package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// ChartData represents generic chart data structure
type ChartData struct {
	Labels   []string               `json:"labels"`
	Datasets []ChartDataset         `json:"datasets"`
	Options  map[string]interface{} `json:"options,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ChartDataset represents a dataset in a chart
type ChartDataset struct {
	Label           string                 `json:"label"`
	Data            []interface{}          `json:"data"`
	BackgroundColor []string               `json:"backgroundColor,omitempty"`
	BorderColor     []string               `json:"borderColor,omitempty"`
	Fill            bool                   `json:"fill"`
	Tension         float64                `json:"tension,omitempty"`
	Options         map[string]interface{} `json:"options,omitempty"`
}

// TimeSeriesPoint represents a point in time series data
type TimeSeriesPoint struct {
	X interface{} `json:"x"` // timestamp
	Y interface{} `json:"y"` // value
}

// NetworkTopologyData represents network topology visualization data
type NetworkTopologyData struct {
	Nodes []TopologyNode `json:"nodes"`
	Links []TopologyLink `json:"links"`
}

// TopologyNode represents a node in network topology
type TopologyNode struct {
	ID       string                 `json:"id"`
	Label    string                 `json:"label"`
	Type     string                 `json:"type"` // "client", "domain", "server"
	Size     int                    `json:"size"`
	Color    string                 `json:"color"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TopologyLink represents a connection in network topology
type TopologyLink struct {
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Weight   int     `json:"weight"`
	Type     string  `json:"type"` // "query", "block", "allow"
	Strength float64 `json:"strength"`
}

// ChartAPIHandler handles chart data API requests
type ChartAPIHandler struct {
	logger     *logger.Logger
	dataSource DataSourceProvider
}

// NewChartAPIHandler creates a new chart API handler
func NewChartAPIHandler(dataSource DataSourceProvider, logger *logger.Logger) *ChartAPIHandler {
	return &ChartAPIHandler{
		logger:     logger.Component("chart-api"),
		dataSource: dataSource,
	}
}

// HandleTimelineChart handles timeline chart data requests
func (h *ChartAPIHandler) HandleTimelineChart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	timeWindow := r.URL.Query().Get("window")
	if timeWindow == "" {
		timeWindow = "24h"
	}

	granularity := r.URL.Query().Get("granularity")
	if granularity == "" {
		granularity = "1h"
	}

	h.logger.InfoFields("Timeline chart request", map[string]any{
		"window":      timeWindow,
		"granularity": granularity,
	})

	// Get analysis data
	result, err := h.dataSource.GetAnalysisResult(ctx)
	if err != nil {
		h.logger.ErrorFields("Failed to get analysis result", map[string]any{
			"error": err.Error(),
		})
		http.Error(w, "Failed to get data", http.StatusInternalServerError)
		return
	}

	// Generate timeline data
	chartData, err := h.generateTimelineData(result, timeWindow, granularity)
	if err != nil {
		h.logger.ErrorFields("Failed to marshal timeline chart data", map[string]any{
			"error": err.Error(),
		})
		http.Error(w, "Failed to process data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chartData)
}

// HandleClientChart handles client distribution chart data requests
func (h *ChartAPIHandler) HandleClientChart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	chartType := r.URL.Query().Get("type")
	if chartType == "" {
		chartType = "pie"
	}

	h.logger.InfoFields("Client chart request", map[string]any{
		"type": chartType,
	})

	// Get analysis data
	result, err := h.dataSource.GetAnalysisResult(ctx)
	if err != nil {
		h.logger.ErrorFields("Failed to get analysis result for client chart", map[string]any{
			"error": err.Error(),
		})
		http.Error(w, "Failed to get data", http.StatusInternalServerError)
		return
	}

	// Generate client distribution data
	chartData, err := h.generateClientDistributionData(result, chartType)
	if err != nil {
		h.logger.ErrorFields("Failed to generate client data", map[string]any{
			"error": err.Error(),
		})
		http.Error(w, "Failed to process data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chartData)
}

// HandleDomainChart handles domain statistics chart data requests
func (h *ChartAPIHandler) HandleDomainChart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}

	h.logger.InfoFields("Domain chart request", map[string]any{
		"limit": limit,
	})

	// Get analysis data
	result, err := h.dataSource.GetAnalysisResult(ctx)
	if err != nil {
		h.logger.ErrorFields("Failed to get analysis result for domain chart", map[string]any{
			"error": err.Error(),
		})
		http.Error(w, "Failed to get data", http.StatusInternalServerError)
		return
	}

	// Generate domain statistics data
	chartData, err := h.generateDomainStatsData(result, limit)
	if err != nil {
		h.logger.ErrorFields("Failed to generate domain data", map[string]any{
			"error": err.Error(),
		})
		http.Error(w, "Failed to process data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chartData)
}

// HandleTopologyChart handles network topology visualization data requests
func (h *ChartAPIHandler) HandleTopologyChart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	maxNodes, _ := strconv.Atoi(r.URL.Query().Get("max_nodes"))
	if maxNodes <= 0 {
		maxNodes = 50
	}

	h.logger.InfoFields("Topology chart request", map[string]any{
		"max_nodes": maxNodes,
	})

	// Get analysis data
	result, err := h.dataSource.GetAnalysisResult(ctx)
	if err != nil {
		h.logger.ErrorFields("Failed to get analysis result for topology chart", map[string]any{
			"error": err.Error(),
		})
		http.Error(w, "Failed to get data", http.StatusInternalServerError)
		return
	}

	// Generate network topology data
	topologyData, err := h.generateNetworkTopologyData(result, maxNodes)
	if err != nil {
		h.logger.ErrorFields("Failed to generate topology data", map[string]any{
			"error": err.Error(),
		})
		http.Error(w, "Failed to process data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topologyData)
}

// HandlePerformanceChart handles performance metrics chart data requests
func (h *ChartAPIHandler) HandlePerformanceChart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	h.logger.Info("Performance chart request")

	// Get analysis data
	result, err := h.dataSource.GetAnalysisResult(ctx)
	if err != nil {
		h.logger.ErrorFields("Failed to get analysis result for performance chart", map[string]any{
			"error": err.Error(),
		})
		http.Error(w, "Failed to get data", http.StatusInternalServerError)
		return
	}

	// Generate performance metrics data
	chartData, err := h.generatePerformanceData(result)
	if err != nil {
		h.logger.ErrorFields("Failed to generate performance data", map[string]any{
			"error": err.Error(),
		})
		http.Error(w, "Failed to process data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chartData)
}

// generateTimelineData generates time series data for query timeline
func (h *ChartAPIHandler) generateTimelineData(result *types.AnalysisResult, timeWindow, granularity string) (*ChartData, error) {
	// Parse time window
	duration, err := time.ParseDuration(timeWindow)
	if err != nil {
		return nil, fmt.Errorf("invalid time window: %w", err)
	}

	now := time.Now()
	startTime := now.Add(-duration)

	// Generate time labels based on granularity
	labels, err := h.generateTimeLabels(startTime, now, granularity)
	if err != nil {
		return nil, fmt.Errorf("failed to generate time labels: %w", err)
	}

	// For now, generate mock time series data
	// In a real implementation, this would query actual time-based data
	totalQueries := make([]interface{}, len(labels))
	allowedQueries := make([]interface{}, len(labels))
	blockedQueries := make([]interface{}, len(labels))

	// Generate realistic mock data based on actual totals
	baseQueries := result.TotalQueries / len(labels)
	for i := range labels {
		// Add some variance to make it realistic
		variance := int(float64(baseQueries) * 0.3 * (2*float64(i%3) - 1))
		total := baseQueries + variance
		if total < 0 {
			total = 0
		}

		// Assume ~15% blocked rate
		blocked := int(float64(total) * 0.15)
		allowed := total - blocked

		totalQueries[i] = total
		allowedQueries[i] = allowed
		blockedQueries[i] = blocked
	}

	chartData := &ChartData{
		Labels: labels,
		Datasets: []ChartDataset{
			{
				Label:           "Total Queries",
				Data:            totalQueries,
				BorderColor:     []string{"#3498db"},
				BackgroundColor: []string{"rgba(52, 152, 219, 0.1)"},
				Fill:            true,
				Tension:         0.4,
			},
			{
				Label:           "Allowed Queries",
				Data:            allowedQueries,
				BorderColor:     []string{"#27ae60"},
				BackgroundColor: []string{"rgba(39, 174, 96, 0.1)"},
				Fill:            true,
				Tension:         0.4,
			},
			{
				Label:           "Blocked Queries",
				Data:            blockedQueries,
				BorderColor:     []string{"#e74c3c"},
				BackgroundColor: []string{"rgba(231, 76, 60, 0.1)"},
				Fill:            true,
				Tension:         0.4,
			},
		},
		Options: map[string]interface{}{
			"responsive": true,
			"interaction": map[string]interface{}{
				"intersect": false,
			},
			"scales": map[string]interface{}{
				"x": map[string]interface{}{
					"display": true,
					"title": map[string]interface{}{
						"display": true,
						"text":    "Time",
					},
				},
				"y": map[string]interface{}{
					"display": true,
					"title": map[string]interface{}{
						"display": true,
						"text":    "Query Count",
					},
				},
			},
		},
		Metadata: map[string]interface{}{
			"timeWindow":  timeWindow,
			"granularity": granularity,
			"dataPoints":  len(labels),
		},
	}

	return chartData, nil
}

// generateClientDistributionData generates client activity distribution data
func (h *ChartAPIHandler) generateClientDistributionData(result *types.AnalysisResult, chartType string) (*ChartData, error) {
	if len(result.ClientStats) == 0 {
		return &ChartData{
			Labels:   []string{},
			Datasets: []ChartDataset{},
		}, nil
	}

	// Sort clients by query count and take top 10
	type clientData struct {
		IP       string
		Hostname string
		Count    int
	}

	clients := make([]clientData, 0, len(result.ClientStats))
	for ip, stats := range result.ClientStats {
		clients = append(clients, clientData{
			IP:       ip,
			Hostname: stats.Hostname,
			Count:    stats.QueryCount,
		})
	}

	// Simple sort by count (descending)
	for i := 0; i < len(clients)-1; i++ {
		for j := i + 1; j < len(clients); j++ {
			if clients[j].Count > clients[i].Count {
				clients[i], clients[j] = clients[j], clients[i]
			}
		}
	}

	// Take top 10
	if len(clients) > 10 {
		clients = clients[:10]
	}

	labels := make([]string, len(clients))
	data := make([]interface{}, len(clients))
	colors := []string{
		"#FF6384", "#36A2EB", "#FFCE56", "#4BC0C0", "#9966FF",
		"#FF9F40", "#C9CBCF", "#4BC0C0", "#FF6384", "#36A2EB",
	}

	for i, client := range clients {
		if client.Hostname != "" {
			labels[i] = client.Hostname
		} else {
			labels[i] = client.IP
		}
		data[i] = client.Count
	}

	chartData := &ChartData{
		Labels: labels,
		Datasets: []ChartDataset{
			{
				Label:           "Query Count",
				Data:            data,
				BackgroundColor: colors[:len(clients)],
				BorderColor:     colors[:len(clients)],
			},
		},
		Options: map[string]interface{}{
			"responsive": true,
			"plugins": map[string]interface{}{
				"legend": map[string]interface{}{
					"position": "right",
				},
				"tooltip": map[string]interface{}{
					"enabled": true,
				},
			},
		},
		Metadata: map[string]interface{}{
			"chartType":    chartType,
			"totalClients": len(result.ClientStats),
			"topClients":   len(clients),
		},
	}

	return chartData, nil
}

// generateDomainStatsData generates domain statistics data
func (h *ChartAPIHandler) generateDomainStatsData(result *types.AnalysisResult, limit int) (*ChartData, error) {
	// Aggregate domain data from all clients
	domainCounts := make(map[string]int)
	for _, clientStats := range result.ClientStats {
		for domain, count := range clientStats.Domains {
			domainCounts[domain] += count
		}
	}

	if len(domainCounts) == 0 {
		return &ChartData{
			Labels:   []string{},
			Datasets: []ChartDataset{},
		}, nil
	}

	// Convert to slice for sorting
	type domainData struct {
		Domain string
		Count  int
	}

	domains := make([]domainData, 0, len(domainCounts))
	for domain, count := range domainCounts {
		domains = append(domains, domainData{
			Domain: domain,
			Count:  count,
		})
	}

	// Sort by count (descending)
	for i := 0; i < len(domains)-1; i++ {
		for j := i + 1; j < len(domains); j++ {
			if domains[j].Count > domains[i].Count {
				domains[i], domains[j] = domains[j], domains[i]
			}
		}
	}

	// Take top N
	if len(domains) > limit {
		domains = domains[:limit]
	}

	labels := make([]string, len(domains))
	data := make([]interface{}, len(domains))

	for i, domain := range domains {
		labels[i] = domain.Domain
		data[i] = domain.Count
	}

	chartData := &ChartData{
		Labels: labels,
		Datasets: []ChartDataset{
			{
				Label:           "Query Count",
				Data:            data,
				BackgroundColor: []string{"#36A2EB"},
				BorderColor:     []string{"#36A2EB"},
			},
		},
		Options: map[string]interface{}{
			"responsive": true,
			"indexAxis":  "y", // Horizontal bar chart
			"plugins": map[string]interface{}{
				"legend": map[string]interface{}{
					"display": false,
				},
			},
			"scales": map[string]interface{}{
				"x": map[string]interface{}{
					"beginAtZero": true,
				},
			},
		},
		Metadata: map[string]interface{}{
			"totalDomains":     len(domainCounts),
			"displayedDomains": len(domains),
			"limit":            limit,
		},
	}

	return chartData, nil
}

// generateNetworkTopologyData generates network topology visualization data
func (h *ChartAPIHandler) generateNetworkTopologyData(result *types.AnalysisResult, maxNodes int) (*NetworkTopologyData, error) {
	nodes := make([]TopologyNode, 0)
	links := make([]TopologyLink, 0)

	// Add Pi-hole server node
	nodes = append(nodes, TopologyNode{
		ID:    "pihole-server",
		Label: "Pi-hole Server",
		Type:  "server",
		Size:  20,
		Color: "#e74c3c",
		Metadata: map[string]interface{}{
			"role": "dns_server",
		},
	})

	// Add client nodes
	clientCount := 0
	for ip, stats := range result.ClientStats {
		if clientCount >= maxNodes-1 { // Reserve one slot for Pi-hole server
			break
		}

		nodeLabel := stats.Hostname
		if nodeLabel == "" {
			nodeLabel = ip
		}

		nodeSize := 5 + int(float64(stats.QueryCount)/float64(result.TotalQueries)*20)
		if nodeSize > 25 {
			nodeSize = 25
		}

		nodes = append(nodes, TopologyNode{
			ID:    ip,
			Label: nodeLabel,
			Type:  "client",
			Size:  nodeSize,
			Color: "#3498db",
			Metadata: map[string]interface{}{
				"ip":        ip,
				"queries":   stats.QueryCount,
				"is_online": stats.IsOnline,
			},
		})

		// Add link from client to Pi-hole server
		linkStrength := float64(stats.QueryCount) / float64(result.TotalQueries)
		links = append(links, TopologyLink{
			Source:   ip,
			Target:   "pihole-server",
			Weight:   stats.QueryCount,
			Type:     "query",
			Strength: linkStrength,
		})

		clientCount++
	}

	return &NetworkTopologyData{
		Nodes: nodes,
		Links: links,
	}, nil
}

// generatePerformanceData generates performance metrics data
func (h *ChartAPIHandler) generatePerformanceData(result *types.AnalysisResult) (*ChartData, error) {
	// Mock performance data - in real implementation, this would come from metrics
	performanceMetrics := map[string]interface{}{
		"totalQueries":    result.TotalQueries,
		"uniqueClients":   result.UniqueClients,
		"activeDevices":   len(result.NetworkDevices),
		"queryRate":       float64(result.TotalQueries) / 3600.0, // Queries per second (assuming 1 hour window)
		"avgResponseTime": 15.5,                                  // Mock average response time in ms
		"uptime":          99.8,                                  // Mock uptime percentage
	}

	labels := []string{"Total Queries", "Unique Clients", "Active Devices", "Queries/sec", "Avg Response (ms)", "Uptime (%)"}
	data := []interface{}{
		result.TotalQueries,
		result.UniqueClients,
		len(result.NetworkDevices),
		int(float64(result.TotalQueries) / 3600.0),
		15.5,
		99.8,
	}

	chartData := &ChartData{
		Labels: labels,
		Datasets: []ChartDataset{
			{
				Label:           "Performance Metrics",
				Data:            data,
				BackgroundColor: []string{"#27ae60", "#3498db", "#f39c12", "#9b59b6", "#e67e22", "#1abc9c"},
				BorderColor:     []string{"#27ae60", "#3498db", "#f39c12", "#9b59b6", "#e67e22", "#1abc9c"},
			},
		},
		Metadata: performanceMetrics,
	}

	return chartData, nil
}

// generateTimeLabels generates time labels for the given time range and granularity
func (h *ChartAPIHandler) generateTimeLabels(start, end time.Time, granularity string) ([]string, error) {
	interval, err := time.ParseDuration(granularity)
	if err != nil {
		return nil, fmt.Errorf("invalid granularity: %w", err)
	}

	labels := make([]string, 0)
	current := start

	for current.Before(end) {
		labels = append(labels, current.Format("15:04"))
		current = current.Add(interval)
	}

	if len(labels) == 0 {
		labels = append(labels, start.Format("15:04"))
	}

	return labels, nil
}
