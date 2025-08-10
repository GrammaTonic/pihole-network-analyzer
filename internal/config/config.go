package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
	"pihole-analyzer/internal/validation"
)

// DefaultConfig returns the default configuration
func DefaultConfig() *types.Config {
	return &types.Config{
		OnlineOnly: false,
		NoExclude:  false,
		TestMode:   false,

		Pihole: types.PiholeConfig{
			Host:        "192.168.1.100", // Changed from empty string to valid IP
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

		NetworkAnalysis: types.NetworkAnalysisConfig{
			Enabled: false,
			DeepPacketInspection: types.DPIConfig{
				Enabled:          false,
				AnalyzeProtocols: []string{"DNS_UDP", "DNS_TCP"},
				PacketSampling:   1.0,
				MaxPacketSize:    1500,
				BufferSize:       10000,
				TimeWindow:       "1h",
			},
			TrafficPatterns: types.TrafficPatternsConfig{
				Enabled:          false,
				PatternTypes:     []string{"bandwidth", "temporal", "client"},
				AnalysisWindow:   "2h",
				MinDataPoints:    10,
				PatternThreshold: 0.6,
				AnomalyDetection: true,
			},
			SecurityAnalysis: types.SecurityAnalysisConfig{
				Enabled:               false,
				ThreatDetection:       true,
				SuspiciousPatterns:    []string{"malware", "phishing", "botnet"},
				BlacklistDomains:      []string{},
				UnusualTrafficThresh:  0.75,
				PortScanDetection:     true,
				DNSTunnelingDetection: true,
			},
			Performance: types.NetworkPerformanceConfig{
				Enabled:             false,
				LatencyAnalysis:     true,
				BandwidthAnalysis:   true,
				ThroughputAnalysis:  true,
				PacketLossDetection: true,
				JitterAnalysis:      true,
				QualityThresholds: types.QualityThresholds{
					MaxLatency:    150.0,
					MinBandwidth:  5.0,
					MaxPacketLoss: 2.0,
					MaxJitter:     100.0,
				},
			},
		},

		Integrations: types.IntegrationsConfig{
			Enabled: false,
			Grafana: types.GrafanaConfig{
				Enabled:      false,
				URL:          "http://localhost:3000",
				Organization: "main",
				DataSource: types.DataSourceConfig{
					CreateIfNotExists: true,
					Name:              "pihole-analyzer-prometheus",
					Type:              "prometheus",
					URL:               "http://localhost:9090",
					Access:            "proxy",
				},
				Dashboards: types.DashboardConfig{
					AutoProvision:     false,
					FolderName:        "Pi-hole Network Analyzer",
					OverwriteExisting: true,
					Tags:              []string{"pihole", "network", "dns"},
				},
				AlertIntegration: types.GrafanaAlertConfig{
					Enabled:         false,
					DefaultSeverity: "warning",
				},
				Timeout:    30,
				VerifyTLS:  true,
				RetryCount: 3,
			},
			Loki: types.LokiConfig{
				Enabled:      false,
				URL:          "http://localhost:3100",
				BatchSize:    100,
				BatchTimeout: "10s",
				BufferSize:   1000,
				StaticLabels: map[string]string{
					"service": "pihole-analyzer",
					"env":     "production",
				},
				DynamicLabels: []string{"level", "component"},
				Timeout:       30,
				VerifyTLS:     true,
				RetryCount:    3,
				RetryInterval: "5s",
			},
			Prometheus: types.PrometheusExtConfig{
				Enabled: false,
				PushGateway: types.PushGatewayConfig{
					Enabled:  false,
					URL:      "http://localhost:9091",
					Job:      "pihole-analyzer",
					Instance: "localhost",
					Timeout:  10,
					Interval: "30s",
				},
				RemoteWrite: types.RemoteWriteConfig{
					Enabled:   false,
					Timeout:   30,
					BatchSize: 100,
				},
				ServiceDiscovery: types.ServiceDiscoveryConfig{
					Enabled:         false,
					Type:            "static",
					RefreshInterval: "60s",
				},
				ExternalLabels: map[string]string{
					"service": "pihole-analyzer",
				},
			},
			Generic: []types.GenericIntegrationConfig{},
		},
	}
}

// LoadConfig loads configuration from file, falling back to defaults
func LoadConfig(configPath string) (*types.Config, error) {
	log := logger.Component("config")
	config := DefaultConfig()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.InfoFields("Config file not found, using defaults", map[string]any{
			"config_path": configPath,
		})
		// Still apply environment variables to defaults
		LoadEnvironmentVariables(config)
		return config, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.ErrorFields("Failed to read config file", map[string]any{
			"config_path": configPath,
			"error":       err.Error(),
		})
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	// Parse JSON
	if err := json.Unmarshal(data, config); err != nil {
		log.ErrorFields("Failed to parse config file", map[string]any{
			"config_path": configPath,
			"error":       err.Error(),
		})
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	// Apply environment variables (override config file values)
	LoadEnvironmentVariables(config)

	// Validate configuration
	validator := validation.NewValidator(log)
	result := validator.ValidateConfig(config)

	if !result.Valid {
		log.ErrorFields("Configuration validation failed", map[string]any{
			"config_path":   configPath,
			"error_count":   len(result.Errors),
			"warning_count": len(result.Warnings),
		})

		// Apply defaults to fix critical issues
		validator.ApplyDefaults(config)

		// Re-validate after applying defaults
		result = validator.ValidateConfig(config)
		if !result.Valid {
			return nil, fmt.Errorf("configuration validation failed even after applying defaults")
		}

		log.InfoFields("Configuration fixed with defaults", map[string]any{
			"config_path": configPath,
		})
	}

	log.Success("Configuration loaded and validated successfully from %s", configPath)
	return config, nil
}

// LoadEnvironmentVariables merges environment variables into the configuration
// Environment variables override config file values but are overridden by CLI flags
func LoadEnvironmentVariables(config *types.Config) {
	log := logger.Component("config-env")
	
	envCount := 0
	
	// Pi-hole configuration
	if host := os.Getenv("PIHOLE_HOST"); host != "" {
		config.Pihole.Host = host
		envCount++
	}
	if port := os.Getenv("PIHOLE_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Pihole.Port = p
			envCount++
		}
	}
	if enabled := os.Getenv("PIHOLE_API_ENABLED"); enabled != "" {
		config.Pihole.APIEnabled = strings.ToLower(enabled) == "true"
		envCount++
	}
	if password := os.Getenv("PIHOLE_API_PASSWORD"); password != "" {
		config.Pihole.APIPassword = password
		envCount++
	}
	if https := os.Getenv("PIHOLE_USE_HTTPS"); https != "" {
		config.Pihole.UseHTTPS = strings.ToLower(https) == "true"
		envCount++
	}
	if timeout := os.Getenv("PIHOLE_API_TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			config.Pihole.APITimeout = t
			envCount++
		}
	}
	if verifySSL := os.Getenv("PIHOLE_VERIFY_SSL"); verifySSL != "" {
		// This maps to UseHTTPS verification - if false, we'd want to allow insecure HTTPS
		// For now, we'll map it to UseHTTPS
		config.Pihole.UseHTTPS = strings.ToLower(verifySSL) == "true"
		envCount++
	}
	
	// Logging configuration
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Logging.Level = strings.ToUpper(level)
		envCount++
	}
	if colors := os.Getenv("LOG_ENABLE_COLORS"); colors != "" {
		config.Logging.EnableColors = strings.ToLower(colors) == "true"
		envCount++
	}
	if emojis := os.Getenv("LOG_ENABLE_EMOJIS"); emojis != "" {
		config.Logging.EnableEmojis = strings.ToLower(emojis) == "true"
		envCount++
	}
	
	// Analysis configuration
	if onlineOnly := os.Getenv("ANALYSIS_ONLINE_ONLY"); onlineOnly != "" {
		config.OnlineOnly = strings.ToLower(onlineOnly) == "true"
		envCount++
	}
	
	// Web configuration
	if webEnabled := os.Getenv("WEB_ENABLED"); webEnabled != "" {
		config.Web.Enabled = strings.ToLower(webEnabled) == "true"
		envCount++
	}
	if webHost := os.Getenv("WEB_HOST"); webHost != "" {
		config.Web.Host = webHost
		envCount++
	}
	if webPort := os.Getenv("WEB_PORT"); webPort != "" {
		if p, err := strconv.Atoi(webPort); err == nil {
			config.Web.Port = p
			envCount++
		}
	}
	if daemon := os.Getenv("WEB_DAEMON_MODE"); daemon != "" {
		config.Web.DaemonMode = strings.ToLower(daemon) == "true"
		envCount++
	}
	
	// Metrics configuration
	if metricsEnabled := os.Getenv("METRICS_ENABLED"); metricsEnabled != "" {
		config.Metrics.Enabled = strings.ToLower(metricsEnabled) == "true"
		envCount++
	}
	if metricsHost := os.Getenv("METRICS_HOST"); metricsHost != "" {
		config.Metrics.Host = metricsHost
		envCount++
	}
	if metricsPort := os.Getenv("METRICS_PORT"); metricsPort != "" {
		config.Metrics.Port = metricsPort
		envCount++
	}
	
	// Container optimization environment variables (Go runtime)
	if memLimit := os.Getenv("GOMEMLIMIT"); memLimit != "" {
		log.InfoFields("Go memory limit set", map[string]any{"limit": memLimit})
	}
	if maxProcs := os.Getenv("GOMAXPROCS"); maxProcs != "" {
		log.InfoFields("Go max processors set", map[string]any{"procs": maxProcs})
	}
	
	if envCount > 0 {
		log.InfoFields("Environment variables loaded", map[string]any{
			"variables_applied": envCount,
		})
	}
}

