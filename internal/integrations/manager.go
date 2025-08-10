package integrations

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"pihole-analyzer/internal/integrations/grafana"
	"pihole-analyzer/internal/integrations/interfaces"
	"pihole-analyzer/internal/integrations/loki"
	"pihole-analyzer/internal/integrations/prometheus"
	"pihole-analyzer/internal/types"
)

// Manager implements the IntegrationManager interface
type Manager struct {
	config       *types.IntegrationsConfig
	logger       *slog.Logger
	integrations map[string]interfaces.MonitoringIntegration
	mu           sync.RWMutex
	initialized  bool
}

// NewManager creates a new integration manager
func NewManager(logger *slog.Logger) *Manager {
	return &Manager{
		logger:       logger,
		integrations: make(map[string]interfaces.MonitoringIntegration),
		mu:           sync.RWMutex{},
		initialized:  false,
	}
}

// Initialize sets up all enabled integrations
func (m *Manager) Initialize(ctx context.Context, config *types.IntegrationsConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = config
	
	if !config.Enabled {
		m.logger.Info("🔌 Integrations disabled in configuration",
			slog.String("component", "integrations"))
		return nil
	}

	m.logger.Info("🚀 Initializing monitoring integrations",
		slog.String("component", "integrations"))

	// Initialize each integration type
	if err := m.initializeGrafana(ctx); err != nil {
		m.logger.Error("❌ Failed to initialize Grafana integration",
			slog.String("component", "integrations"),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to initialize Grafana: %w", err)
	}

	if err := m.initializeLoki(ctx); err != nil {
		m.logger.Error("❌ Failed to initialize Loki integration",
			slog.String("component", "integrations"),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to initialize Loki: %w", err)
	}

	if err := m.initializePrometheus(ctx); err != nil {
		m.logger.Error("❌ Failed to initialize Prometheus integration",
			slog.String("component", "integrations"),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to initialize Prometheus: %w", err)
	}

	if err := m.initializeGeneric(ctx); err != nil {
		m.logger.Error("❌ Failed to initialize generic integrations",
			slog.String("component", "integrations"),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to initialize generic integrations: %w", err)
	}

	m.initialized = true
	
	// Count enabled integrations manually to avoid deadlock
	enabledCount := 0
	for _, integration := range m.integrations {
		if integration.IsEnabled() {
			enabledCount++
		}
	}
	
	m.logger.Info("✅ Monitoring integrations initialized successfully",
		slog.String("component", "integrations"),
		slog.Int("enabled_count", enabledCount))

	return nil
}

// RegisterIntegration registers a new integration
func (m *Manager) RegisterIntegration(integration interfaces.MonitoringIntegration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := integration.GetName()
	if _, exists := m.integrations[name]; exists {
		return fmt.Errorf("integration %s already registered", name)
	}

	m.integrations[name] = integration
	
	m.logger.Info("📋 Registered monitoring integration",
		slog.String("component", "integrations"),
		slog.String("integration", name))

	return nil
}

// GetIntegration retrieves an integration by name
func (m *Manager) GetIntegration(name string) (interfaces.MonitoringIntegration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	integration, exists := m.integrations[name]
	if !exists {
		return nil, fmt.Errorf("integration %s not found", name)
	}

	return integration, nil
}

// GetEnabledIntegrations returns all enabled integrations
func (m *Manager) GetEnabledIntegrations() []interfaces.MonitoringIntegration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var enabled []interfaces.MonitoringIntegration
	for _, integration := range m.integrations {
		if integration.IsEnabled() {
			enabled = append(enabled, integration)
		}
	}

	return enabled
}

// SendToAll sends data to all enabled integrations
func (m *Manager) SendToAll(ctx context.Context, data *types.AnalysisResult) error {
	if !m.initialized || !m.config.Enabled {
		return nil
	}

	enabled := m.GetEnabledIntegrations()
	if len(enabled) == 0 {
		return nil
	}

	m.logger.Debug("📤 Sending data to all enabled integrations",
		slog.String("component", "integrations"),
		slog.Int("integration_count", len(enabled)))

	// Send data to all integrations concurrently
	errChan := make(chan error, len(enabled))
	var wg sync.WaitGroup

	for _, integration := range enabled {
		wg.Add(1)
		go func(integ interfaces.MonitoringIntegration) {
			defer wg.Done()

			start := time.Now()
			if err := integ.SendMetrics(ctx, data); err != nil {
				m.logger.Error("❌ Failed to send metrics to integration",
					slog.String("component", "integrations"),
					slog.String("integration", integ.GetName()),
					slog.String("error", err.Error()))
				errChan <- fmt.Errorf("integration %s: %w", integ.GetName(), err)
				return
			}

			m.logger.Debug("✅ Successfully sent metrics to integration",
				slog.String("component", "integrations"),
				slog.String("integration", integ.GetName()),
				slog.Duration("duration", time.Since(start)))
		}(integration)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Collect any errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send to %d integrations: %v", len(errors), errors)
	}

	return nil
}

// TestAll tests all enabled integrations
func (m *Manager) TestAll(ctx context.Context) map[string]error {
	results := make(map[string]error)
	
	if !m.initialized || !m.config.Enabled {
		return results
	}

	enabled := m.GetEnabledIntegrations()
	m.logger.Info("🧪 Testing all enabled integrations",
		slog.String("component", "integrations"),
		slog.Int("integration_count", len(enabled)))

	for _, integration := range enabled {
		name := integration.GetName()
		if err := integration.TestConnection(ctx); err != nil {
			results[name] = err
			m.logger.Error("❌ Integration test failed",
				slog.String("component", "integrations"),
				slog.String("integration", name),
				slog.String("error", err.Error()))
		} else {
			m.logger.Info("✅ Integration test passed",
				slog.String("component", "integrations"),
				slog.String("integration", name))
		}
	}

	return results
}

// Close closes all integrations
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return nil
	}

	m.logger.Info("🔌 Closing all monitoring integrations",
		slog.String("component", "integrations"))

	var errors []error
	for name, integration := range m.integrations {
		if err := integration.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close %s: %w", name, err))
			m.logger.Error("❌ Failed to close integration",
				slog.String("component", "integrations"),
				slog.String("integration", name),
				slog.String("error", err.Error()))
		} else {
			m.logger.Debug("✅ Successfully closed integration",
				slog.String("component", "integrations"),
				slog.String("integration", name))
		}
	}

	m.initialized = false

	if len(errors) > 0 {
		return fmt.Errorf("failed to close some integrations: %v", errors)
	}

	m.logger.Info("✅ All monitoring integrations closed successfully",
		slog.String("component", "integrations"))
	return nil
}

