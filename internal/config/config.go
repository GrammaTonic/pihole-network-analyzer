package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// DefaultConfig returns the default configuration
func DefaultConfig() *types.Config {
	return &types.Config{
		OnlineOnly: false,
		NoExclude:  false,
		TestMode:   false,

		Pihole: types.PiholeConfig{
			Host:        "",
			Port:        80,
			APIEnabled:  true,
			APIPassword: "",
			UseHTTPS:    false,
			APITimeout:  30,
		},

		Exclusions: types.ExclusionConfig{
			ExcludeNetworks: []string{
				"172.16.0.0/12", // Docker default networks
				"127.0.0.0/8",   // Loopback
			},
			ExcludeIPs:   []string{},
			ExcludeHosts: []string{"pi.hole"},
		},

		Output: types.OutputConfig{
			SaveReports:   true,
			ReportDir:     ".",
			VerboseOutput: false,
			MaxClients:    20,
			MaxDomains:    10,
		},

		Logging: types.LoggingConfig{
			Level:         "INFO",
			EnableColors:  true,
			EnableEmojis:  true,
			ShowTimestamp: true,
			ShowCaller:    false,
		},

		Web: types.WebConfig{
			Enabled:      false,
			Port:         8080,
			Host:         "localhost",
			DaemonMode:   false,
			ReadTimeout:  10,
			WriteTimeout: 10,
			IdleTimeout:  60,
		},

		Metrics: types.MetricsConfig{
			Enabled:               true,
			Port:                  "9090",
			Host:                  "localhost",
			EnableEndpoint:        true,
			CollectMetrics:        true,
			EnableDetailedMetrics: true,
		},
	}
}

// LoadConfig loads configuration from file, falling back to defaults
func LoadConfig(configPath string) (*types.Config, error) {
	config := DefaultConfig()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Info("Config file not found at %s, using defaults", configPath)
		return config, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	// Parse JSON
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	logger.Success("Configuration loaded from %s", configPath)
	return config, nil
}

// SaveConfig saves the current configuration to file
func SaveConfig(config *types.Config, configPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %v", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	logger.Success("Configuration saved to %s", configPath)
	return nil
}

// MergeFlags merges command-line flag values into configuration
func MergeFlags(config *types.Config, onlineOnly, noExclude, testMode bool, piholeConfig string) {
	if onlineOnly {
		config.OnlineOnly = true
	}
	if noExclude {
		config.NoExclude = true
	}
	if testMode {
		config.TestMode = true
	}
	if piholeConfig != "" {
		// Parse pihole config if provided
		// This would load pihole-specific settings
	}
}

// CreateDefaultConfigFile creates a default configuration file
func CreateDefaultConfigFile(configPath string) error {
	config := DefaultConfig()
	return SaveConfig(config, configPath)
}

// ShowConfig displays the current configuration
func ShowConfig(config *types.Config) {
	logger.Info("\nCurrent Configuration:")
	logger.Info("======================")
	logger.Info("Online Only:       %t", config.OnlineOnly)
	logger.Info("No Exclude:        %t", config.NoExclude)
	logger.Info("Test Mode:         %t", config.TestMode)
	logger.Info("Max Clients:       %d", config.Output.MaxClients)
	logger.Info("Max Domains:       %d", config.Output.MaxDomains)
	logger.Info("Save Reports:      %t", config.Output.SaveReports)
	logger.Info("Report Directory:  %s", config.Output.ReportDir)
	logger.Info("Verbose Output:    %t", config.Output.VerboseOutput)

	logger.Info("\nPi-hole Configuration:")
	logger.Info("  Host:            %s", config.Pihole.Host)
	logger.Info("  Port:            %d", config.Pihole.Port)
	logger.Info("  API Enabled:     %t", config.Pihole.APIEnabled)
	if config.Pihole.APIPassword != "" {
		logger.Info("  API Password:    %s", "***configured***")
	} else {
		logger.Info("  API Password:    %s", "not set")
	}
	logger.Info("  Use HTTPS:       %t", config.Pihole.UseHTTPS)
	logger.Info("  API Timeout:     %d", config.Pihole.APITimeout)

	logger.Info("\nExclusion Networks:")
	for _, network := range config.Exclusions.ExcludeNetworks {
		logger.Info("  - %s", network)
	}
	if len(config.Exclusions.ExcludeIPs) > 0 {
		logger.Info("Exclusion IPs:")
		for _, ip := range config.Exclusions.ExcludeIPs {
			logger.Info("  - %s", ip)
		}
	}
	if len(config.Exclusions.ExcludeHosts) > 0 {
		logger.Info("Exclusion Hosts:")
		for _, host := range config.Exclusions.ExcludeHosts {
			logger.Info("  - %s", host)
		}
	}
	logger.Info("")
}

// GetConfigPath returns the default configuration file path
func GetConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".pihole-analyzer", "config.json")
}
