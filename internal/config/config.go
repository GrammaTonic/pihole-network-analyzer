package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"pihole-network-analyzer/internal/types"
)

// DefaultConfig returns the default configuration
func DefaultConfig() *types.Config {
	homeDir, _ := os.UserHomeDir()

	return &types.Config{
		OnlineOnly: false,
		NoExclude:  false,
		TestMode:   false,

		Pihole: types.PiholeConfig{
			Host:     "",
			Port:     22,
			Username: "pi",
			Password: "",
			KeyFile:  filepath.Join(homeDir, ".ssh", "id_rsa"),
			DBPath:   "/etc/pihole/pihole-FTL.db",
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
	}
}

// LoadConfig loads configuration from file, falling back to defaults
func LoadConfig(configPath string) (*types.Config, error) {
	config := DefaultConfig()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("Config file not found at %s, using defaults\n", configPath)
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

	fmt.Printf("Configuration loaded from %s\n", configPath)
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

	fmt.Printf("Configuration saved to %s\n", configPath)
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
	fmt.Println("\nCurrent Configuration:")
	fmt.Println("======================")
	fmt.Printf("Online Only:       %t\n", config.OnlineOnly)
	fmt.Printf("No Exclude:        %t\n", config.NoExclude)
	fmt.Printf("Test Mode:         %t\n", config.TestMode)
	fmt.Printf("Max Clients:       %d\n", config.Output.MaxClients)
	fmt.Printf("Max Domains:       %d\n", config.Output.MaxDomains)
	fmt.Printf("Save Reports:      %t\n", config.Output.SaveReports)
	fmt.Printf("Report Directory:  %s\n", config.Output.ReportDir)
	fmt.Printf("Verbose Output:    %t\n", config.Output.VerboseOutput)

	fmt.Println("\nPi-hole Configuration:")
	fmt.Printf("  Host:            %s\n", config.Pihole.Host)
	fmt.Printf("  Port:            %d\n", config.Pihole.Port)
	fmt.Printf("  Username:        %s\n", config.Pihole.Username)
	if config.Pihole.Password != "" {
		fmt.Printf("  Password:        %s\n", "***configured***")
	} else {
		fmt.Printf("  Password:        %s\n", "not set")
	}
	fmt.Printf("  Key File:        %s\n", config.Pihole.KeyFile)
	fmt.Printf("  Database Path:   %s\n", config.Pihole.DBPath)

	fmt.Println("\nExclusion Networks:")
	for _, network := range config.Exclusions.ExcludeNetworks {
		fmt.Printf("  - %s\n", network)
	}
	if len(config.Exclusions.ExcludeIPs) > 0 {
		fmt.Println("Exclusion IPs:")
		for _, ip := range config.Exclusions.ExcludeIPs {
			fmt.Printf("  - %s\n", ip)
		}
	}
	if len(config.Exclusions.ExcludeHosts) > 0 {
		fmt.Println("Exclusion Hosts:")
		for _, host := range config.Exclusions.ExcludeHosts {
			fmt.Printf("  - %s\n", host)
		}
	}
	fmt.Println()
}

// GetConfigPath returns the default configuration file path
func GetConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".pihole-analyzer", "config.json")
}
