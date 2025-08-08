package interfaces

import (
	"context"
	"fmt"
	"time"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// MigrationMode represents different migration phases
type MigrationMode string

const (
	// MigrationModeSSHOnly uses only SSH (legacy mode)
	MigrationModeSSHOnly MigrationMode = "ssh-only"

	// MigrationModeAPIWithSSHFallback tries API first, falls back to SSH
	MigrationModeAPIWithSSHFallback MigrationMode = "api-first"

	// MigrationModeAPIWithWarnings uses API only but warns about SSH deprecation
	MigrationModeAPIWithWarnings MigrationMode = "api-only-warn"

	// MigrationModeAPIOnly uses API only (future default)
	MigrationModeAPIOnly MigrationMode = "api-only"

	// MigrationModeAuto automatically detects best mode (current default)
	MigrationModeAuto MigrationMode = "auto"
)

// MigrationStrategy defines the migration approach and timeline
type MigrationStrategy struct {
	CurrentMode   MigrationMode `json:"current_mode"`
	TargetMode    MigrationMode `json:"target_mode"`
	MigrationDate time.Time     `json:"migration_date"`
	WarningPeriod time.Duration `json:"warning_period"`
	ForceAPIOnly  bool          `json:"force_api_only"`
	SSHDeprecated bool          `json:"ssh_deprecated"`
}

// MigrationManager handles the transition from SSH to API
type MigrationManager struct {
	logger   *logger.Logger
	strategy *MigrationStrategy
	config   *types.Config
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(config *types.Config, logger *logger.Logger) *MigrationManager {
	strategy := &MigrationStrategy{
		CurrentMode:   MigrationModeAuto,
		TargetMode:    MigrationModeAPIOnly,
		MigrationDate: time.Now().AddDate(0, 6, 0), // 6 months from now
		WarningPeriod: 30 * 24 * time.Hour,         // 30 days
		ForceAPIOnly:  false,
		SSHDeprecated: false,
	}

	return &MigrationManager{
		logger:   logger.Component("migration-manager"),
		strategy: strategy,
		config:   config,
	}
}

// GetMigrationMode determines the appropriate migration mode based on configuration and date
func (m *MigrationManager) GetMigrationMode() MigrationMode {
	pihole := &m.config.Pihole

	// Check for explicit migration mode override
	if pihole.MigrationMode != "" {
		return MigrationMode(pihole.MigrationMode)
	}

	// Auto-detection based on configuration and timeline
	now := time.Now()

	// If we're past the migration date, force API-only
	if now.After(m.strategy.MigrationDate) {
		m.logger.Info("Migration date passed, using API-only mode")
		return MigrationModeAPIOnly
	}

	// If we're in warning period, use API with warnings
	if now.After(m.strategy.MigrationDate.Add(-m.strategy.WarningPeriod)) {
		m.logger.Warn("Approaching SSH deprecation date: %s", m.strategy.MigrationDate.Format("2006-01-02"))
		return MigrationModeAPIWithWarnings
	}

	// Default auto-detection logic
	if pihole.APIPassword != "" {
		return MigrationModeAPIWithSSHFallback
	}

	if pihole.Username != "" && (pihole.Password != "" || pihole.KeyPath != "") {
		return MigrationModeSSHOnly
	}

	// No valid configuration found
	m.logger.Error("No valid Pi-hole configuration found for SSH or API")
	return MigrationModeAPIOnly // Force API to encourage proper setup
}

// CreateDataSourceWithMigration creates a data source with migration strategy applied
func (m *MigrationManager) CreateDataSourceWithMigration(ctx context.Context) (DataSource, error) {
	mode := m.GetMigrationMode()
	m.logger.Info("Migration mode determined: %s", mode)

	switch mode {
	case MigrationModeSSHOnly:
		return m.createSSHOnlyDataSource(ctx)
	case MigrationModeAPIWithSSHFallback:
		return m.createAPIWithSSHFallback(ctx)
	case MigrationModeAPIWithWarnings:
		return m.createAPIWithWarnings(ctx)
	case MigrationModeAPIOnly:
		return m.createAPIOnlyDataSource(ctx)
	default:
		return nil, fmt.Errorf("unsupported migration mode: %s", mode)
	}
}

// createSSHOnlyDataSource creates SSH-only data source (legacy mode)
func (m *MigrationManager) createSSHOnlyDataSource(ctx context.Context) (DataSource, error) {
	m.logger.Info("Creating SSH-only data source (legacy mode)")
	m.logSSHDeprecationWarning()

	// Import SSH package and create SSH data source
	// Note: This would require importing ssh package, which creates import cycles
	// For now, return error to encourage API migration
	return nil, fmt.Errorf("SSH-only mode is deprecated, please configure API access")
}

// createAPIWithSSHFallback creates API data source with SSH fallback
func (m *MigrationManager) createAPIWithSSHFallback(ctx context.Context) (DataSource, error) {
	m.logger.Info("Creating API data source with SSH fallback capability")

	// Try API first
	apiDataSource, err := m.createAPIDataSource()
	if err == nil {
		// Test API connection
		if err := apiDataSource.Connect(ctx); err == nil {
			m.logger.Info("✅ API connection successful, using API mode")
			return apiDataSource, nil
		}
		m.logger.Warn("API connection failed, attempting SSH fallback: %v", err)
		apiDataSource.Close()
	}

	// Fallback to SSH
	m.logger.Info("Falling back to SSH mode")
	m.logSSHDeprecationWarning()
	return nil, fmt.Errorf("SSH fallback not implemented in migration manager (import cycle protection)")
}

// createAPIWithWarnings creates API-only data source with deprecation warnings
func (m *MigrationManager) createAPIWithWarnings(ctx context.Context) (DataSource, error) {
	m.logger.Info("Creating API-only data source with SSH deprecation warnings")
	m.logSSHDeprecationWarning()
	m.logMigrationTimeline()

	apiDataSource, err := m.createAPIDataSource()
	if err != nil {
		return nil, fmt.Errorf("API data source creation failed: %w", err)
	}

	if err := apiDataSource.Connect(ctx); err != nil {
		return nil, fmt.Errorf("API connection failed (SSH fallback disabled in migration mode): %w", err)
	}

	m.logger.Info("✅ API connection successful, SSH support will be removed soon")
	return apiDataSource, nil
}

// createAPIOnlyDataSource creates API-only data source (future default)
func (m *MigrationManager) createAPIOnlyDataSource(ctx context.Context) (DataSource, error) {
	m.logger.Info("Creating API-only data source (SSH support removed)")

	apiDataSource, err := m.createAPIDataSource()
	if err != nil {
		return nil, fmt.Errorf("API data source creation failed: %w", err)
	}

	if err := apiDataSource.Connect(ctx); err != nil {
		return nil, fmt.Errorf("API connection failed: %w", err)
	}

	m.logger.Info("✅ API connection successful")
	return apiDataSource, nil
}

// createAPIDataSource creates an API data source instance
func (m *MigrationManager) createAPIDataSource() (DataSource, error) {
	// This would import pihole package, causing import cycles
	// For now, return an interface that can be implemented by the concrete factory
	return nil, fmt.Errorf("API data source creation must be handled by concrete implementation")
}

// logSSHDeprecationWarning logs SSH deprecation warnings
func (m *MigrationManager) logSSHDeprecationWarning() {
	m.logger.Warn("🚨 SSH MODE DEPRECATION WARNING")
	m.logger.Warn("SSH-based Pi-hole access is being phased out in favor of the REST API")
	m.logger.Warn("Please configure API access for better performance and future compatibility")
	m.logger.Warn("Migration guide: docs/migration-ssh-to-api.md")
}

// logMigrationTimeline logs the migration timeline and important dates
func (m *MigrationManager) logMigrationTimeline() {
	now := time.Now()
	daysUntilDeprecation := int(m.strategy.MigrationDate.Sub(now).Hours() / 24)

	m.logger.Warn("📅 MIGRATION TIMELINE")
	m.logger.Warn("SSH deprecation date: %s (%d days remaining)",
		m.strategy.MigrationDate.Format("2006-01-02"), daysUntilDeprecation)

	if daysUntilDeprecation <= 30 {
		m.logger.Error("⚠️  SSH support removal imminent! Please migrate to API immediately")
	} else if daysUntilDeprecation <= 60 {
		m.logger.Warn("⚠️  SSH support will be removed soon, please plan migration")
	} else {
		m.logger.Info("ℹ️  SSH support available for %d more days", daysUntilDeprecation)
	}
}

// ValidateMigrationConfiguration checks if configuration is ready for migration
func (m *MigrationManager) ValidateMigrationConfiguration() *MigrationValidationResult {
	result := &MigrationValidationResult{
		IsValid:         true,
		Issues:          []string{},
		Recommendations: []string{},
		ReadinessScore:  0,
	}

	pihole := &m.config.Pihole

	// Check API configuration
	if pihole.APIPassword == "" {
		result.Issues = append(result.Issues, "API password not configured")
		result.Recommendations = append(result.Recommendations, "Set 'api_password' in Pi-hole configuration")
		result.IsValid = false
	} else {
		result.ReadinessScore += 40
		result.Recommendations = append(result.Recommendations, "✅ API password configured")
	}

	// Check API connectivity requirements
	if pihole.Host == "" {
		result.Issues = append(result.Issues, "Pi-hole host not configured")
		result.IsValid = false
	} else {
		result.ReadinessScore += 20
	}

	// Check for mixed configuration (good for transition)
	hasSSH := pihole.Username != "" && (pihole.Password != "" || pihole.KeyPath != "")
	hasAPI := pihole.APIPassword != ""

	if hasSSH && hasAPI {
		result.ReadinessScore += 30
		result.Recommendations = append(result.Recommendations, "✅ Both SSH and API configured - perfect for gradual migration")
	} else if hasAPI {
		result.ReadinessScore += 20
		result.Recommendations = append(result.Recommendations, "✅ API-only configuration - ready for modern deployment")
	} else if hasSSH {
		result.ReadinessScore += 10
		result.Issues = append(result.Issues, "Only SSH configured - migration needed")
		result.Recommendations = append(result.Recommendations, "Configure API access for better performance")
	}

	// Check migration mode setting
	if pihole.MigrationMode != "" {
		result.ReadinessScore += 10
		result.Recommendations = append(result.Recommendations, fmt.Sprintf("✅ Migration mode explicitly set: %s", pihole.MigrationMode))
	}

	return result
}

// MigrationValidationResult represents the result of migration configuration validation
type MigrationValidationResult struct {
	IsValid         bool     `json:"is_valid"`
	Issues          []string `json:"issues"`
	Recommendations []string `json:"recommendations"`
	ReadinessScore  int      `json:"readiness_score"` // 0-100
}

// GetMigrationStatus returns detailed migration status information
func (m *MigrationManager) GetMigrationStatus() *MigrationStatus {
	now := time.Now()
	mode := m.GetMigrationMode()
	validation := m.ValidateMigrationConfiguration()

	status := &MigrationStatus{
		CurrentMode:        mode,
		ConfigurationOK:    validation.IsValid,
		ReadinessScore:     validation.ReadinessScore,
		DaysUntilMigration: int(m.strategy.MigrationDate.Sub(now).Hours() / 24),
		SSHDeprecated:      now.After(m.strategy.MigrationDate),
		Recommendations:    validation.Recommendations,
		Issues:             validation.Issues,
		NextSteps:          m.generateNextSteps(mode, validation),
	}

	return status
}

// MigrationStatus provides comprehensive migration status information
type MigrationStatus struct {
	CurrentMode        MigrationMode `json:"current_mode"`
	ConfigurationOK    bool          `json:"configuration_ok"`
	ReadinessScore     int           `json:"readiness_score"`
	DaysUntilMigration int           `json:"days_until_migration"`
	SSHDeprecated      bool          `json:"ssh_deprecated"`
	Recommendations    []string      `json:"recommendations"`
	Issues             []string      `json:"issues"`
	NextSteps          []string      `json:"next_steps"`
}

// generateNextSteps generates actionable next steps based on current migration state
func (m *MigrationManager) generateNextSteps(mode MigrationMode, validation *MigrationValidationResult) []string {
	steps := []string{}

	switch mode {
	case MigrationModeSSHOnly:
		steps = append(steps, "1. Configure Pi-hole API password in Pi-hole admin interface")
		steps = append(steps, "2. Add 'api_password' to configuration file")
		steps = append(steps, "3. Test API connectivity")
		steps = append(steps, "4. Set migration_mode to 'api-first' for gradual transition")

	case MigrationModeAPIWithSSHFallback:
		steps = append(steps, "1. Monitor API connection reliability")
		steps = append(steps, "2. Verify all features work correctly with API")
		steps = append(steps, "3. Set migration_mode to 'api-only-warn' when confident")

	case MigrationModeAPIWithWarnings:
		steps = append(steps, "1. Remove SSH configuration to complete migration")
		steps = append(steps, "2. Update documentation and deployment scripts")
		steps = append(steps, "3. Set migration_mode to 'api-only'")

	case MigrationModeAPIOnly:
		steps = append(steps, "✅ Migration complete! Consider enabling additional API features")
	}

	// Add issue-specific steps
	if !validation.IsValid {
		steps = append([]string{"🚨 Fix configuration issues first:"}, steps...)
		for _, issue := range validation.Issues {
			steps = append(steps, fmt.Sprintf("   - %s", issue))
		}
	}

	return steps
}

// LogMigrationSummary logs a comprehensive migration summary
func (m *MigrationManager) LogMigrationSummary() {
	status := m.GetMigrationStatus()

	m.logger.Info("=== SSH-to-API Migration Status ===")
	m.logger.Info("Current Mode: %s", status.CurrentMode)
	m.logger.Info("Configuration Valid: %v", status.ConfigurationOK)
	m.logger.Info("Readiness Score: %d/100", status.ReadinessScore)

	if status.DaysUntilMigration > 0 {
		m.logger.Info("Days Until SSH Deprecation: %d", status.DaysUntilMigration)
	} else {
		m.logger.Warn("SSH Support Deprecated: %d days ago", -status.DaysUntilMigration)
	}

	if len(status.Issues) > 0 {
		m.logger.Error("⚠️ Issues found:")
		for _, issue := range status.Issues {
			m.logger.Error("  - %s", issue)
		}
	}

	if len(status.NextSteps) > 0 {
		m.logger.Info("📋 Next Steps:")
		for _, step := range status.NextSteps {
			m.logger.Info("  %s", step)
		}
	}
}
