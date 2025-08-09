package metrics

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server represents the metrics HTTP server
type Server struct {
	server    *http.Server
	collector *Collector
	logger    *slog.Logger
}

// ServerConfig holds configuration for the metrics server
type ServerConfig struct {
	Port    string
	Host    string
	Enabled bool
}

// NewServer creates a new metrics HTTP server
func NewServer(config ServerConfig, collector *Collector, logger *slog.Logger) *Server {
	mux := http.NewServeMux()

	// Prometheus metrics endpoint using the collector's registry
	mux.Handle("/metrics", promhttp.HandlerFor(collector.GetRegistry(), promhttp.HandlerOpts{}))

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","component":"metrics-server"}`))
	})

	// Root endpoint with information
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Pi-hole Network Analyzer Metrics</title>
</head>
<body>
    <h1>Pi-hole Network Analyzer Metrics Server</h1>
    <p>Available endpoints:</p>
    <ul>
        <li><a href="/metrics">/metrics</a> - Prometheus metrics</li>
        <li><a href="/health">/health</a> - Health check</li>
    </ul>
</body>
</html>
		`))
	})

	addr := config.Host + ":" + config.Port
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		server:    server,
		collector: collector,
		logger:    logger,
	}
}

// Start starts the metrics HTTP server
func (s *Server) Start() error {
	s.logger.Info("🚀 Starting metrics server",
		slog.String("component", "metrics-server"),
		slog.String("address", s.server.Addr))

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error("❌ Failed to start metrics server",
			slog.String("component", "metrics-server"),
			slog.String("error", err.Error()))
		return err
	}

	return nil
}

// StartInBackground starts the metrics server in a background goroutine
func (s *Server) StartInBackground() {
	go func() {
		if err := s.Start(); err != nil {
			s.logger.Error("❌ Metrics server error",
				slog.String("component", "metrics-server"),
				slog.String("error", err.Error()))
		}
	}()

	s.logger.Info("📡 Metrics server started in background",
		slog.String("component", "metrics-server"),
		slog.String("address", s.server.Addr))
}

// Stop gracefully stops the metrics server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("🛑 Stopping metrics server",
		slog.String("component", "metrics-server"))

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("❌ Error stopping metrics server",
			slog.String("component", "metrics-server"),
			slog.String("error", err.Error()))
		return err
	}

	s.logger.Info("✅ Metrics server stopped gracefully",
		slog.String("component", "metrics-server"))

	return nil
}

// GetCollector returns the metrics collector
func (s *Server) GetCollector() *Collector {
	return s.collector
}
