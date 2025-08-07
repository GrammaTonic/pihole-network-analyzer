package internal
package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	// Analysis options
	OnlineOnly bool `json:"online_only"`
	NoExclude  bool `json:"no_exclude"`
	TestMode   bool `json:"test_mode"`

	// Pi-hole configuration
	Pihole PiholeConfig `json:"pihole"`

	// Exclusion configuration
	Exclusions ExclusionConfig `json:"exclusions"`

	// Output configuration
	Output OutputConfig `json:"output"`
}

// OutputConfig holds output-related settings
type OutputConfig struct {
	SaveReports   bool   `json:"save_reports"`
	ReportDir     string `json:"report_dir"`
	VerboseOutput bool   `json:"verbose_output"`
	MaxClients    int    `json:"max_clients_display"`
	MaxDomains    int    `json:"max_domains_display"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()

	return &Config{
		OnlineOnly: false,
		NoExclude:  false,
		TestMode:   false,

		Pihole: PiholeConfig{
			Host:     "",
			Port:     "22",
			Username: "pi",
			Password: "",
			KeyFile:  filepath.Join(homeDir, ".ssh", "id_rsa"),
			DBPath:   "/etc/pihole/pihole-FTL.db",
		},

		Exclusions: ExclusionConfig{
			ExcludeNetworks: []string{
				"172.16.0.0/12", // Docker default networks
				"127.0.0.0/8",   // Loopback
			},
			ExcludeIPs:   []string{},
			ExcludeHosts: []string{"pi.hole"},
		},

		Output: OutputConfig{
			SaveReports:   true,
			ReportDir:     ".",
			VerboseOutput: false,
			MaxClients:    20,
			MaxDomains:    10,
		},
	}
}

// LoadConfig loads configuration from file, falling back to defaults
func LoadConfig(configPath string) (*Config, error) {
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
func SaveConfig(config *Config, configPath string) error {
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

// MergeFlags merges command line flags with config file settings
// Command line flags take precedence over config file
func MergeFlags(config *Config) {
	// Only override config if flag was explicitly set
	if *onlineOnlyFlag {
		config.OnlineOnly = true
	}
	if *noExcludeFlag {
		config.NoExclude = true
	}
	if *testModeFlag {
		config.TestMode = true
	}
	if *piholeFlag != "" {
		// Pi-hole flag specified, load that specific config
		// This will be handled in main()
	}
}

// CreateDefaultConfigFile creates a default configuration file
func CreateDefaultConfigFile(configPath string) error {
	config := DefaultConfig()
	return SaveConfig(config, configPath)
}

// ShowConfig displays the current configuration
func ShowConfig(config *Config) {
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
	fmt.Printf("  Port:            %s\n", config.Pihole.Port)
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
	return filepath.Join(homeDir, ".dns-analyzer", "config.json")
}
