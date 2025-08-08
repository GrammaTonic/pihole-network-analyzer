package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// MigrationConfigManager handles migration-related configuration operations
type MigrationConfigManager struct {
	logger     *logger.Logger
	configPath string
}

// NewMigrationConfigManager creates a new migration configuration manager
func NewMigrationConfigManager(configPath string, logger *logger.Logger) *MigrationConfigManager {
	return &MigrationConfigManager{
		logger:     logger.Component("migration-config"),
		configPath: configPath,
	}
}

// BackupConfig creates a backup of the current configuration before migration
func (m *MigrationConfigManager) BackupConfig() error {
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s", m.configPath)
	}

	// Create backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.backup-%s", m.configPath, timestamp)

	// Read original config
	data, err := ioutil.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Write backup
	if err := ioutil.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	m.logger.Info("Configuration backup created: %s", backupPath)
	return nil
}

// MigrateConfigToAPI migrates SSH-only configuration to include API settings
func (m *MigrationConfigManager) MigrateConfigToAPI(config *types.Config, apiPassword string) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Backup current config first
	if err := m.BackupConfig(); err != nil {
		m.logger.Warn("Failed to create backup: %v", err)
	}

	// Update Pi-hole configuration with API settings
	pihole := &config.Pihole

	// Set API configuration
	pihole.APIEnabled = true
	pihole.APIPassword = apiPassword
	pihole.UseHTTPS = false // Default to HTTP, user can change later
	pihole.APITimeout = 30  // 30 second timeout

	// Set migration mode to api-first for gradual transition
	if pihole.MigrationMode == "" {
		pihole.MigrationMode = "api-first"
	}

	// Log the migration changes
	m.logger.Info("🔄 Migrating configuration to include API settings")
	m.logger.Info("API enabled: %v", pihole.APIEnabled)
	m.logger.Info("Migration mode: %s", pihole.MigrationMode)

	// Keep SSH settings for fallback during transition
	if pihole.Username != "" {
		m.logger.Info("SSH configuration preserved for fallback during migration")
	}

	return m.SaveConfig(config)
}

