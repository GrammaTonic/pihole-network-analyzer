package web

import (
	"context"
	"time"

	"pihole-analyzer/internal/types"
)

// StartRealTimeUpdates starts the real-time update broadcaster for WebSocket clients
func (s *Server) StartRealTimeUpdates(ctx context.Context, updateInterval time.Duration) {
	if s.wsManager == nil {
		s.logger.Info("WebSocket manager not available, skipping real-time updates")
		return
	}

	s.logger.InfoFields("Starting real-time updates", map[string]any{
		"update_interval": updateInterval.String(),
	})

	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Real-time updates stopping")
			return
		case <-ticker.C:
			s.broadcastCurrentData()
		}
	}
}

// broadcastCurrentData fetches current data and broadcasts it to WebSocket clients
func (s *Server) broadcastCurrentData() {
	if s.wsManager.GetConnectionCount() == 0 {
		// No clients connected, skip update
		return
	}

	ctx := context.Background()

	// Get current analysis data
	analysisResult, err := s.dataSource.GetAnalysisResult(ctx)
	if err != nil {
		s.logger.ErrorFields("Failed to get analysis result for broadcast", map[string]any{
			"error":       err.Error(),
			"connections": s.wsManager.GetConnectionCount(),
		})
		return
	}

	// Get connection status
	connectionStatus := s.dataSource.GetConnectionStatus()

	// Create dashboard update
	update := &DashboardUpdate{
		ConnectionStatus: connectionStatus,
		AnalysisResult:   analysisResult,
		ClientStats:      analysisResult.ClientStats,
		Summary:          s.createDashboardSummary(analysisResult),
	}

	// Broadcast the update
	s.wsManager.BroadcastUpdate("dashboard_update", update)

	s.logger.DebugFields("Broadcasted real-time update", map[string]any{
		"connections":    s.wsManager.GetConnectionCount(),
		"total_clients":  update.Summary.TotalClients,
		"online_clients": update.Summary.OnlineClients,
	})
}

// BroadcastClientUpdate broadcasts a client-specific update
func (s *Server) BroadcastClientUpdate(clientIP string, action string, client *types.ClientStats) {
	if s.wsManager == nil {
		return
	}

	update := &ClientUpdate{
		Client: *client,
		Action: action,
	}

	s.wsManager.BroadcastUpdate("client_update", update)

	s.logger.DebugFields("Broadcasted client update", map[string]any{
		"client_ip":   clientIP,
		"action":      action,
		"connections": s.wsManager.GetConnectionCount(),
	})
}

// BroadcastQueryUpdate broadcasts a new DNS query update
func (s *Server) BroadcastQueryUpdate(query *types.PiholeRecord, isNew bool) {
	if s.wsManager == nil {
		return
	}

	update := &QueryUpdate{
		Query:     *query,
		Client:    query.Client,
		IsNew:     isNew,
		Timestamp: time.Now(),
	}

	s.wsManager.BroadcastUpdate("query_update", update)

	s.logger.DebugFields("Broadcasted query update", map[string]any{
		"domain":      query.Domain,
		"client":      query.Client,
		"is_new":      isNew,
		"connections": s.wsManager.GetConnectionCount(),
	})
}

// BroadcastMetricsUpdate broadcasts metrics data update
func (s *Server) BroadcastMetricsUpdate(metrics *MetricsUpdate) {
	if s.wsManager == nil {
		return
	}

	s.wsManager.BroadcastUpdate("metrics_update", metrics)

	s.logger.DebugFields("Broadcasted metrics update", map[string]any{
		"queries_per_second": metrics.QueriesPerSecond,
		"response_time":      metrics.ResponseTime,
		"connections":        s.wsManager.GetConnectionCount(),
	})
}

// GetWebSocketStats returns WebSocket connection statistics
func (s *Server) GetWebSocketStats() map[string]interface{} {
	if s.wsManager == nil {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	stats := s.wsManager.GetConnectionStats()
	stats["enabled"] = true
	return stats
}
