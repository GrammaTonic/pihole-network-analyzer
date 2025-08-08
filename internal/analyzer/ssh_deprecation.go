package analyzer

import (
	"time"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// SSHDeprecationManager handles the gradual deprecation of SSH functionality
type SSHDeprecationManager struct {
	logger   *logger.Logger
	config   *types.Config
	warnings []string
}

// NewSSHDeprecationManager creates a new SSH deprecation manager
func NewSSHDeprecationManager(config *types.Config, logger *logger.Logger) *SSHDeprecationManager {
	return &SSHDeprecationManager{
		logger:   logger,
		config:   config,
		warnings: make([]string, 0),
	}
}

// DeprecationWarning represents a single deprecation warning
type DeprecationWarning struct {
	Level    string `json:"level"`
	Message  string `json:"message"`
	Action   string `json:"action"`
	Timeline string `json:"timeline"`
	Priority int    `json:"priority"`
}

// GetDeprecationWarnings returns all relevant SSH deprecation warnings
func (sdm *SSHDeprecationManager) GetDeprecationWarnings() []DeprecationWarning {
	warnings := []DeprecationWarning{}

	// Check if using SSH mode
	if sdm.isUsingSSH() {
		warnings = append(warnings, DeprecationWarning{
			Level:    "CRITICAL",
			Message:  "SSH mode is deprecated and will be removed in future versions",
			Action:   "Configure Pi-hole API access for continued service",
			Timeline: "SSH removal planned for v2.0.0",
			Priority: 1,
		})

		warnings = append(warnings, DeprecationWarning{
			Level:    "WARNING",
			Message:  "SSH-based database access has security and performance limitations",
			Action:   "Migrate to REST API for enhanced security and performance",
			Timeline: "Immediate migration recommended",
			Priority: 2,
		})
	}

	// Check migration readiness
	if sdm.canMigrateToAPI() {
		warnings = append(warnings, DeprecationWarning{
			Level:    "INFO",
			Message:  "Your configuration is ready for API migration",
			Action:   "Run with --migration-mode api-first to test API connectivity",
			Timeline: "Ready for immediate migration",
			Priority: 3,
		})
	} else {
		warnings = append(warnings, DeprecationWarning{
			Level:    "ACTION_REQUIRED",
			Message:  "API configuration incomplete - migration not possible",
			Action:   "Configure api_password in your Pi-hole settings and config file",
			Timeline: "Complete before SSH removal",
			Priority: 1,
		})
	}

	return warnings
}

// ShowDeprecationWarnings displays all relevant deprecation warnings
func (sdm *SSHDeprecationManager) ShowDeprecationWarnings() {
	warnings := sdm.GetDeprecationWarnings()

	if len(warnings) == 0 {
		return
	}

	sdm.logger.Warn("🚨 SSH DEPRECATION WARNINGS")
	sdm.logger.Warn("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for _, warning := range warnings {
		switch warning.Level {
		case "CRITICAL":
			sdm.logger.Error("🔴 CRITICAL: %s", warning.Message)
			sdm.logger.Error("   Action: %s", warning.Action)
			if warning.Timeline != "" {
				sdm.logger.Error("   Timeline: %s", warning.Timeline)
			}
		case "WARNING":
			sdm.logger.Warn("🟡 WARNING: %s", warning.Message)
			sdm.logger.Warn("   Action: %s", warning.Action)
		case "INFO":
			sdm.logger.Info("🟢 INFO: %s", warning.Message)
			sdm.logger.Info("   Action: %s", warning.Action)
		case "ACTION_REQUIRED":
			sdm.logger.Error("🟠 ACTION REQUIRED: %s", warning.Message)
			sdm.logger.Error("   Action: %s", warning.Action)
		}
		sdm.logger.Info("")
	}

	sdm.showMigrationTools()
	sdm.logger.Warn("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// showMigrationTools displays available migration tools and resources
func (sdm *SSHDeprecationManager) showMigrationTools() {
	sdm.logger.Info("🛠️  Migration Tools & Resources:")
	sdm.logger.Info("   • phase4-migration-tool    - Comprehensive migration utility")
	sdm.logger.Info("   • --migration-status       - Check current migration readiness")
	sdm.logger.Info("   • --validate-migration     - Validate API configuration")
	sdm.logger.Info("   • --migration-mode         - Control migration behavior")
	sdm.logger.Info("")
	sdm.logger.Info("📖 Documentation:")
	sdm.logger.Info("   • docs/migration-ssh-to-api.md  - Complete migration guide")
	sdm.logger.Info("   • docs/api.md                   - API configuration reference")
	sdm.logger.Info("   • docs/troubleshooting.md       - Common migration issues")
	sdm.logger.Info("")
}

// isUsingSSH checks if the current configuration is using SSH mode
func (sdm *SSHDeprecationManager) isUsingSSH() bool {
	// Check migration mode
	switch sdm.config.Pihole.MigrationMode {
	case "ssh-only", "auto":
		return true
	case "api-first":
		// Using API first but SSH is still configured as fallback
		return sdm.config.Pihole.Host != ""
	case "api-only", "api-only-warn":
		return false
	default:
		// Default behavior - check if SSH is configured
		return sdm.config.Pihole.Host != "" && sdm.config.Pihole.Username != ""
	}
}

// canMigrateToAPI checks if the configuration allows migration to API mode
func (sdm *SSHDeprecationManager) canMigrateToAPI() bool {
	// Check if API configuration is present
	return sdm.config.Pihole.APIEnabled &&
		sdm.config.Pihole.APIPassword != ""
}

// GetMigrationTimeline returns the deprecation timeline for SSH functionality
func (sdm *SSHDeprecationManager) GetMigrationTimeline() []TimelineEvent {
	return []TimelineEvent{
		{
			Version:     "v1.4.0",
			Date:        "2024-12-01",
			Status:      "COMPLETED",
			Description: "Pi-hole API client implementation",
			Impact:      "API support added alongside SSH",
		},
		{
			Version:     "v1.5.0",
			Date:        "2024-12-15",
			Status:      "COMPLETED",
			Description: "Migration framework and tooling",
			Impact:      "Migration tools and validation available",
		},
		{
			Version:     "v1.6.0",
			Date:        "2025-01-01",
			Status:      "CURRENT",
			Description: "API-first mode and SSH deprecation warnings",
			Impact:      "API becomes default, SSH shows warnings",
		},
		{
			Version:     "v1.7.0",
			Date:        "2025-02-01",
			Status:      "PLANNED",
			Description: "SSH mode requires explicit flag",
			Impact:      "SSH mode disabled by default, requires --force-ssh",
		},
		{
			Version:     "v2.0.0",
			Date:        "2025-03-01",
			Status:      "PLANNED",
			Description: "SSH mode removal",
			Impact:      "SSH functionality completely removed",
		},
	}
}

// TimelineEvent represents a single event in the deprecation timeline
type TimelineEvent struct {
	Version     string `json:"version"`
	Date        string `json:"date"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
}

// ShowMigrationTimeline displays the complete SSH deprecation timeline
func (sdm *SSHDeprecationManager) ShowMigrationTimeline() {
	timeline := sdm.GetMigrationTimeline()

	sdm.logger.Info("📅 SSH Deprecation Timeline:")
	sdm.logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for _, event := range timeline {
		status := ""
		switch event.Status {
		case "COMPLETED":
			status = "✅"
		case "CURRENT":
			status = "🔄"
		case "PLANNED":
			status = "📋"
		}

		sdm.logger.Info("%s %s (%s)", status, event.Version, event.Date)
		sdm.logger.Info("   %s", event.Description)
		sdm.logger.Info("   Impact: %s", event.Impact)
		sdm.logger.Info("")
	}

	sdm.logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// CheckDeprecationStatus returns the current deprecation status
func (sdm *SSHDeprecationManager) CheckDeprecationStatus() DeprecationStatus {
	status := DeprecationStatus{
		CurrentMode:       sdm.getCurrentMode(),
		IsSSHDeprecated:   true,
		MigrationRequired: sdm.isUsingSSH(),
		APIReady:          sdm.canMigrateToAPI(),
		DeprecationPhase:  sdm.getDeprecationPhase(),
		Timestamp:         time.Now(),
	}

	status.Recommendations = sdm.getRecommendations(status)
	return status
}

// DeprecationStatus represents the current deprecation status
type DeprecationStatus struct {
	CurrentMode       string    `json:"current_mode"`
	IsSSHDeprecated   bool      `json:"is_ssh_deprecated"`
	MigrationRequired bool      `json:"migration_required"`
	APIReady          bool      `json:"api_ready"`
	DeprecationPhase  string    `json:"deprecation_phase"`
	Recommendations   []string  `json:"recommendations"`
	Timestamp         time.Time `json:"timestamp"`
}

// getCurrentMode returns the current data access mode
func (sdm *SSHDeprecationManager) getCurrentMode() string {
	if sdm.config.Pihole.MigrationMode != "" {
		return sdm.config.Pihole.MigrationMode
	}

	if sdm.canMigrateToAPI() {
		return "api-capable"
	}

	return "ssh-only"
}

// getDeprecationPhase returns the current phase in the deprecation timeline
func (sdm *SSHDeprecationManager) getDeprecationPhase() string {
	// Based on current implementation, we're in the warning phase
	return "warning-phase"
}

// getRecommendations returns migration recommendations based on current status
func (sdm *SSHDeprecationManager) getRecommendations(status DeprecationStatus) []string {
	recommendations := []string{}

	if status.MigrationRequired && !status.APIReady {
		recommendations = append(recommendations, "Configure Pi-hole API password in admin interface")
		recommendations = append(recommendations, "Add 'api_password' to configuration file")
		recommendations = append(recommendations, "Test API connectivity with --validate-migration")
	}

	if status.MigrationRequired && status.APIReady {
		recommendations = append(recommendations, "Test API mode with --migration-mode api-first")
		recommendations = append(recommendations, "Complete migration with --migration-mode api-only")
		recommendations = append(recommendations, "Remove SSH configuration once API is confirmed working")
	}

	if status.CurrentMode == "ssh-only" {
		recommendations = append(recommendations, "SSH mode will be removed in v2.0.0")
		recommendations = append(recommendations, "Begin migration process immediately")
		recommendations = append(recommendations, "Use phase4-migration-tool for guided migration")
	}

	return recommendations
}

// ValidateDeprecationCompliance checks if configuration is compliant with deprecation timeline
func (sdm *SSHDeprecationManager) ValidateDeprecationCompliance() (bool, []string) {
	issues := []string{}

	// Check if using deprecated SSH mode
	if sdm.isUsingSSH() && !sdm.canMigrateToAPI() {
		issues = append(issues, "Using SSH mode without API configuration")
		issues = append(issues, "SSH mode will be removed in v2.0.0")
	}

	// Check for API readiness
	if !sdm.canMigrateToAPI() {
		issues = append(issues, "API configuration incomplete")
		issues = append(issues, "Configure api_password for future compatibility")
	}

	return len(issues) == 0, issues
}