// SetMigrationMode updates the migration mode in the configuration
func (m *MigrationConfigManager) SetMigrationMode(config *types.Config, mode string) error {
	validModes := []string{"ssh-only", "api-first", "api-only-warn", "api-only", "auto"}

	// Validate migration mode
	valid := false
	for _, validMode := range validModes {
		if mode == validMode {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid migration mode: %s. Valid modes: %v", mode, validModes)
	}

	// Backup current config first
	if err := m.BackupConfig(); err != nil {
		m.logger.Warn("Failed to create backup: %v", err)
	}

	// Update migration mode
	config.Pihole.MigrationMode = mode
	m.logger.Info("Migration mode updated: %s", mode)

	return m.SaveConfig(config)
}

// RemoveSSHConfig removes SSH configuration settings (final migration step)
func (m *MigrationConfigManager) RemoveSSHConfig(config *types.Config) error {
	// Backup current config first
	if err := m.BackupConfig(); err != nil {
		m.logger.Warn("Failed to create backup: %v", err)
	}

	pihole := &config.Pihole

	// Ensure API is properly configured before removing SSH
	if pihole.APIPassword == "" {
		return fmt.Errorf("cannot remove SSH configuration: API password not set")
	}

	// Log what we're removing
	m.logger.Info("🗑️  Removing SSH configuration as part of migration completion")
	if pihole.Username != "" {
		m.logger.Info("Removing SSH username: %s", pihole.Username)
	}
	if pihole.KeyPath != "" {
		m.logger.Info("Removing SSH key path: %s", pihole.KeyPath)
	}

	// Clear SSH settings
	pihole.Username = ""
	pihole.Password = ""
	pihole.KeyPath = ""
	pihole.SSHKeyPath = ""
	pihole.KeyFile = ""
	pihole.UseSSH = false

	// Set final migration mode
	pihole.MigrationMode = "api-only"

	m.logger.Info("✅ SSH configuration removed, migration complete")
	return m.SaveConfig(config)
}

// SaveConfig saves the configuration to file
func (m *MigrationConfigManager) SaveConfig(config *types.Config) error {
	// Ensure directory exists
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := ioutil.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	m.logger.Info("Configuration saved: %s", m.configPath)
	return nil
}

// CreateMigrationGuideConfig creates a sample configuration for migration
func (m *MigrationConfigManager) CreateMigrationGuideConfig() *types.Config {
	config := &types.Config{
		OnlineOnly: false,
		NoExclude:  false,
		TestMode:   false,
		Quiet:      false,
		Pihole: types.PiholeConfig{
			Host:         "pi.hole",
			Port:         22,
			Username:     "pi",
			Password:     "", // Leave empty, use key auth
			KeyPath:      "/home/user/.ssh/id_rsa",
			DatabasePath: "/etc/pihole/pihole-FTL.db",

			// API Configuration (Phase 4 additions)
			APIEnabled:  true,
			APIPassword: "your-api-password-here",
			UseHTTPS:    false,
			APITimeout:  30,

			// Migration settings
			MigrationMode: "api-first", // Start with API-first mode
		},
		Output: types.OutputConfig{
			Colors:     true,
			Emojis:     true,
			Verbose:    false,
			Format:     "terminal",
			MaxClients: 50,
			MaxDomains: 20,
		},
		Exclusions: types.ExclusionConfig{
			Networks:        []string{"127.0.0.0/8", "169.254.0.0/16"},
			IPs:             []string{"127.0.0.1"},
			ExcludeNetworks: []string{"172.17.0.0/16", "172.18.0.0/16"},
			EnableDocker:    false,
		},
		Logging: types.LoggingConfig{
			Level:        "info",
			EnableColors: true,
			EnableEmojis: true,
			OutputFile:   "",
		},
	}

	return config
}

// ValidateConfigForMigration validates configuration readiness for migration
func (m *MigrationConfigManager) ValidateConfigForMigration(config *types.Config) (*MigrationValidationReport, error) {
	report := &MigrationValidationReport{
		IsReady:     true,
		Issues:      []string{},
		Warnings:    []string{},
		Suggestions: []string{},
		Score:       0,
	}

	pihole := &config.Pihole

	// Check basic connectivity information
	if pihole.Host == "" {
		report.Issues = append(report.Issues, "Pi-hole host not configured")
		report.IsReady = false
	} else {
		report.Score += 20
	}

	// Check SSH configuration
	hasSSHAuth := (pihole.Username != "" && pihole.Password != "") ||
		(pihole.Username != "" && pihole.KeyPath != "")
	if hasSSHAuth {
		report.Score += 25
		report.Suggestions = append(report.Suggestions, "SSH configuration detected - good for fallback during migration")
	}

	// Check API configuration
	if pihole.APIPassword == "" {
		report.Issues = append(report.Issues, "API password not configured")
		report.Suggestions = append(report.Suggestions, "Configure API password in Pi-hole admin interface")
		report.IsReady = false
	} else {
		report.Score += 40
		report.Suggestions = append(report.Suggestions, "✅ API password configured")
	}

	// Check migration mode
	if pihole.MigrationMode == "" {
		report.Warnings = append(report.Warnings, "Migration mode not explicitly set")
		report.Suggestions = append(report.Suggestions, "Set migration_mode to 'api-first' for gradual transition")
	} else {
		report.Score += 15
		report.Suggestions = append(report.Suggestions, fmt.Sprintf("Migration mode set: %s", pihole.MigrationMode))
	}

	// Evaluate overall readiness
	if report.Score >= 80 {
		report.Suggestions = append(report.Suggestions, "✅ Configuration is ready for migration")
	} else if report.Score >= 60 {
		report.Warnings = append(report.Warnings, "Configuration partially ready - address suggestions")
	} else {
		report.Issues = append(report.Issues, "Configuration needs significant updates before migration")
	}

	return report, nil
}

// MigrationValidationReport represents the result of migration validation
type MigrationValidationReport struct {
	IsReady     bool     `json:"is_ready"`
	Score       int      `json:"score"`       // 0-100
	Issues      []string `json:"issues"`      // Blocking issues
	Warnings    []string `json:"warnings"`    // Non-blocking concerns
	Suggestions []string `json:"suggestions"` // Recommendations
}

// GenerateConfigMigrationPlan creates a step-by-step migration plan
func (m *MigrationConfigManager) GenerateConfigMigrationPlan(config *types.Config) *ConfigMigrationPlan {
	plan := &ConfigMigrationPlan{
		CurrentState: m.analyzeCurrentState(config),
		Steps:        []MigrationStep{},
		Timeline:     "Immediate to 30 days",
		RiskLevel:    "Low",
	}

	pihole := &config.Pihole

	// Step 1: API Configuration
	if pihole.APIPassword == "" {
		plan.Steps = append(plan.Steps, MigrationStep{
			Phase:       1,
			Title:       "Configure Pi-hole API Access",
			Description: "Set up API password in Pi-hole admin interface",
			Commands: []string{
				"1. Open Pi-hole admin interface (http://pi.hole/admin)",
				"2. Go to Settings → API",
				"3. Generate or set API password",
				"4. Add 'api_password' to configuration file",
			},
			Validation: "Test API connectivity with migration tool",
			Required:   true,
		})
	}

	// Step 2: Configuration Migration
	plan.Steps = append(plan.Steps, MigrationStep{
		Phase:       2,
		Title:       "Update Configuration for Migration",
		Description: "Add migration settings to configuration",
		Commands: []string{
			"pihole-analyzer migrate config --api-password YOUR_PASSWORD",
			"# Or manually add to config:",
			`"migration_mode": "api-first"`,
		},
		Validation: "Verify configuration with: pihole-analyzer validate-config",
		Required:   true,
	})

	// Step 3: Testing Phase
	plan.Steps = append(plan.Steps, MigrationStep{
		Phase:       3,
		Title:       "Test API Connectivity",
		Description: "Verify API works correctly with fallback to SSH",
		Commands: []string{
			"pihole-analyzer --migration-mode api-first",
			"# Monitor logs for API usage vs SSH fallback",
		},
		Validation: "Check logs for successful API connections",
		Required:   true,
	})

	// Step 4: Full Migration
	plan.Steps = append(plan.Steps, MigrationStep{
		Phase:       4,
		Title:       "Complete Migration",
		Description: "Switch to API-only mode",
		Commands: []string{
			"pihole-analyzer migrate config --mode api-only",
			"# Remove SSH settings when confident",
			"pihole-analyzer migrate config --remove-ssh",
		},
		Validation: "Verify all functionality works without SSH",
		Required:   false,
	})

	return plan
}

// ConfigMigrationPlan represents a complete migration plan
type ConfigMigrationPlan struct {
	CurrentState string          `json:"current_state"`
	Steps        []MigrationStep `json:"steps"`
	Timeline     string          `json:"timeline"`
	RiskLevel    string          `json:"risk_level"`
}

// MigrationStep represents a single step in the migration process
type MigrationStep struct {
	Phase       int      `json:"phase"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Commands    []string `json:"commands"`
	Validation  string   `json:"validation"`
	Required    bool     `json:"required"`
}

// analyzeCurrentState analyzes the current configuration state
func (m *MigrationConfigManager) analyzeCurrentState(config *types.Config) string {
	pihole := &config.Pihole

	hasSSH := pihole.Username != "" && (pihole.Password != "" || pihole.KeyPath != "")
	hasAPI := pihole.APIPassword != ""

	if hasSSH && hasAPI {
		return "Mixed SSH and API configuration - ready for gradual migration"
	} else if hasAPI {
		return "API-only configuration - migration complete"
	} else if hasSSH {
		return "SSH-only configuration - needs API setup"
	} else {
		return "No valid configuration - needs complete setup"
	}
}
