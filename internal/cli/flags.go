package cli

import (
	"flag"
	"fmt"
	"os"

	"pihole-analyzer/internal/colors"
	"pihole-analyzer/internal/config"
	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// Flags represents command-line flags (production version - no test flags)
type Flags struct {
	OnlineOnly   *bool
	NoExclude    *bool
	Pihole       *string
	PiholeSetup  *bool
	Config       *string
	ShowConfig   *bool
	CreateConfig *bool
	NoColor      *bool
	NoEmoji      *bool
	Quiet        *bool
}

// ParseFlags parses command-line flags and returns the flags struct
func ParseFlags() *Flags {
	flags := &Flags{
		OnlineOnly:   flag.Bool("online-only", false, "Show only clients that are currently online (have MAC addresses in ARP table)"),
		NoExclude:    flag.Bool("no-exclude", false, "Disable default exclusions (Docker networks, Pi-hole host)"),
		Pihole:       flag.String("pihole", "", "Analyze Pi-hole live data using the specified config file"),
		PiholeSetup:  flag.Bool("pihole-setup", false, "Setup Pi-hole configuration"),
		Config:       flag.String("config", "", "Configuration file path (default: ~/.pihole-analyzer/config.json)"),
		ShowConfig:   flag.Bool("show-config", false, "Show current configuration and exit"),
		CreateConfig: flag.Bool("create-config", false, "Create default configuration file and exit"),
		NoColor:      flag.Bool("no-color", false, "Disable colored output"),
		NoEmoji:      flag.Bool("no-emoji", false, "Disable emoji in output"),
		Quiet:        flag.Bool("quiet", false, "Suppress non-essential output (useful for CI/testing)"),
	}

	flag.Parse()
	return flags
}

// HandleSpecialFlags handles flags that should exit immediately
func HandleSpecialFlags(flags *Flags) bool {
	if *flags.PiholeSetup {
		return true // Caller should handle pihole setup
	}

	if *flags.ShowConfig {
		cfg, err := config.LoadConfig(*flags.Config)
		if err != nil {
			logger.Error("Error loading config: %v", err)
			return true
		}

		logger.Info("%s", colors.Header("Current Configuration"))
		config.ShowConfig(cfg)
		return true
	}

	if *flags.CreateConfig {
		configPath := config.GetConfigPath()
		if *flags.Config != "" {
			configPath = *flags.Config
		}

		err := config.CreateDefaultConfigFile(configPath)
		if err != nil {
			logger.Error("Error creating config: %v", err)
		} else {
			logger.Success("Default configuration created at: %s", configPath)
		}
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
		colors.DisableColors()
	}
	if *flags.NoEmoji {
		colors.DisableEmojis()
	}
	if *flags.Quiet {
		cfg.Quiet = true
	}
}

// ShowUsage displays usage information
func ShowUsage() {
	logger.Info("Usage: %s [options]\n", os.Args[0])
	logger.Info("Pi-hole Network Analyzer - Analyze Pi-hole DNS queries and network traffic\n")
	logger.Info("Examples:")
	logger.Info("  %s --pihole config.json     # Analyze live Pi-hole data", os.Args[0])
	logger.Info("  %s --pihole-setup           # Setup Pi-hole SSH configuration", os.Args[0])
	logger.Info("  %s --show-config            # Show current configuration", os.Args[0])
	logger.Info("  %s --create-config          # Create default config file", os.Args[0])
	logger.Info("\nOptions:")
	flag.PrintDefaults()
}

// ValidateInput validates command-line input
func ValidateInput(flags *Flags) error {
	if *flags.Pihole == "" && !*flags.PiholeSetup {
		return fmt.Errorf("Pi-hole configuration required. Use --pihole <config.json> or --pihole-setup to create one")
	}
	return nil
}

// PrintStartupInfo prints startup information
func PrintStartupInfo(flags *Flags, cfg *types.Config) {
	if !cfg.Quiet {
		logger.Info("%s", colors.Header("🕳️ Pi-hole Network Analyzer"))
		logger.Info("Analyzing Pi-hole DNS data with network insights\n")
	}
}