// SaveConfig saves the current configuration to file
func SaveConfig(config *types.Config, configPath string) error {
	log := logger.Component("config")

	// Validate configuration before saving
	validator := validation.NewValidator(log)
	result := validator.ValidateConfig(config)

	if !result.Valid {
		log.ErrorFields("Cannot save invalid configuration", map[string]any{
			"config_path": configPath,
			"error_count": len(result.Errors),
		})
		return fmt.Errorf("configuration validation failed, cannot save")
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.ErrorFields("Failed to create config directory", map[string]any{
			"directory": dir,
			"error":     err.Error(),
		})
		return fmt.Errorf("error creating config directory: %v", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.ErrorFields("Failed to marshal config", map[string]any{
			"error": err.Error(),
		})
		return fmt.Errorf("error marshaling config: %v", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		log.ErrorFields("Failed to write config file", map[string]any{
			"config_path": configPath,
			"error":       err.Error(),
		})
		return fmt.Errorf("error writing config file: %v", err)
	}

	log.Success("Configuration saved to %s", configPath)
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
	log := logger.Component("config")
	config := DefaultConfig()

	log.InfoFields("Creating default configuration file", map[string]any{
		"config_path": configPath,
	})

	return SaveConfig(config, configPath)
}

// ShowConfig displays the current configuration
func ShowConfig(config *types.Config) {
	log := logger.Component("config")

	log.Info("\nCurrent Configuration:")
	log.Info("======================")
	log.InfoFields("Global settings", map[string]any{
		"online_only": config.OnlineOnly,
		"no_exclude":  config.NoExclude,
		"test_mode":   config.TestMode,
		"quiet":       config.Quiet,
	})

	log.InfoFields("Output settings", map[string]any{
		"max_clients":    config.Output.MaxClients,
		"max_domains":    config.Output.MaxDomains,
		"save_reports":   config.Output.SaveReports,
		"report_dir":     config.Output.ReportDir,
		"verbose_output": config.Output.VerboseOutput,
	})

	piholeInfo := map[string]any{
		"host":        config.Pihole.Host,
		"port":        config.Pihole.Port,
		"api_enabled": config.Pihole.APIEnabled,
		"use_https":   config.Pihole.UseHTTPS,
		"api_timeout": config.Pihole.APITimeout,
	}

	if config.Pihole.APIPassword != "" {
		piholeInfo["api_password"] = "***configured***"
	} else {
		piholeInfo["api_password"] = "not set"
	}

	log.InfoFields("Pi-hole settings", piholeInfo)

	log.InfoFields("Exclusion settings", map[string]any{
		"exclude_networks_count": len(config.Exclusions.ExcludeNetworks),
		"exclude_ips_count":      len(config.Exclusions.ExcludeIPs),
		"exclude_hosts_count":    len(config.Exclusions.ExcludeHosts),
		"exclude_networks":       config.Exclusions.ExcludeNetworks,
		"exclude_ips":            config.Exclusions.ExcludeIPs,
		"exclude_hosts":          config.Exclusions.ExcludeHosts,
	})

	log.InfoFields("Logging settings", map[string]any{
		"level":          config.Logging.Level,
		"output_file":    config.Logging.OutputFile,
		"enable_colors":  config.Logging.EnableColors,
		"enable_emojis":  config.Logging.EnableEmojis,
		"show_timestamp": config.Logging.ShowTimestamp,
		"show_caller":    config.Logging.ShowCaller,
	})
}

// GetConfigPath returns the default configuration file path
func GetConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".pihole-analyzer", "config.json")
}
