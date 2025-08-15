package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

//go:embed templates/*
var templatesFS embed.FS

// Server represents the web server for Pi-hole analysis
type Server struct {
	logger     *logger.Logger
	config     *Config
	dataSource DataSourceProvider
	server     *http.Server
	templates  *template.Template
	wsManager  *WebSocketManager
}

// Config holds web server configuration
type Config struct {
	Port            int
	Host            string
	EnableWeb       bool
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	EnableWebSocket bool
	WebSocketConfig *WebSocketConfig
}

// DataSourceProvider interface for getting analysis data
type DataSourceProvider interface {
	GetAnalysisResult(ctx context.Context) (*types.AnalysisResult, error)
	GetConnectionStatus() *types.ConnectionStatus
}

// DefaultConfig returns default web server configuration
func DefaultConfig() *Config {
	return &Config{
		Port:            8080,
		Host:            "localhost",
		EnableWeb:       false,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     60 * time.Second,
		EnableWebSocket: true,
		WebSocketConfig: DefaultWebSocketConfig(),
	}
}

// NewServer creates a new web server instance
func NewServer(config *Config, dataSource DataSourceProvider, logger *logger.Logger) (*Server, error) {
	if config == nil {
		config = DefaultConfig()
	}

	webLogger := logger.Component("web-server")

	// Parse embedded templates
	templates, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		webLogger.Error("Failed to parse templates: %v", err)
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	server := &Server{
		logger:     webLogger,
		config:     config,
		dataSource: dataSource,
		templates:  templates,
	}

	// Initialize WebSocket manager if enabled
	if config.EnableWebSocket {
		server.wsManager = NewWebSocketManager(webLogger, config.WebSocketConfig)
	}

	// Setup HTTP server
	mux := http.NewServeMux()
	server.setupRoutes(mux)

	server.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:      server.loggingMiddleware(mux),
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	webLogger.InfoFields("Web server configured", map[string]any{
		"host": config.Host,
		"port": config.Port,
	})

	return server, nil
}

// Start starts the web server
func (s *Server) Start(ctx context.Context) error {
	s.logger.Success("Starting web server: %s", s.server.Addr)

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		s.logger.Info("Shutting down web server...")
		return s.Stop()
	case err := <-errChan:
		s.logger.Error("Web server error: %v", err)
		return err
	}
}

// Stop gracefully stops the web server
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.logger.Info("Stopping web server...")

	// Shutdown WebSocket manager first
	if s.wsManager != nil {
		s.wsManager.Shutdown()
	}

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("Error stopping web server: %v", err)
		return err
	}

	s.logger.Success("Web server stopped successfully")
	return nil
}

// setupRoutes configures HTTP routes
func (s *Server) setupRoutes(mux *http.ServeMux) {
<<<<<<< HEAD
=======
	// Serve static files
	fs := http.FileServer(http.Dir("internal/web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
>>>>>>> main
	// Main dashboard (original)
	mux.HandleFunc("/", s.handleDashboard)

	// Enhanced dashboard
	mux.HandleFunc("/enhanced", s.handleEnhancedDashboard)

	// WebSocket endpoint
	if s.wsManager != nil {
		mux.HandleFunc("/ws", s.HandleWebSocket)
	}

	// API endpoints
	mux.HandleFunc("/api/status", s.handleAPIStatus)
	mux.HandleFunc("/api/analysis", s.handleAPIAnalysis)
	mux.HandleFunc("/api/clients", s.handleAPIClients)

	// Chart API endpoints
	chartHandler := NewChartAPIHandler(s.dataSource, s.logger)
	mux.HandleFunc("/api/charts/timeline", chartHandler.HandleTimelineChart)
	mux.HandleFunc("/api/charts/clients", chartHandler.HandleClientChart)
	mux.HandleFunc("/api/charts/domains", chartHandler.HandleDomainChart)
	mux.HandleFunc("/api/charts/topology", chartHandler.HandleTopologyChart)
	mux.HandleFunc("/api/charts/performance", chartHandler.HandlePerformanceChart)

	// Health check
	mux.HandleFunc("/health", s.handleHealth)

	s.logger.Debug("HTTP routes configured")
}

// loggingMiddleware adds request logging
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapper := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		// Process request
		next.ServeHTTP(wrapper, r)

		// Log request
		duration := time.Since(start)
		s.logger.InfoFields("HTTP request", map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status_code": wrapper.statusCode,
			"duration_ms": duration.Milliseconds(),
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
		})
	})
}

