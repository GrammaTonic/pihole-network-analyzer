package web

import (
	"context"
	"encoding/json"
	"net/http"
	"pihole-analyzer/internal/types"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// HandleWebSocket handles WebSocket upgrade and connection management
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	if s.wsManager == nil {
		s.logger.Error("WebSocket manager not initialized")
		http.Error(w, "WebSocket not available", http.StatusServiceUnavailable)
		return
	}

	// Upgrade the connection to WebSocket
	conn, err := s.wsManager.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.ErrorFields("Failed to upgrade WebSocket connection", map[string]any{
			"error":       err.Error(),
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
		})
		return
	}

	// Create WebSocket connection
	connectionID := uuid.New().String()
	wsConn := &WebSocketConnection{
		ID:          connectionID,
		Conn:        conn,
		Send:        make(chan []byte, 256),
		Manager:     s.wsManager,
		LastPing:    time.Now(),
		IsAlive:     true,
		ConnectedAt: time.Now(),
		RemoteAddr:  r.RemoteAddr,
		UserAgent:   r.UserAgent(),
	}

	// Add connection to manager
	s.wsManager.AddConnection(wsConn)

	// Start connection handlers
	go wsConn.writePump()
	go wsConn.readPump()

	// Send initial data
	s.sendInitialData(wsConn)
}

// writePump handles writing messages to the WebSocket connection
func (c *WebSocketConnection) writePump() {
	defer func() {
		c.Manager.logger.DebugFields("WebSocket write pump stopping", map[string]any{
			"connection_id": c.ID,
			"remote_addr":   c.RemoteAddr,
		})
		c.Conn.Close()
	}()

	// Set up ping ticker
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Channel closed
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.Manager.logger.DebugFields("Failed to write WebSocket message", map[string]any{
					"connection_id": c.ID,
					"error":         err.Error(),
				})
				return
			}

		case <-pingTicker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.Manager.logger.DebugFields("Failed to send WebSocket ping", map[string]any{
					"connection_id": c.ID,
					"error":         err.Error(),
				})
				return
			}
		}
	}
}

// readPump handles reading messages from the WebSocket connection
func (c *WebSocketConnection) readPump() {
	defer func() {
		c.Manager.logger.DebugFields("WebSocket read pump stopping", map[string]any{
			"connection_id": c.ID,
			"remote_addr":   c.RemoteAddr,
		})
		c.Manager.RemoveConnection(c.ID)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.LastPing = time.Now()
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Manager.logger.DebugFields("WebSocket connection closed unexpectedly", map[string]any{
					"connection_id": c.ID,
					"error":         err.Error(),
				})
			}
			break
		}

		// Handle incoming messages (commands from client)
		c.handleClientMessage(message)
	}
}

// handleClientMessage processes messages received from the WebSocket client
func (c *WebSocketConnection) handleClientMessage(message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		c.Manager.logger.DebugFields("Failed to parse WebSocket message", map[string]any{
			"connection_id": c.ID,
			"error":         err.Error(),
			"message":       string(message),
		})
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		c.Manager.logger.DebugFields("WebSocket message missing type", map[string]any{
			"connection_id": c.ID,
			"message":       msg,
		})
		return
	}

	c.Manager.logger.DebugFields("Received WebSocket message", map[string]any{
		"connection_id": c.ID,
		"type":          msgType,
	})

	switch msgType {
	case "ping":
		c.sendPong()
	case "subscribe":
		c.handleSubscription(msg)
	case "unsubscribe":
		c.handleUnsubscription(msg)
	case "request_update":
		c.handleUpdateRequest(msg)
	default:
		c.Manager.logger.DebugFields("Unknown WebSocket message type", map[string]any{
			"connection_id": c.ID,
			"type":          msgType,
		})
	}
}

// sendPong sends a pong response to the client
func (c *WebSocketConnection) sendPong() {
	response := map[string]interface{}{
		"type":      "pong",
		"timestamp": time.Now(),
	}

	data, err := json.Marshal(response)
	if err != nil {
		c.Manager.logger.ErrorFields("Failed to marshal pong response", map[string]any{
			"connection_id": c.ID,
			"error":         err.Error(),
		})
		return
	}

	select {
	case c.Send <- data:
		// Sent successfully
	default:
		c.Manager.logger.WarnFields("Failed to send pong, connection buffer full", map[string]any{
			"connection_id": c.ID,
		})
	}
}

