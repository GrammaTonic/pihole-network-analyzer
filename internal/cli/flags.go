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
	// Network Analysis flags
	EnableNetworkAnalysis     *bool
	EnableDPI                 *bool
	EnableTrafficPatterns     *bool
	EnableSecurityAnalysis    *bool
	EnablePerformanceAnalysis *bool
	NetworkAnalysisConfig     *string
	// DNS Server flags
	EnableDNS       *bool
	DNSPort         *int
	DNSHost         *string
	DNSConfig       *string
	EnableDNSCache  *bool
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
		// Network Analysis flags
		EnableNetworkAnalysis:     flag.Bool("network-analysis", false, "Enable enhanced network analysis (DPI, patterns, security, performance)"),
		EnableDPI:                 flag.Bool("enable-dpi", false, "Enable deep packet inspection analysis"),
		EnableTrafficPatterns:     flag.Bool("enable-traffic-patterns", false, "Enable traffic pattern analysis"),
		EnableSecurityAnalysis:    flag.Bool("enable-security-analysis", false, "Enable security threat analysis"),
		EnablePerformanceAnalysis: flag.Bool("enable-performance-analysis", false, "Enable network performance analysis"),
		NetworkAnalysisConfig:     flag.String("network-config", "", "Path to network analysis configuration file"),
		// DNS Server flags
		EnableDNS:      flag.Bool("dns", false, "Enable DNS server with caching and super fast responses"),
		DNSPort:        flag.Int("dns-port", 5353, "Port for DNS server (default: 5353)"),
		DNSHost:        flag.String("dns-host", "0.0.0.0", "Host for DNS server (default: 0.0.0.0)"),
		DNSConfig:      flag.String("dns-config", "", "Path to DNS server configuration file"),
		EnableDNSCache: flag.Bool("dns-cache", true, "Enable DNS response caching (default: true)"),
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

	// Apply network analysis flags
	ApplyNetworkAnalysisFlags(flags, cfg)
	
	// Apply DNS server flags
	ApplyDNSFlags(flags, cfg)
}

// ApplyNetworkAnalysisFlags applies network analysis related flags to configuration
func ApplyNetworkAnalysisFlags(flags *Flags, cfg *types.Config) {
	// Enable network analysis if any of the sub-components are enabled
	if *flags.EnableNetworkAnalysis || *flags.EnableDPI || *flags.EnableTrafficPatterns ||
		*flags.EnableSecurityAnalysis || *flags.EnablePerformanceAnalysis {
		cfg.NetworkAnalysis.Enabled = true
	}

	// Configure individual components
	if *flags.EnableDPI {
		cfg.NetworkAnalysis.DeepPacketInspection.Enabled = true
	}

	if *flags.EnableTrafficPatterns {
		cfg.NetworkAnalysis.TrafficPatterns.Enabled = true
	}

	if *flags.EnableSecurityAnalysis {
		cfg.NetworkAnalysis.SecurityAnalysis.Enabled = true
	}

	if *flags.EnablePerformanceAnalysis {
		cfg.NetworkAnalysis.Performance.Enabled = true
	}
}

// ApplyDNSFlags applies DNS server related flags to configuration
func ApplyDNSFlags(flags *Flags, cfg *types.Config) {
	if *flags.EnableDNS {
		cfg.DNS.Enabled = true
		cfg.DNS.Host = *flags.DNSHost
		cfg.DNS.Port = *flags.DNSPort
		cfg.DNS.Cache.Enabled = *flags.EnableDNSCache
	}
}

// IsDNSEnabled returns true if DNS server is requested
func IsDNSEnabled(flags *Flags) bool {
	return *flags.EnableDNS
}

// GetDNSConfig extracts DNS configuration from flags
func GetDNSConfig(flags *Flags) map[string]any {
	return map[string]any{
		"enabled":     IsDNSEnabled(flags),
		"host":        *flags.DNSHost,
		"port":        *flags.DNSPort,
		"cache":       *flags.EnableDNSCache,
		"config_file": *flags.DNSConfig,
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

// IsNetworkAnalysisEnabled returns true if network analysis is requested
func IsNetworkAnalysisEnabled(flags *Flags) bool {
	return *flags.EnableNetworkAnalysis || *flags.EnableDPI || *flags.EnableTrafficPatterns ||
		*flags.EnableSecurityAnalysis || *flags.EnablePerformanceAnalysis
}

// GetNetworkAnalysisConfig extracts network analysis configuration from flags
func GetNetworkAnalysisConfig(flags *Flags) map[string]any {
	return map[string]any{
		"enabled":              IsNetworkAnalysisEnabled(flags),
		"dpi_enabled":          *flags.EnableDPI,
		"traffic_patterns":     *flags.EnableTrafficPatterns,
		"security_analysis":    *flags.EnableSecurityAnalysis,
		"performance_analysis": *flags.EnablePerformanceAnalysis,
		"config_file":          *flags.NetworkAnalysisConfig,
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
		if IsNetworkAnalysisEnabled(flags) {
			fmt.Println("🔬 Enhanced Network Analysis enabled:")
			if *flags.EnableDPI {
				fmt.Println("  • Deep Packet Inspection (DPI)")
			}
			if *flags.EnableTrafficPatterns {
				fmt.Println("  • Traffic Pattern Analysis")
			}
			if *flags.EnableSecurityAnalysis {
				fmt.Println("  • Security Threat Analysis")
			}
			if *flags.EnablePerformanceAnalysis {
				fmt.Println("  • Network Performance Analysis")
			}
		}
		if IsDNSEnabled(flags) {
			fmt.Printf("🚀 DNS server enabled on %s:%d\n", *flags.DNSHost, *flags.DNSPort)
			if *flags.EnableDNSCache {
				fmt.Println("  • DNS response caching enabled")
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