// responseWrapper captures HTTP response status code
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// handleDashboard serves the main dashboard page
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Get analysis data
	analysisResult, err := s.dataSource.GetAnalysisResult(ctx)
	if err != nil {
		s.logger.Error("Failed to get analysis result: %v", err)
		http.Error(w, "Failed to load analysis data", http.StatusInternalServerError)
		return
	}

	// Get connection status
	connectionStatus := s.dataSource.GetConnectionStatus()

	// Template data
	data := struct {
		Title            string
		AnalysisResult   *types.AnalysisResult
		ConnectionStatus *types.ConnectionStatus
		ServerInfo       map[string]string
	}{
		Title:            "Pi-hole Network Analyzer",
		AnalysisResult:   analysisResult,
		ConnectionStatus: connectionStatus,
		ServerInfo: map[string]string{
			"Version":   "1.0.0",
			"BuildTime": time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		s.logger.Error("Failed to execute template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	s.logger.Debug("Dashboard served successfully")
}

// handleAPIStatus returns connection status as JSON
func (s *Server) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := s.dataSource.GetConnectionStatus()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		s.logger.Error("Failed to encode status response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleAPIAnalysis returns analysis data as JSON
func (s *Server) handleAPIAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	analysisResult, err := s.dataSource.GetAnalysisResult(ctx)
	if err != nil {
		s.logger.Error("Failed to get analysis result for API: %v", err)
		http.Error(w, "Failed to load analysis data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(analysisResult); err != nil {
		s.logger.Error("Failed to encode analysis response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleAPIClients returns client data as JSON
func (s *Server) handleAPIClients(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	analysisResult, err := s.dataSource.GetAnalysisResult(ctx)
	if err != nil {
		s.logger.Error("Failed to get client data for API: %v", err)
		http.Error(w, "Failed to load client data", http.StatusInternalServerError)
		return
	}

	// Parse query parameters for filtering
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Extract and limit client data
	clients := make([]*types.ClientStats, 0, len(analysisResult.ClientStats))
	count := 0
	for _, client := range analysisResult.ClientStats {
		if count >= limit {
			break
		}
		clients = append(clients, client)
		count++
	}

	response := struct {
		Clients      []*types.ClientStats `json:"clients"`
		TotalCount   int                  `json:"total_count"`
		LimitApplied int                  `json:"limit_applied"`
	}{
		Clients:      clients,
		TotalCount:   len(analysisResult.ClientStats),
		LimitApplied: limit,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to encode clients response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleHealth returns health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	connectionStatus := s.dataSource.GetConnectionStatus()

	health := struct {
		Status     string `json:"status"`
		Timestamp  string `json:"timestamp"`
		Connected  bool   `json:"connected"`
		ServerInfo string `json:"server_info"`
	}{
		Status:     "healthy",
		Timestamp:  time.Now().Format(time.RFC3339),
		Connected:  connectionStatus.Connected,
		ServerInfo: fmt.Sprintf("Pi-hole Network Analyzer Web UI (Port %d)", s.config.Port),
	}

	if !connectionStatus.Connected {
		health.Status = "degraded"
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(health); err != nil {
		s.logger.Error("Failed to encode health response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleEnhancedDashboard serves the enhanced dashboard page with interactive charts
func (s *Server) handleEnhancedDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Get analysis data
	analysisResult, err := s.dataSource.GetAnalysisResult(ctx)
	if err != nil {
		s.logger.Error("Failed to get analysis result: %v", err)
		http.Error(w, "Failed to load analysis data", http.StatusInternalServerError)
		return
	}

	// Get connection status
	connectionStatus := s.dataSource.GetConnectionStatus()

	// Template data
	data := struct {
		Title            string
		AnalysisResult   *types.AnalysisResult
		ConnectionStatus *types.ConnectionStatus
		ServerInfo       map[string]string
	}{
		Title:            "Pi-hole Network Analyzer - Enhanced Dashboard",
		AnalysisResult:   analysisResult,
		ConnectionStatus: connectionStatus,
		ServerInfo: map[string]string{
			"Version":   "1.0.0",
			"BuildTime": time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "dashboard_enhanced.html", data); err != nil {
		s.logger.Error("Failed to execute enhanced dashboard template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	s.logger.Debug("Enhanced dashboard served successfully")
}
