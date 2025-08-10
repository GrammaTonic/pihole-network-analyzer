package web

import (
	"encoding/json"
	"testing"
	"time"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

func TestNewWebSocketManager(t *testing.T) {
	logger := logger.New(&logger.Config{Component: "test"})
	config := DefaultWebSocketConfig()

	manager := NewWebSocketManager(logger, config)

	if manager == nil {
		t.Fatal("Expected WebSocket manager to be created")
	}

	if manager.connections == nil {
		t.Error("Expected connections map to be initialized")
	}

	if manager.updateChan == nil {
		t.Error("Expected update channel to be initialized")
	}

	// Check configuration
	if cap(manager.updateChan) != 1000 {
		t.Errorf("Expected update channel capacity to be 1000, got %d", cap(manager.updateChan))
	}

	// Cleanup
	manager.Shutdown()
}

func TestWebSocketManager_GetConnectionCount(t *testing.T) {
	logger := logger.New(&logger.Config{Component: "test"})
	config := DefaultWebSocketConfig()
	manager := NewWebSocketManager(logger, config)
	defer manager.Shutdown()

	// Initially should have 0 connections
	if count := manager.GetConnectionCount(); count != 0 {
		t.Errorf("Expected 0 connections, got %d", count)
	}

	// Add a mock connection (don't close it since it's nil)
	mockConn := &WebSocketConnection{
		ID:      "test-123",
		IsAlive: true,
		Send:    make(chan []byte, 256),
	}
	manager.AddConnection(mockConn)

	if count := manager.GetConnectionCount(); count != 1 {
		t.Errorf("Expected 1 connection, got %d", count)
	}

	// Remove connection without closing the underlying connection
	manager.connMutex.Lock()
	if conn, exists := manager.connections["test-123"]; exists {
		conn.IsAlive = false
		// Don't close Send channel here since we'll do cleanup
		delete(manager.connections, "test-123")
	}
	manager.connMutex.Unlock()

	if count := manager.GetConnectionCount(); count != 0 {
		t.Errorf("Expected 0 connections after removal, got %d", count)
	}
}

func TestWebSocketManager_BroadcastUpdate(t *testing.T) {
	logger := logger.New(&logger.Config{Component: "test"})
	config := DefaultWebSocketConfig()
	manager := NewWebSocketManager(logger, config)
	defer manager.Shutdown()

	// Add a mock connection to ensure the broadcast actually processes the message
	mockConn := &WebSocketConnection{
		ID:      "test-broadcast",
		IsAlive: true,
		Send:    make(chan []byte, 256),
	}
	manager.AddConnection(mockConn)

	// Test broadcasting
	manager.BroadcastUpdate("test", map[string]string{"key": "value"})

	// Check that message was sent to the connection
	select {
	case data := <-mockConn.Send:
		var message UpdateMessage
		if err := json.Unmarshal(data, &message); err != nil {
			t.Fatalf("Failed to unmarshal broadcast message: %v", err)
		}
		if message.Type != "test" {
			t.Errorf("Expected message type 'test', got '%s'", message.Type)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected message to be sent to connection")
	}
}

func TestWebSocketManager_GetConnectionStats(t *testing.T) {
	logger := logger.New(&logger.Config{Component: "test"})
	config := DefaultWebSocketConfig()
	manager := NewWebSocketManager(logger, config)
	defer manager.Shutdown()

	stats := manager.GetConnectionStats()

	expectedKeys := []string{
		"total_connections",
		"update_queue_size",
		"update_queue_capacity",
	}

	for _, key := range expectedKeys {
		if _, exists := stats[key]; !exists {
			t.Errorf("Expected stats to contain key '%s'", key)
		}
	}

	if stats["total_connections"] != 0 {
		t.Errorf("Expected 0 total connections, got %v", stats["total_connections"])
	}

	if stats["update_queue_capacity"] != 1000 {
		t.Errorf("Expected update queue capacity 1000, got %v", stats["update_queue_capacity"])
	}
}

func TestWebSocketConnection_SendPong(t *testing.T) {
	logger := logger.New(&logger.Config{Component: "test"})
	config := DefaultWebSocketConfig()
	manager := NewWebSocketManager(logger, config)
	defer manager.Shutdown()

	conn := &WebSocketConnection{
		ID:      "test-pong",
		Send:    make(chan []byte, 256),
		Manager: manager,
		IsAlive: true,
	}

	// Test sending pong
	conn.sendPong()

	// Check that pong was sent
	select {
	case data := <-conn.Send:
		var response map[string]interface{}
		if err := json.Unmarshal(data, &response); err != nil {
			t.Fatalf("Failed to unmarshal pong response: %v", err)
		}

		if response["type"] != "pong" {
			t.Errorf("Expected response type 'pong', got '%v'", response["type"])
		}

		if _, exists := response["timestamp"]; !exists {
			t.Error("Expected pong response to contain timestamp")
		}

	case <-time.After(100 * time.Millisecond):
		t.Error("Expected pong response to be sent")
	}
}

func TestWebSocketConnection_HandleClientMessage(t *testing.T) {
	logger := logger.New(&logger.Config{Component: "test"})
	config := DefaultWebSocketConfig()
	manager := NewWebSocketManager(logger, config)
	defer manager.Shutdown()

	conn := &WebSocketConnection{
		ID:      "test-message",
		Send:    make(chan []byte, 256),
		Manager: manager,
		IsAlive: true,
	}

	// Test ping message
	pingMessage := map[string]interface{}{
		"type": "ping",
	}

	data, err := json.Marshal(pingMessage)
	if err != nil {
		t.Fatalf("Failed to marshal ping message: %v", err)
	}

	conn.handleClientMessage(data)

	// Check that pong was sent in response
	select {
	case responseData := <-conn.Send:
		var response map[string]interface{}
		if err := json.Unmarshal(responseData, &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["type"] != "pong" {
			t.Errorf("Expected pong response, got '%v'", response["type"])
		}

	case <-time.After(100 * time.Millisecond):
		t.Error("Expected pong response to ping")
	}
}

func TestWebSocketConnection_HandleSubscription(t *testing.T) {
	logger := logger.New(&logger.Config{Component: "test"})
	config := DefaultWebSocketConfig()
	manager := NewWebSocketManager(logger, config)
	defer manager.Shutdown()

	conn := &WebSocketConnection{
		ID:      "test-subscription",
		Send:    make(chan []byte, 256),
		Manager: manager,
		IsAlive: true,
	}

	// Test subscription message
	subscribeMessage := map[string]interface{}{
		"type":    "subscribe",
		"channel": "dashboard",
	}

	data, err := json.Marshal(subscribeMessage)
	if err != nil {
		t.Fatalf("Failed to marshal subscribe message: %v", err)
	}

	conn.handleClientMessage(data)

	// Check that subscription confirmation was sent
	select {
	case responseData := <-conn.Send:
		var response map[string]interface{}
		if err := json.Unmarshal(responseData, &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["type"] != "subscription_confirmed" {
			t.Errorf("Expected subscription_confirmed response, got '%v'", response["type"])
		}

		if response["channel"] != "dashboard" {
			t.Errorf("Expected channel 'dashboard', got '%v'", response["channel"])
		}

	case <-time.After(100 * time.Millisecond):
		t.Error("Expected subscription confirmation")
	}
}

func TestDefaultWebSocketConfig(t *testing.T) {
	config := DefaultWebSocketConfig()

	if config == nil {
		t.Fatal("Expected config to be created")
	}

	if !config.EnableWebSocket {
		t.Error("Expected WebSocket to be enabled by default")
	}

	if config.UpdateInterval != 2*time.Second {
		t.Errorf("Expected update interval 2s, got %v", config.UpdateInterval)
	}

	if config.PingInterval != 30*time.Second {
		t.Errorf("Expected ping interval 30s, got %v", config.PingInterval)
	}

	if config.MaxConnections != 100 {
		t.Errorf("Expected max connections 100, got %d", config.MaxConnections)
	}

	if config.BufferSize != 256 {
		t.Errorf("Expected buffer size 256, got %d", config.BufferSize)
	}

	if !config.EnableCompression {
		t.Error("Expected compression to be enabled by default")
	}
}

func TestUpdateMessage_Structure(t *testing.T) {
	now := time.Now()
	testData := map[string]string{"test": "data"}

	update := &UpdateMessage{
		Type:      "test_update",
		Timestamp: now,
		Data:      testData,
	}

	// Test JSON marshaling
	data, err := json.Marshal(update)
	if err != nil {
		t.Fatalf("Failed to marshal UpdateMessage: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled UpdateMessage
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal UpdateMessage: %v", err)
	}

	if unmarshaled.Type != "test_update" {
		t.Errorf("Expected type 'test_update', got '%s'", unmarshaled.Type)
	}

	// Note: Timestamp comparison might be tricky due to JSON precision
	if unmarshaled.Timestamp.Unix() != now.Unix() {
		t.Errorf("Expected timestamp %v, got %v", now.Unix(), unmarshaled.Timestamp.Unix())
	}
}

func TestDashboardSummary_Structure(t *testing.T) {
	summary := &DashboardSummary{
		TotalQueries:      100,
		TotalClients:      10,
		OnlineClients:     8,
		TopDomain:         "example.com",
		QueryRate:         5.5,
		LastUpdateTime:    time.Now(),
		AverageResponse:   15.2,
		BlockedQueries:    25,
		BlockedPercentage: 25.0,
	}

	// Test JSON marshaling
	data, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("Failed to marshal DashboardSummary: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled DashboardSummary
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal DashboardSummary: %v", err)
	}

	if unmarshaled.TotalQueries != 100 {
		t.Errorf("Expected TotalQueries 100, got %d", unmarshaled.TotalQueries)
	}

	if unmarshaled.TopDomain != "example.com" {
		t.Errorf("Expected TopDomain 'example.com', got '%s'", unmarshaled.TopDomain)
	}

	if unmarshaled.BlockedPercentage != 25.0 {
		t.Errorf("Expected BlockedPercentage 25.0, got %f", unmarshaled.BlockedPercentage)
	}
}

func TestServer_CreateDashboardSummary(t *testing.T) {
	logger := logger.New(&logger.Config{Component: "test"})
	config := DefaultConfig()
	dataSource := &MockDataSourceProvider{}

	server, err := NewServer(config, dataSource, logger)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Stop()

	// Test with nil result
	summary := server.createDashboardSummary(nil)
	if summary == nil {
		t.Fatal("Expected summary to be created even with nil result")
	}

	// Test with empty result
	result := &types.AnalysisResult{
		ClientStats:   make(map[string]*types.ClientStats),
		TotalQueries:  50,
		UniqueClients: 5,
	}

	summary = server.createDashboardSummary(result)
	if summary.TotalQueries != 50 {
		t.Errorf("Expected TotalQueries 50, got %d", summary.TotalQueries)
	}

	if summary.TotalClients != 5 {
		t.Errorf("Expected TotalClients 5, got %d", summary.TotalClients)
	}

	if summary.OnlineClients != 0 {
		t.Errorf("Expected OnlineClients 0, got %d", summary.OnlineClients)
	}

	// Test with clients data
	result.ClientStats["192.168.1.100"] = &types.ClientStats{
		IP:       "192.168.1.100",
		IsOnline: true,
		TopDomains: []types.DomainStat{
			{Domain: "google.com", Count: 10},
		},
	}

	result.ClientStats["192.168.1.101"] = &types.ClientStats{
		IP:       "192.168.1.101",
		IsOnline: false,
	}

	summary = server.createDashboardSummary(result)
	if summary.OnlineClients != 1 {
		t.Errorf("Expected OnlineClients 1, got %d", summary.OnlineClients)
	}

	if summary.TopDomain != "google.com" {
		t.Errorf("Expected TopDomain 'google.com', got '%s'", summary.TopDomain)
	}
}
