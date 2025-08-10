package web

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"

	"github.com/gorilla/websocket"
)

// WebSocketManager manages WebSocket connections and real-time updates
type WebSocketManager struct {
	logger      *logger.Logger
	connections map[string]*WebSocketConnection
	connMutex   sync.RWMutex
	upgrader    websocket.Upgrader
	updateChan  chan *UpdateMessage
	ctx         context.Context
	cancel      context.CancelFunc
}

// WebSocketConnection represents a single WebSocket connection
type WebSocketConnection struct {
	ID          string
	Conn        *websocket.Conn
	Send        chan []byte
	Manager     *WebSocketManager
	LastPing    time.Time
	IsAlive     bool
	ConnectedAt time.Time
	RemoteAddr  string
	UserAgent   string
}

// UpdateMessage represents a real-time update message
type UpdateMessage struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// DashboardUpdate represents the complete dashboard data update
type DashboardUpdate struct {
	ConnectionStatus *types.ConnectionStatus       `json:"connection_status"`
	AnalysisResult   *types.AnalysisResult         `json:"analysis_result"`
	ClientStats      map[string]*types.ClientStats `json:"client_stats"`
	Summary          *DashboardSummary             `json:"summary"`
}

// DashboardSummary provides key metrics for the dashboard
type DashboardSummary struct {
	TotalQueries      int       `json:"total_queries"`
	TotalClients      int       `json:"total_clients"`
	OnlineClients     int       `json:"online_clients"`
	TopDomain         string    `json:"top_domain"`
	QueryRate         float64   `json:"query_rate"`
	LastUpdateTime    time.Time `json:"last_update_time"`
	AverageResponse   float64   `json:"average_response"`
	BlockedQueries    int       `json:"blocked_queries"`
	BlockedPercentage float64   `json:"blocked_percentage"`
}

// ClientUpdate represents a client-specific update
type ClientUpdate struct {
	Client       types.ClientStats `json:"client"`
	Action       string            `json:"action"` // "added", "updated", "removed", "online", "offline"
	UpdateFields []string          `json:"update_fields,omitempty"`
}

// QueryUpdate represents a new DNS query update
type QueryUpdate struct {
	Query     types.PiholeRecord `json:"query"`
	Client    string             `json:"client"`
	IsNew     bool               `json:"is_new"`
	Timestamp time.Time          `json:"timestamp"`
}

// MetricsUpdate represents metrics data update
type MetricsUpdate struct {
	QueriesPerSecond   float64        `json:"queries_per_second"`
	ResponseTime       float64        `json:"response_time"`
	ClientDistribution map[string]int `json:"client_distribution"`
	DomainDistribution map[string]int `json:"domain_distribution"`
	QueryTypes         map[string]int `json:"query_types"`
	Timestamp          time.Time      `json:"timestamp"`
}

// WebSocketConfig holds WebSocket configuration
type WebSocketConfig struct {
	EnableWebSocket   bool          `json:"enable_websocket"`
	UpdateInterval    time.Duration `json:"update_interval"`
	PingInterval      time.Duration `json:"ping_interval"`
	WriteTimeout      time.Duration `json:"write_timeout"`
	ReadTimeout       time.Duration `json:"read_timeout"`
	MaxConnections    int           `json:"max_connections"`
	BufferSize        int           `json:"buffer_size"`
	EnableCompression bool          `json:"enable_compression"`
}

