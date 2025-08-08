package cli

import (
	"flag"
	"fmt"
	"os"

	"pihole-network-analyzer/internal/colors"
	"pihole-network-analyzer/internal/config"
	"pihole-network-analyzer/internal/types"
)

// Flags represents command-line flags
type Flags struct {
	OnlineOnly   *bool
	NoExclude    *bool
	Pihole       *string
	PiholeSetup  *bool
	Test         *bool
	TestMode     *bool
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
		Test:         flag.Bool("test", false, "Run test suite with mock data"),
		TestMode:     flag.Bool("test-mode", false, "Enable test mode for development (uses mock data)"),
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

// GetCSVFile returns the CSV file from command-line arguments
func GetCSVFile() string {
	args := flag.Args()
	if len(args) > 0 {
		return args[0]
	}
	return ""
}

// ApplyFlags applies command-line flags to configuration
func ApplyFlags(flags *Flags, cfg *types.Config) {
	if cfg == nil {
		return
	}

	// Apply boolean flags
	if *flags.OnlineOnly {
		cfg.OnlineOnly = true
	}
	if *flags.NoExclude {
		cfg.NoExclude = true
	}
	if *flags.TestMode {
		cfg.TestMode = true
	}
	if *flags.Quiet {
		cfg.Quiet = true
	}

	// Apply color/emoji flags
	if *flags.NoColor {
		cfg.Output.Colors = false
		colors.DisableColors()
	}
	if *flags.NoEmoji {
		cfg.Output.Emojis = false
		colors.DisableEmojis()
	}
}

// ShowUsage displays usage information
func ShowUsage() {
	fmt.Printf("%s\n", colors.Header("Pi-hole Network Analyzer"))
	fmt.Printf("Usage: %s [options] [csv-file]\n\n", os.Args[0])

	fmt.Printf("%s:\n", colors.BoldCyan("Options"))
	flag.PrintDefaults()

	fmt.Printf("\n%s:\n", colors.BoldCyan("Examples"))
	fmt.Printf("  %s                          # Analyze test.csv (default)\n", os.Args[0])
	fmt.Printf("  %s queries.csv              # Analyze specific CSV file\n", os.Args[0])
	fmt.Printf("  %s --online-only            # Show only online clients\n", os.Args[0])
	fmt.Printf("  %s --pihole config.json     # Analyze Pi-hole live data\n", os.Args[0])
	fmt.Printf("  %s --pihole-setup           # Setup Pi-hole SSH config\n", os.Args[0])
	fmt.Printf("  %s --test                   # Run with mock test data\n", os.Args[0])
	fmt.Printf("  %s --no-color --quiet       # Plain output for scripting\n", os.Args[0])

	fmt.Printf("\n%s:\n", colors.BoldCyan("Configuration"))
	fmt.Printf("  Default config: ~/.pihole-analyzer/config.json\n")
	fmt.Printf("  Create config:  %s --create-config\n", os.Args[0])
	fmt.Printf("  Show config:    %s --show-config\n", os.Args[0])
}

// HandleSpecialFlags handles flags that should exit immediately
func HandleSpecialFlags(flags *Flags) bool {
	if *flags.PiholeSetup {
		return true // Caller should handle pihole setup
	}

	if *flags.ShowConfig {
		cfg, err := config.LoadConfig(*flags.Config)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return true
		}

		fmt.Printf("%s\n", colors.Header("Current Configuration"))
		config.ShowConfig(cfg)
		return true
	}

	if *flags.CreateConfig {
		err := config.CreateDefaultConfigFile(*flags.Config)
		if err != nil {
			fmt.Printf("Error creating config: %v\n", err)
		} else {
			fmt.Printf("Default configuration created successfully\n")
		}
		return true
	}

	return false
}

// ValidateInput validates command-line input
func ValidateInput(flags *Flags, csvFile string) error {
	// If analyzing Pi-hole data, config file is required
	if *flags.Pihole != "" {
		if *flags.Pihole == "" {
			return fmt.Errorf("Pi-hole config file required when using --pihole")
		}
		if _, err := os.Stat(*flags.Pihole); os.IsNotExist(err) {
			return fmt.Errorf("Pi-hole config file not found: %s", *flags.Pihole)
		}
	}

	// If CSV file specified, check if it exists
	if csvFile != "" && !*flags.Test && !*flags.TestMode {
		if _, err := os.Stat(csvFile); os.IsNotExist(err) {
			return fmt.Errorf("CSV file not found: %s", csvFile)
		}
	}

	return nil
}

// PrintStartupInfo prints startup information
func PrintStartupInfo(flags *Flags, cfg *types.Config) {
	if cfg != nil && cfg.Quiet {
		return
	}

	fmt.Printf("%s\n", colors.Header("Pi-hole Network Analyzer"))

	if *flags.Test || *flags.TestMode {
		fmt.Printf("%s\n", colors.Info("Running in test mode with mock data"))
	}

	if *flags.OnlineOnly {
		fmt.Printf("%s\n", colors.Info("Showing only online clients"))
	}

	if *flags.NoExclude {
		fmt.Printf("%s\n", colors.Warning("Default exclusions disabled"))
	}

	if *flags.Pihole != "" {
		fmt.Printf("%s %s\n", colors.Info("Analyzing Pi-hole data from:"), *flags.Pihole)
	}
}

// GetDefaultCSVFile returns the default CSV file name
func GetDefaultCSVFile() string {
	return "test.csv"
}