// initializeGrafana initializes the Grafana integration if enabled
func (m *Manager) initializeGrafana(ctx context.Context) error {
	if !m.config.Grafana.Enabled {
		m.logger.Debug("⏭️ Grafana integration disabled",
			slog.String("component", "integrations"))
		return nil
	}

	// Import and create Grafana client
	grafanaClient := grafana.NewClient(&m.config.Grafana, m.logger)
	
	if err := grafanaClient.Initialize(ctx, &m.config.Grafana); err != nil {
		return fmt.Errorf("failed to initialize Grafana client: %w", err)
	}

	return m.RegisterIntegration(grafanaClient)
}

// initializeLoki initializes the Loki integration if enabled
func (m *Manager) initializeLoki(ctx context.Context) error {
	if !m.config.Loki.Enabled {
		m.logger.Debug("⏭️ Loki integration disabled",
			slog.String("component", "integrations"))
		return nil
	}

	// Import and create Loki client
	lokiClient := loki.NewClient(&m.config.Loki, m.logger)
	
	if err := lokiClient.Initialize(ctx, &m.config.Loki); err != nil {
		return fmt.Errorf("failed to initialize Loki client: %w", err)
	}

	return m.RegisterIntegration(lokiClient)
}

// initializePrometheus initializes the Prometheus integration if enabled
func (m *Manager) initializePrometheus(ctx context.Context) error {
	if !m.config.Prometheus.Enabled {
		m.logger.Debug("⏭️ Prometheus extended integration disabled",
			slog.String("component", "integrations"))
		return nil
	}

	// Import and create Prometheus client
	prometheusClient := prometheus.NewClient(&m.config.Prometheus, m.logger)
	
	if err := prometheusClient.Initialize(ctx, &m.config.Prometheus); err != nil {
		return fmt.Errorf("failed to initialize Prometheus client: %w", err)
	}

	return m.RegisterIntegration(prometheusClient)
}

// initializeGeneric initializes generic integrations
func (m *Manager) initializeGeneric(ctx context.Context) error {
	if len(m.config.Generic) == 0 {
		m.logger.Debug("⏭️ No generic integrations configured",
			slog.String("component", "integrations"))
		return nil
	}

	m.logger.Info("🔧 Generic integrations will be implemented",
		slog.String("component", "integrations"),
		slog.Int("count", len(m.config.Generic)))
	
	return nil
}

// GetStatus returns the status of all integrations
func (m *Manager) GetStatus() map[string]interfaces.IntegrationStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]interfaces.IntegrationStatus)
	for name, integration := range m.integrations {
		status[name] = integration.GetStatus()
	}

	return status
}

// SendLogs sends logs to all enabled log integrations
func (m *Manager) SendLogs(ctx context.Context, logs []interfaces.LogEntry) error {
	if !m.initialized || !m.config.Enabled {
		return nil
	}

	enabled := m.GetEnabledIntegrations()
	var logIntegrations []interfaces.LogsIntegration

	// Filter for log integrations
	for _, integration := range enabled {
		if logInteg, ok := integration.(interfaces.LogsIntegration); ok {
			logIntegrations = append(logIntegrations, logInteg)
		}
	}

	if len(logIntegrations) == 0 {
		return nil
	}

	m.logger.Debug("📝 Sending logs to enabled log integrations",
		slog.String("component", "integrations"),
		slog.Int("integration_count", len(logIntegrations)),
		slog.Int("log_count", len(logs)))

	// Send logs to all log integrations concurrently
	errChan := make(chan error, len(logIntegrations))
	var wg sync.WaitGroup

	for _, integration := range logIntegrations {
		wg.Add(1)
		go func(integ interfaces.LogsIntegration) {
			defer wg.Done()

			if err := integ.SendLogs(ctx, logs); err != nil {
				errChan <- fmt.Errorf("log integration error: %w", err)
			}
		}(integration)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Collect any errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send logs to %d integrations: %v", len(errors), errors)
	}

	return nil
}

// IsInitialized returns whether the manager has been initialized
func (m *Manager) IsInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.initialized
}

// GetConfig returns the current integrations configuration
func (m *Manager) GetConfig() *types.IntegrationsConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// Compile-time interface check
var _ interfaces.IntegrationManager = (*Manager)(nil)