// DefaultWebSocketConfig returns default WebSocket configuration
func DefaultWebSocketConfig() *WebSocketConfig {
	return &WebSocketConfig{
		EnableWebSocket:   true,
		UpdateInterval:    2 * time.Second,
		PingInterval:      30 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadTimeout:       60 * time.Second,
		MaxConnections:    100,
		BufferSize:        256,
		EnableCompression: true,
	}
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager(logger *logger.Logger, config *WebSocketConfig) *WebSocketManager {
	if config == nil {
		config = DefaultWebSocketConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	wsLogger := logger.Component("websocket-manager")

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// In production, implement proper origin checking
			return true
		},
		EnableCompression: config.EnableCompression,
		ReadBufferSize:    config.BufferSize,
		WriteBufferSize:   config.BufferSize,
	}

	manager := &WebSocketManager{
		logger:      wsLogger,
		connections: make(map[string]*WebSocketConnection),
		upgrader:    upgrader,
		updateChan:  make(chan *UpdateMessage, 1000),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start the update broadcaster
	go manager.startUpdateBroadcaster()

	// Start connection cleanup routine
	go manager.startConnectionCleanup(config.PingInterval)

	wsLogger.InfoFields("WebSocket manager initialized", map[string]any{
		"update_interval": config.UpdateInterval.String(),
		"ping_interval":   config.PingInterval.String(),
		"max_connections": config.MaxConnections,
		"compression":     config.EnableCompression,
	})

	return manager
}

// AddConnection adds a new WebSocket connection
func (wsm *WebSocketManager) AddConnection(conn *WebSocketConnection) {
	wsm.connMutex.Lock()
	defer wsm.connMutex.Unlock()

	wsm.connections[conn.ID] = conn

	wsm.logger.InfoFields("WebSocket connection added", map[string]any{
		"connection_id":     conn.ID,
		"remote_addr":       conn.RemoteAddr,
		"total_connections": len(wsm.connections),
	})
}

// RemoveConnection removes a WebSocket connection
func (wsm *WebSocketManager) RemoveConnection(connectionID string) {
	wsm.connMutex.Lock()
	defer wsm.connMutex.Unlock()

	if conn, exists := wsm.connections[connectionID]; exists {
		conn.IsAlive = false
		if conn.Send != nil {
			close(conn.Send)
		}
		delete(wsm.connections, connectionID)

		wsm.logger.InfoFields("WebSocket connection removed", map[string]any{
			"connection_id":     connectionID,
			"remote_addr":       conn.RemoteAddr,
			"duration":          time.Since(conn.ConnectedAt).String(),
			"total_connections": len(wsm.connections),
		})
	}
}

// BroadcastUpdate broadcasts an update to all connected clients
func (wsm *WebSocketManager) BroadcastUpdate(updateType string, data interface{}) {
	message := &UpdateMessage{
		Type:      updateType,
		Timestamp: time.Now(),
		Data:      data,
	}

	select {
	case wsm.updateChan <- message:
		wsm.logger.DebugFields("Update queued for broadcast", map[string]any{
			"type":        updateType,
			"connections": len(wsm.connections),
		})
	default:
		wsm.logger.WarnFields("Update channel full, dropping message", map[string]any{
			"type":             updateType,
			"channel_capacity": cap(wsm.updateChan),
		})
	}
}

// GetConnectionCount returns the number of active connections
func (wsm *WebSocketManager) GetConnectionCount() int {
	wsm.connMutex.RLock()
	defer wsm.connMutex.RUnlock()
	return len(wsm.connections)
}

// GetConnectionStats returns connection statistics
func (wsm *WebSocketManager) GetConnectionStats() map[string]interface{} {
	wsm.connMutex.RLock()
	defer wsm.connMutex.RUnlock()

	stats := map[string]interface{}{
		"total_connections":     len(wsm.connections),
		"update_queue_size":     len(wsm.updateChan),
		"update_queue_capacity": cap(wsm.updateChan),
	}

	if len(wsm.connections) > 0 {
		var totalDuration time.Duration
		for _, conn := range wsm.connections {
			totalDuration += time.Since(conn.ConnectedAt)
		}
		stats["average_connection_duration"] = totalDuration / time.Duration(len(wsm.connections))
	}

	return stats
}

// Shutdown gracefully shuts down the WebSocket manager
func (wsm *WebSocketManager) Shutdown() {
	wsm.logger.Info("Shutting down WebSocket manager")

	// Cancel context to stop goroutines
	wsm.cancel()

	// Close all connections
	wsm.connMutex.Lock()
	for id, conn := range wsm.connections {
		conn.IsAlive = false
		if conn.Conn != nil {
			conn.Conn.Close()
		}
		if conn.Send != nil {
			close(conn.Send)
		}
		delete(wsm.connections, id)
	}
	wsm.connMutex.Unlock()

	// Close update channel
	close(wsm.updateChan)

	wsm.logger.Info("WebSocket manager shutdown complete")
} // startUpdateBroadcaster handles broadcasting updates to all connections
func (wsm *WebSocketManager) startUpdateBroadcaster() {
	wsm.logger.Debug("Starting update broadcaster")

	for {
		select {
		case <-wsm.ctx.Done():
			wsm.logger.Debug("Update broadcaster stopping")
			return
		case message := <-wsm.updateChan:
			wsm.broadcastToConnections(message)
		}
	}
}

// broadcastToConnections sends a message to all active connections
func (wsm *WebSocketManager) broadcastToConnections(message *UpdateMessage) {
	wsm.connMutex.RLock()
	connections := make([]*WebSocketConnection, 0, len(wsm.connections))
	for _, conn := range wsm.connections {
		if conn.IsAlive {
			connections = append(connections, conn)
		}
	}
	wsm.connMutex.RUnlock()

	if len(connections) == 0 {
		return
	}

	data, err := json.Marshal(message)
	if err != nil {
		wsm.logger.ErrorFields("Failed to marshal update message", map[string]any{
			"error": err.Error(),
			"type":  message.Type,
		})
		return
	}

	wsm.logger.DebugFields("Broadcasting update", map[string]any{
		"type":        message.Type,
		"connections": len(connections),
		"data_size":   len(data),
	})

	// Send to all connections in parallel
	for _, conn := range connections {
		go func(c *WebSocketConnection) {
			select {
			case c.Send <- data:
				// Message sent successfully
			default:
				// Connection send buffer is full, remove connection
				wsm.logger.WarnFields("Connection send buffer full, removing connection", map[string]any{
					"connection_id": c.ID,
					"remote_addr":   c.RemoteAddr,
				})
				wsm.RemoveConnection(c.ID)
			}
		}(conn)
	}
}

// startConnectionCleanup periodically checks and cleans up dead connections
func (wsm *WebSocketManager) startConnectionCleanup(pingInterval time.Duration) {
	wsm.logger.Debug("Starting connection cleanup routine")

	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-wsm.ctx.Done():
			wsm.logger.Debug("Connection cleanup stopping")
			return
		case <-ticker.C:
			wsm.cleanupConnections()
		}
	}
}

