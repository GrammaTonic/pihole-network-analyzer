package pihole

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pihole-analyzer/internal/logger"
)

// MockPiholeAPI creates a mock Pi-hole API server for testing
type MockPiholeAPI struct {
	server   *httptest.Server
	sessions map[string]bool // Track active sessions
}

// NewMockPiholeAPI creates a new mock Pi-hole API server
func NewMockPiholeAPI() *MockPiholeAPI {
	mock := &MockPiholeAPI{
		sessions: make(map[string]bool),
	}

	mux := http.NewServeMux()

	// Authentication endpoint
	mux.HandleFunc("/api/auth", mock.handleAuth)

	// Queries endpoint
	mux.HandleFunc("/api/queries", mock.handleQueries)

	// Statistics endpoint
	mux.HandleFunc("/api/stats", mock.handleStats)

	// Network endpoint
	mux.HandleFunc("/api/network", mock.handleNetwork)

	mock.server = httptest.NewServer(mux)
	return mock
}

// Close shuts down the mock server
func (m *MockPiholeAPI) Close() {
	m.server.Close()
}

// URL returns the base URL of the mock server
func (m *MockPiholeAPI) URL() string {
	return m.server.URL
}

// handleAuth handles authentication requests
func (m *MockPiholeAPI) handleAuth(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Login
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Check password
		if payload["password"] != "test-password" {
			errorResp := ErrorResponse{
				Error: struct {
					Key     string      `json:"key"`
					Message string      `json:"message"`
					Hint    interface{} `json:"hint"`
				}{
					Key:     "unauthorized",
					Message: "Unauthorized",
					Hint:    nil,
				},
				Took: 0.001,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(errorResp)
			return
		}

		// Generate session
		sid := "test-session-id-12345"
		csrf := "test-csrf-token"
		m.sessions[sid] = true

		authResp := AuthResponse{
			Session: SessionInfo{
				SID:        sid,
				CSRFToken:  csrf,
				ValidUntil: time.Now().Add(time.Hour),
				Validity:   3600,
				TOTPReq:    false,
			},
			Took: 0.001,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(authResp)

	case "DELETE":
		// Logout
		sid := r.Header.Get("X-FTL-SID")
		if sid != "" {
			delete(m.sessions, sid)
		}
		w.WriteHeader(http.StatusGone)

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// handleQueries handles DNS queries requests
func (m *MockPiholeAPI) handleQueries(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	sid := r.Header.Get("X-FTL-SID")
	if !m.sessions[sid] {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Generate mock query data
	mockQueries := []APIQueryRecord{
		{
			Timestamp: float64(time.Now().Unix()),
			Type:      "A",
			Domain:    "example.com",
			Client:    "192.168.1.100",
			Status:    2, // Forwarded
			Reply:     0.023,
			DNSSEC:    "N/A",
		},
		{
			Timestamp: float64(time.Now().Unix() - 60),
			Type:      "AAAA",
			Domain:    "google.com",
			Client:    "192.168.1.101",
			Status:    3, // Cached
			Reply:     0.001,
			DNSSEC:    "N/A",
		},
		{
			Timestamp: float64(time.Now().Unix() - 120),
			Type:      "A",
			Domain:    "malware.example",
			Client:    "192.168.1.102",
			Status:    1, // Blocked
			Reply:     0.0,
			DNSSEC:    "N/A",
		},
	}

	resp := QueriesResponse{
		Data: mockQueries,
		Took: 0.005,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleStats handles statistics requests
func (m *MockPiholeAPI) handleStats(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	sid := r.Header.Get("X-FTL-SID")
	if !m.sessions[sid] {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := StatsResponse{
		DomainsBlocked:     145238,
		DNSQueriesToday:    2847,
		AdsBlockedToday:    312,
		AdsPercentageToday: 10.95,
		UniqueClients:      8,
		QueriesForwarded:   1823,
		QueriesCached:      712,
		TopClients: map[string]int{
			"192.168.1.100": 523,
			"192.168.1.101": 412,
			"192.168.1.102": 298,
		},
		TopQueries: map[string]int{
			"example.com":  145,
			"google.com":   98,
			"facebook.com": 67,
			"youtube.com":  54,
		},
		Took: 0.003,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleNetwork handles network information requests
func (m *MockPiholeAPI) handleNetwork(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	sid := r.Header.Get("X-FTL-SID")
	if !m.sessions[sid] {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	network := NetworkResponse{
		Network: []ClientInfo{
			{
				Name: "laptop",
				IP:   "192.168.1.100",
				MAC:  "aa:bb:cc:dd:ee:ff",
			},
			{
				Name: "phone",
				IP:   "192.168.1.101",
				MAC:  "11:22:33:44:55:66",
			},
			{
				Name: "tablet",
				IP:   "192.168.1.102",
				MAC:  "77:88:99:aa:bb:cc",
			},
		},
		Took: 0.002,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(network)
}

// Test functions

func TestClientAuthentication(t *testing.T) {
	mock := NewMockPiholeAPI()
	defer mock.Close()

	config := &Config{
		Host:      "localhost",
		Port:      8080, // This will be ignored, mock server URL used instead
		Password:  "test-password",
		UseHTTPS:  false,
		Timeout:   5 * time.Second,
		VerifyTLS: false,
	}

	log := logger.New(&logger.Config{
		Level:        logger.LevelInfo,
		EnableColors: false,
		EnableEmojis: false,
	})

	client := NewClient(config, log)
	// Override base URL to use mock server
	client.BaseURL = mock.URL()

	ctx := context.Background()

	// Test successful authentication
	err := client.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	if client.SID == "" {
		t.Error("Expected SID to be set after authentication")
	}

	if client.CSRFToken == "" {
		t.Error("Expected CSRF token to be set after authentication")
	}

	// Test logout
	err = client.Close(ctx)
	if err != nil {
		t.Errorf("Logout failed: %v", err)
	}
}

func TestClientAuthenticationFailure(t *testing.T) {
	mock := NewMockPiholeAPI()
	defer mock.Close()

	config := &Config{
		Host:      "localhost",
		Port:      8080,
		Password:  "wrong-password",
		UseHTTPS:  false,
		Timeout:   5 * time.Second,
		VerifyTLS: false,
	}

	log := logger.New(&logger.Config{
		Level:        logger.LevelError,
		EnableColors: false,
		EnableEmojis: false,
	})

	client := NewClient(config, log)
	client.BaseURL = mock.URL()

	ctx := context.Background()

	// Test failed authentication
	err := client.Authenticate(ctx)
	if err == nil {
		t.Error("Expected authentication to fail with wrong password")
	}

	if client.SID != "" {
		t.Error("Expected SID to be empty after failed authentication")
	}
}

func TestGetDNSQueries(t *testing.T) {
	mock := NewMockPiholeAPI()
	defer mock.Close()

	client := createAuthenticatedTestClient(t, mock)
	defer client.Close(context.Background())

	ctx := context.Background()
	params := QueryParams{
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   time.Now(),
		Limit:     100,
	}

	queries, err := client.GetDNSQueries(ctx, params)
	if err != nil {
		t.Fatalf("GetDNSQueries failed: %v", err)
	}

	if len(queries) == 0 {
		t.Error("Expected queries to be returned")
	}

	// Validate query structure
	for _, query := range queries {
		if query.Client == "" {
			t.Error("Expected client IP to be set")
		}
		if query.Domain == "" {
			t.Error("Expected domain to be set")
		}
		if query.Status == 0 {
			t.Error("Expected status to be set")
		}
	}
}

func TestGetStatistics(t *testing.T) {
	mock := NewMockPiholeAPI()
	defer mock.Close()

	client := createAuthenticatedTestClient(t, mock)
	defer client.Close(context.Background())

	ctx := context.Background()

	stats, err := client.GetStatistics(ctx)
	if err != nil {
		t.Fatalf("GetStatistics failed: %v", err)
	}

	if stats.DNSQueriesToday == 0 {
		t.Error("Expected DNS queries today to be greater than 0")
	}

	if len(stats.TopClients) == 0 {
		t.Error("Expected top clients data")
	}

	if len(stats.TopQueries) == 0 {
		t.Error("Expected top queries data")
	}
}

func TestGetNetworkInfo(t *testing.T) {
	mock := NewMockPiholeAPI()
	defer mock.Close()

	client := createAuthenticatedTestClient(t, mock)
	defer client.Close(context.Background())

	ctx := context.Background()

	networkInfo, err := client.GetNetworkInfo(ctx)
	if err != nil {
		t.Fatalf("GetNetworkInfo failed: %v", err)
	}

	if len(networkInfo) == 0 {
		t.Error("Expected network info to be returned")
	}

	// Validate network info structure
	for _, info := range networkInfo {
		if info.IP == "" {
			t.Error("Expected IP to be set")
		}
		if info.MAC == "" {
			t.Error("Expected MAC to be set")
		}
	}
}

func TestBuildClientStats(t *testing.T) {
	mock := NewMockPiholeAPI()
	defer mock.Close()

	client := createAuthenticatedTestClient(t, mock)
	defer client.Close(context.Background())

	ctx := context.Background()

	// Get test data
	queries, err := client.GetDNSQueries(ctx, QueryParams{
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   time.Now(),
		Limit:     100,
	})
	if err != nil {
		t.Fatalf("Failed to get queries: %v", err)
	}

	networkInfo, err := client.GetNetworkInfo(ctx)
	if err != nil {
		t.Fatalf("Failed to get network info: %v", err)
	}

	// Build client stats
	clientStats, err := client.BuildClientStats(ctx, queries, networkInfo)
	if err != nil {
		t.Fatalf("BuildClientStats failed: %v", err)
	}

	if len(clientStats) == 0 {
		t.Error("Expected client stats to be generated")
	}

	// Validate client stats structure
	for clientIP, stats := range clientStats {
		if stats.Client != clientIP {
			t.Errorf("Expected client IP %s to match stats client %s", clientIP, stats.Client)
		}
		if stats.TotalQueries == 0 {
			t.Error("Expected total queries to be greater than 0")
		}
		if len(stats.Domains) == 0 {
			t.Error("Expected domains map to have entries")
		}
		if len(stats.StatusCodes) == 0 {
			t.Error("Expected status codes map to have entries")
		}
	}
}

// Helper function to create an authenticated test client
func createAuthenticatedTestClient(t *testing.T, mock *MockPiholeAPI) *Client {
	config := &Config{
		Host:      "localhost",
		Port:      8080,
		Password:  "test-password",
		UseHTTPS:  false,
		Timeout:   5 * time.Second,
		VerifyTLS: false,
	}

	log := logger.New(&logger.Config{
		Level:        logger.LevelError, // Reduce log noise in tests
		EnableColors: false,
		EnableEmojis: false,
	})

	client := NewClient(config, log)
	client.BaseURL = mock.URL()

	ctx := context.Background()
	err := client.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Failed to authenticate test client: %v", err)
	}

	return client
}