// handleSubscription handles client subscription requests
func (c *WebSocketConnection) handleSubscription(msg map[string]interface{}) {
	channel, ok := msg["channel"].(string)
	if !ok {
		c.Manager.logger.DebugFields("Subscription missing channel", map[string]any{
			"connection_id": c.ID,
		})
		return
	}

	c.Manager.logger.InfoFields("Client subscribed to channel", map[string]any{
		"connection_id": c.ID,
		"channel":       channel,
	})

	// Send confirmation
	response := map[string]interface{}{
		"type":      "subscription_confirmed",
		"channel":   channel,
		"timestamp": time.Now(),
	}

	data, err := json.Marshal(response)
	if err != nil {
		c.Manager.logger.ErrorFields("Failed to marshal subscription confirmation", map[string]any{
			"connection_id": c.ID,
			"error":         err.Error(),
		})
		return
	}

	select {
	case c.Send <- data:
		// Sent successfully
	default:
		c.Manager.logger.WarnFields("Failed to send subscription confirmation", map[string]any{
			"connection_id": c.ID,
		})
	}
}

// handleUnsubscription handles client unsubscription requests
func (c *WebSocketConnection) handleUnsubscription(msg map[string]interface{}) {
	channel, ok := msg["channel"].(string)
	if !ok {
		c.Manager.logger.DebugFields("Unsubscription missing channel", map[string]any{
			"connection_id": c.ID,
		})
		return
	}

	c.Manager.logger.InfoFields("Client unsubscribed from channel", map[string]any{
		"connection_id": c.ID,
		"channel":       channel,
	})
}

// handleUpdateRequest handles client requests for immediate updates
func (c *WebSocketConnection) handleUpdateRequest(msg map[string]interface{}) {
	updateType, ok := msg["update_type"].(string)
	if !ok {
		updateType = "full"
	}

	c.Manager.logger.DebugFields("Client requested update", map[string]any{
		"connection_id": c.ID,
		"update_type":   updateType,
	})

	// This will be handled by the server to send current data
	// The actual implementation will depend on the server's data source
}

// sendInitialData sends initial dashboard data to a newly connected client
func (s *Server) sendInitialData(conn *WebSocketConnection) {
	ctx := context.Background()

	// Get current analysis data
	analysisResult, err := s.dataSource.GetAnalysisResult(ctx)
	if err != nil {
		s.logger.ErrorFields("Failed to get analysis result for WebSocket", map[string]any{
			"connection_id": conn.ID,
			"error":         err.Error(),
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

	// Send initial data
	s.wsManager.BroadcastUpdate("dashboard_init", update)

	s.logger.InfoFields("Sent initial data to WebSocket client", map[string]any{
		"connection_id": conn.ID,
	})
}

// createDashboardSummary creates a summary of the dashboard data
func (s *Server) createDashboardSummary(result *types.AnalysisResult) *DashboardSummary {
	if result == nil {
		return &DashboardSummary{
			LastUpdateTime: time.Now(),
		}
	}

	summary := &DashboardSummary{
		TotalQueries:   result.TotalQueries,
		TotalClients:   result.UniqueClients,
		LastUpdateTime: time.Now(),
	}

	// Count online clients from ClientStats map
	for _, client := range result.ClientStats {
		if client != nil && client.IsOnline {
			summary.OnlineClients++
		}
	}

	// Find top domain from ClientStats
	var topDomainCount int
	for _, client := range result.ClientStats {
		if client != nil && len(client.TopDomains) > 0 {
			for _, domain := range client.TopDomains {
				if domain.Count > topDomainCount {
					topDomainCount = domain.Count
					summary.TopDomain = domain.Domain
				}
			}
		}
	}

	// Set basic metrics (these would need to be calculated from actual query data)
	// For now, we'll use placeholder values since we don't have direct access to Records
	summary.BlockedQueries = 0
	summary.BlockedPercentage = 0.0
	summary.AverageResponse = 0.0

	return summary
}