// cleanupConnections removes dead connections and sends ping to active ones
func (wsm *WebSocketManager) cleanupConnections() {
	wsm.connMutex.Lock()
	defer wsm.connMutex.Unlock()

	now := time.Now()
	deadConnections := make([]string, 0)

	for id, conn := range wsm.connections {
		if !conn.IsAlive || now.Sub(conn.LastPing) > 90*time.Second {
			deadConnections = append(deadConnections, id)
			continue
		}

		// Send ping to check if connection is still alive
		if err := conn.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
			wsm.logger.DebugFields("Failed to ping connection", map[string]any{
				"connection_id": id,
				"error":         err.Error(),
			})
			deadConnections = append(deadConnections, id)
		} else {
			conn.LastPing = now
		}
	}

	// Remove dead connections
	for _, id := range deadConnections {
		if conn, exists := wsm.connections[id]; exists {
			conn.IsAlive = false
			conn.Conn.Close()
			close(conn.Send)
			delete(wsm.connections, id)

			wsm.logger.DebugFields("Cleaned up dead connection", map[string]any{
				"connection_id": id,
				"remote_addr":   conn.RemoteAddr,
			})
		}
	}

	if len(deadConnections) > 0 {
		wsm.logger.InfoFields("Connection cleanup completed", map[string]any{
			"removed_connections": len(deadConnections),
			"active_connections":  len(wsm.connections),
		})
	}
}
