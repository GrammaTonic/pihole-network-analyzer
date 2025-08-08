package cli

import (
	"flag"
	"fmt"
	"os"

	"pihole-analyzer/internal/colors"
	"pihole-analyzer/internal/config"
	"pihole-analyzer/internal/interfaces"
	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// Flags represents command-line flags (production version with Phase 4 migration support)
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

	// Phase 4: Migration flags
	MigrationMode   *string
	ValidateConfig  *bool
	MigrationStatus *bool
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

		// Phase 4: Migration flags
		MigrationMode:   flag.String("migration-mode", "", "Force migration mode: ssh-only, api-first, api-only-warn, api-only, auto"),
		ValidateConfig:  flag.Bool("validate-migration", false, "Validate configuration for migration and exit"),
		MigrationStatus: flag.Bool("migration-status", false, "Show migration status and exit"),
	}

	flag.Parse()
	return flags
}

// HandleSpecialFlags handles flags that should exit immediately
func HandleSpecialFlags(flags *Flags) bool {
	if *flags.PiholeSetup {
		return true // Caller should handle pihole setup
	}

	// Phase 4: Handle migration flags
	if *flags.ValidateConfig {
		handleMigrationValidation(flags)
		return true
	}

	if *flags.MigrationStatus {
		handleMigrationStatus(flags)
		return true
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

// handleMigrationValidation handles migration validation flag
func handleMigrationValidation(flags *Flags) {
	lgr := logger.New(&logger.Config{
		Level:        logger.LevelInfo,
		EnableColors: !*flags.NoColor,
		EnableEmojis: !*flags.NoEmoji,
		Component:    "migration",
	})

	configPath := config.GetConfigPath()
	if *flags.Config != "" {
		configPath = *flags.Config
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		lgr.Error("Failed to load configuration: %v", err)
		return
	}

	configMgr := config.NewMigrationConfigManager(configPath, lgr)
	report, err := configMgr.ValidateConfigForMigration(cfg)
	if err != nil {
		lgr.Error("Validation failed: %v", err)
		return
	}

	// Display validation results
	lgr.Info("=== Migration Validation Report ===")
	if report.IsReady {
		lgr.Info("✅ Configuration is ready for migration")
	} else {
		lgr.Error("❌ Configuration is not ready for migration")
	}
	lgr.Info("Readiness Score: %d/100", report.Score)

	for _, issue := range report.Issues {
		lgr.Error("Issue: %s", issue)
	}
	for _, warning := range report.Warnings {
		lgr.Warn("Warning: %s", warning)
	}
	for _, suggestion := range report.Suggestions {
		lgr.Info("Suggestion: %s", suggestion)
	}
}

// handleMigrationStatus handles migration status flag
func handleMigrationStatus(flags *Flags) {
	lgr := logger.New(&logger.Config{
		Level:        logger.LevelInfo,
		EnableColors: !*flags.NoColor,
		EnableEmojis: !*flags.NoEmoji,
		Component:    "migration",
	})

	configPath := config.GetConfigPath()
	if *flags.Config != "" {
		configPath = *flags.Config
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		lgr.Error("Failed to load configuration: %v", err)
		return
	}

	migrationMgr := interfaces.NewMigrationManager(cfg, lgr)
	migrationMgr.LogMigrationSummary()
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

	// Phase 4: Apply migration mode if specified
	if *flags.MigrationMode != "" {
		cfg.Pihole.MigrationMode = *flags.MigrationMode
	}
}

// ShowUsage displays usage information
func ShowUsage() {
	logger.Info("Usage: %s [options]\n", os.Args[0])
	logger.Info("Pi-hole Network Analyzer - Analyze Pi-hole DNS queries and network traffic\n")
	logger.Info("Examples:")
	logger.Info("  %s --pihole config.json           # Analyze live Pi-hole data", os.Args[0])
	logger.Info("  %s --pihole-setup                 # Setup Pi-hole SSH configuration", os.Args[0])
	logger.Info("  %s --migration-status             # Show SSH-to-API migration status", os.Args[0])
	logger.Info("  %s --validate-migration           # Validate configuration for migration", os.Args[0])
	logger.Info("  %s --migration-mode api-first     # Force API-first mode", os.Args[0])
	logger.Info("  %s --show-config                  # Show current configuration", os.Args[0])
	logger.Info("  %s --create-config                # Create default config file", os.Args[0])
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
		logger.Info("%s", colors.Header("Pi-hole Network Analyzer"))
		logger.Info("Analyzing Pi-hole data from: %s", cfg.Pihole.Host)

		// Phase 4: Show migration mode if set
		if cfg.Pihole.MigrationMode != "" {
			logger.Info("Migration mode: %s", cfg.Pihole.MigrationMode)
		}
	}
}
