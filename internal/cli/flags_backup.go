package cli

import (
	"flag"
	"fmt"
	"os"

	"pihole-analyzer/internal/config"
	"pihole-analyzer/internal/types"
)

// Flags represents command-line flags (API-only version)
type Flags struct {
	OnlineOnly   *bool
	NoExclude    *bool
	Pihole       *string
	Config       *string
	NoColor      *bool
	NoEmoji      *bool
	Quiet        *bool
	CreateConfig *bool
	ShowConfig   *bool
	PiholeSetup  *bool
	// Web UI flags
	EnableWeb  *bool
	WebPort    *int
	WebHost    *string
	DaemonMode *bool
	// Machine Learning flags
	EnableML        *bool
	MLTrain         *bool
	MLAnomalies     *bool
	MLTrends        *bool
	MLSensitivity   *float64
	MLMinConfidence *float64
}

// ParseFlags parses command-line flags and returns the flags struct
func ParseFlags() *Flags {
	flags := &Flags{
		OnlineOnly:   flag.Bool("online-only", false, "Show only clients that are currently online (have MAC addresses in ARP table)"),
		NoExclude:    flag.Bool("no-exclude", false, "Disable default exclusions (Docker networks, Pi-hole host)"),
		Pihole:       flag.String("pihole", "", "Analyze Pi-hole live data using the specified config file"),
		Config:       flag.String("config", "", "Configuration file path (default: ~/.pihole-analyzer/config.json)"),
		NoColor:      flag.Bool("no-color", false, "Disable colored output"),
		NoEmoji:      flag.Bool("no-emoji", false, "Disable emoji in output"),
		Quiet:        flag.Bool("quiet", false, "Suppress non-essential output (useful for CI/testing)"),
		CreateConfig: flag.Bool("create-config", false, "Create default configuration file and exit"),
		ShowConfig:   flag.Bool("show-config", false, "Show current configuration and exit"),
		PiholeSetup:  flag.Bool("pihole-setup", false, "Setup Pi-hole configuration"),
		// Web UI flags
		EnableWeb:  flag.Bool("web", false, "Enable web interface (starts HTTP server)"),
		WebPort:    flag.Int("web-port", 8080, "Port for web interface (default: 8080)"),
		WebHost:    flag.String("web-host", "localhost", "Host for web interface (default: localhost)"),
		DaemonMode: flag.Bool("daemon", false, "Run in daemon mode (implies --web)"),
		// Machine Learning flags
		EnableML:        flag.Bool("ml", false, "Enable machine learning analysis (anomaly detection and trend analysis)"),
		MLTrain:         flag.Bool("ml-train", false, "Train ML models with current data"),
		MLAnomalies:     flag.Bool("ml-anomalies", false, "Run anomaly detection only"),
		MLTrends:        flag.Bool("ml-trends", false, "Run trend analysis only"),
		MLSensitivity:   flag.Float64("ml-sensitivity", 0.7, "ML sensitivity level (0.0-1.0, default: 0.7)"),
		MLMinConfidence: flag.Float64("ml-min-confidence", 0.6, "Minimum confidence for ML predictions (0.0-1.0, default: 0.6)"),
	}

	flag.Parse()
	return flags
}

// HandleSpecialFlags handles flags that should cause immediate exit
func HandleSpecialFlags(flags *Flags) bool {
	if *flags.CreateConfig {
		if err := createDefaultConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating config: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Default configuration created successfully")
		return true
	}

	if *flags.ShowConfig {
		if err := showCurrentConfig(*flags.Config); err != nil {
			fmt.Fprintf(os.Stderr, "Error showing config: %v\n", err)
			os.Exit(1)
		}
		return true
	}

	if *flags.PiholeSetup {
		fmt.Println("Pi-hole setup wizard not yet implemented")
		return true
	}

	return false
}

// ApplyFlags applies command-line flags to configuration
func ApplyFlags(flags *Flags, cfg *types.Config) {
	if *flags.OnlineOnly {
		cfg.OnlineOnly = true
	}
	if *flags.NoExclude {
		cfg.NoExclude = true
	}
	if *flags.NoColor {
		cfg.Output.Colors = false
		cfg.Logging.EnableColors = false
	}
	if *flags.NoEmoji {
		cfg.Output.Emojis = false
		cfg.Logging.EnableEmojis = false
	}
	if *flags.Quiet {
		cfg.Quiet = true
		cfg.Logging.Level = "ERROR"
	}
}

// IsWebModeEnabled returns true if web mode is requested
func IsWebModeEnabled(flags *Flags) bool {
	return *flags.EnableWeb || *flags.DaemonMode
}

// IsDaemonMode returns true if daemon mode is requested
func IsDaemonMode(flags *Flags) bool {
	return *flags.DaemonMode
}

// GetWebConfig extracts web configuration from flags
func GetWebConfig(flags *Flags) map[string]any {
	return map[string]any{
		"enabled":     IsWebModeEnabled(flags),
		"port":        *flags.WebPort,
		"host":        *flags.WebHost,
		"daemon_mode": *flags.DaemonMode,
	}
}

// ValidateInput validates command-line input
func ValidateInput(flags *Flags) error {
	// No specific validation needed for API-only version
	return nil
}

// PrintStartupInfo prints startup information
func PrintStartupInfo(flags *Flags, cfg *types.Config) {
	if !cfg.Quiet {
		fmt.Println("🔍 Pi-hole Network Analyzer (API-only)")
		if *flags.Pihole != "" {
			fmt.Printf("📊 Analyzing Pi-hole data from: %s\n", *flags.Pihole)
		}
		if IsWebModeEnabled(flags) {
			fmt.Printf("🌐 Web interface enabled on http://%s:%d\n", *flags.WebHost, *flags.WebPort)
			if *flags.DaemonMode {
				fmt.Println("🔄 Running in daemon mode")
			}
		}
	}
}

func createDefaultConfig() error {
	configPath := config.GetConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("configuration file already exists at %s", configPath)
	}

	defaultConfig := config.DefaultConfig()
	return config.SaveConfig(defaultConfig, configPath)
}

func showCurrentConfig(configPath string) error {
	if configPath == "" {
		configPath = config.GetConfigPath()
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	fmt.Printf("Configuration loaded from: %s\n", configPath)
	fmt.Printf("Pi-hole API URL: %s\n", cfg.Pihole.Host)
	fmt.Printf("API Enabled: %t\n", cfg.Pihole.APIEnabled)
	fmt.Printf("Use HTTPS: %t\n", cfg.Pihole.UseHTTPS)
	fmt.Printf("Online Only: %t\n", cfg.OnlineOnly)
	fmt.Printf("Quiet Mode: %t\n", cfg.Quiet)

	return nil
}
