package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pihole-analyzer/internal/analyzer"
	"pihole-analyzer/internal/cli"
	"pihole-analyzer/internal/config"
	"pihole-analyzer/internal/dhcp"
	"pihole-analyzer/internal/interfaces"
	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/metrics"
	"pihole-analyzer/internal/reporting"
	"pihole-analyzer/internal/types"
	"pihole-analyzer/internal/web"
)

func main() {
	// Parse command-line flags using CLI package
	flags := cli.ParseFlags()

	// Handle special flags that should exit immediately
	if cli.HandleSpecialFlags(flags) {
		return
	}

	// Load configuration
	configPath := config.GetConfigPath()
	if *flags.Config != "" {
		configPath = *flags.Config
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize structured logger
	loggerConfig := &logger.Config{
		Level:         logger.LogLevel(cfg.Logging.Level),
		EnableColors:  cfg.Logging.EnableColors && !*flags.NoColor,
		EnableEmojis:  cfg.Logging.EnableEmojis && !*flags.NoEmoji,
		OutputFile:    cfg.Logging.OutputFile,
		ShowTimestamp: cfg.Logging.ShowTimestamp,
		Component:     "pihole-analyzer",
	}
	appLogger := logger.New(loggerConfig)

	// Apply CLI flags to configuration
	cli.ApplyFlags(flags, cfg)

	// Apply web flags to configuration
	if cli.IsWebModeEnabled(flags) {
		cfg.Web.Enabled = true
		cfg.Web.Port = *flags.WebPort
		cfg.Web.Host = *flags.WebHost
		cfg.Web.DaemonMode = *flags.DaemonMode
	}

	// Validate CLI input
	if err := cli.ValidateInput(flags); err != nil {
		appLogger.Error("Input validation failed: %v", err)
		os.Exit(1)
	}

	// Print startup information
	cli.PrintStartupInfo(flags, cfg)

	// Handle web mode
	if cfg.Web.Enabled {
		if err := runWebMode(flags, cfg, appLogger); err != nil {
			appLogger.Error("Error running web mode: %v", err)
			os.Exit(1)
		}
		return
	}

	// Handle Pi-hole specific operations
	if *flags.Pihole != "" {
		if err := analyzePihole(*flags.Pihole, cfg, appLogger); err != nil {
			appLogger.Error("Error analyzing Pi-hole data: %v", err)
			os.Exit(1)
		}
		return
	}

	// Handle test mode
	if cfg.TestMode {
		appLogger.Info("Running in test mode with mock data")
		if err := runTestMode(appLogger); err != nil {
			appLogger.Error("Error running test mode: %v", err)
			os.Exit(1)
		}
		return
	}

	appLogger.Info("No operation specified. Use --help for usage information.")
}

func analyzePihole(configFile string, config *types.Config, appLogger *logger.Logger) error {
	ctx := context.Background()

	if !config.Quiet {
		appLogger.Info("🚀 Starting Pi-hole analysis...")
	}

	// Initialize metrics collector if enabled
	var metricsCollector *metrics.Collector
	var metricsServer *metrics.Server

	if config.Metrics.Enabled && config.Metrics.CollectMetrics {
		metricsCollector = metrics.New(appLogger.GetSlogger())

		if config.Metrics.EnableEndpoint {
			// Start metrics server in background
			serverConfig := metrics.ServerConfig{
				Port:    config.Metrics.Port,
				Host:    config.Metrics.Host,
				Enabled: config.Metrics.EnableEndpoint,
			}
			metricsServer = metrics.NewServer(serverConfig, metricsCollector, appLogger.GetSlogger())
			metricsServer.StartInBackground()

			// Ensure server is stopped when function exits
			defer func() {
				if metricsServer != nil {
					metricsServer.Stop(ctx)
				}
			}()
		}
	}

	// Try enhanced analysis first
	result, err := analyzer.AnalyzePiholeData(ctx, config, appLogger, metricsCollector)
	if err != nil {
		// Record error in metrics if available
		if metricsCollector != nil {
			metricsCollector.RecordError("enhanced_analysis_failed")
		}

		// Fallback to traditional analysis
		appLogger.Warn("⚠️  Enhanced analysis failed, using traditional method: %v", err)
		return runTraditionalAnalysis(configFile, config, appLogger, metricsCollector)
	}

	// Display enhanced results
	if err := displayEnhancedResults(result, config, appLogger); err != nil {
		if metricsCollector != nil {
			metricsCollector.RecordError("display_results_failed")
		}
		return fmt.Errorf("failed to display results: %w", err)
	}

	if !config.Quiet {
		appLogger.Info("✅ Analysis completed successfully!")
		if metricsServer != nil {
			appLogger.Info("📊 Metrics endpoint available at: http://%s:%s/metrics",
				config.Metrics.Host, config.Metrics.Port)
		}
	}

	return nil
}

func runTraditionalAnalysis(configFile string, config *types.Config, appLogger *logger.Logger, metricsCollector *metrics.Collector) error {
	ctx := context.Background()

	// Use traditional analyzer as fallback
	result, err := analyzer.AnalyzePiholeData(ctx, config, appLogger, metricsCollector)
	if err != nil {
		if metricsCollector != nil {
			metricsCollector.RecordError("traditional_analysis_failed")
		}
		return fmt.Errorf("traditional analysis failed: %w", err)
	}

	if !config.Quiet {
		appLogger.Info("📊 Generating report...")
	}

	// Display traditional results
	reporting.DisplayResultsWithConfig(result.ClientStats, config)
	return nil
}

func displayEnhancedResults(result *types.AnalysisResult, cfg *types.Config, appLogger *logger.Logger) error {
	// Display analysis summary
	appLogger.Info("=== Enhanced Analysis Results ===")
	appLogger.Info("Analysis Mode: %s", result.AnalysisMode)
	appLogger.Info("Data Source: %s", result.DataSourceType)
	appLogger.Info("Total Queries: %d", result.TotalQueries)
	appLogger.Info("Unique Clients: %d", result.UniqueClients)
	appLogger.Info("Analysis Time: %s", result.Timestamp)

	if len(result.NetworkDevices) > 0 {
		appLogger.Info("Network Devices: %d", len(result.NetworkDevices))
	}

	// Use traditional reporting for client statistics display
	reporting.DisplayResultsWithConfig(result.ClientStats, cfg)

	// Show performance metrics
	if result.Performance != nil {
		appLogger.Info("📈 Performance Metrics:")
		appLogger.Info("  Average Response Time: %.2fms", result.Performance.AverageResponseTime)
		appLogger.Info("  Queries Per Second: %.2f", result.Performance.QueriesPerSecond)
	}

	return nil
}

func runTestMode(appLogger *logger.Logger) error {
	// This would implement test mode functionality
	appLogger.Info("Test mode is not yet implemented")
	return nil
}

func runWebMode(flags *cli.Flags, cfg *types.Config, appLogger *logger.Logger) error {
	webLogger := appLogger.Component("web-mode")

	webLogger.InfoFields("Starting web mode", map[string]any{
		"port":        cfg.Web.Port,
		"host":        cfg.Web.Host,
		"daemon_mode": cfg.Web.DaemonMode,
	})

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create data source based on configuration
	var dataSource interfaces.DataSource
	var err error

	if *flags.Pihole != "" {
		// Use Pi-hole API data source
		webLogger.Info("Configuring Pi-hole API data source")

		// Load Pi-hole configuration
		piholeConfig, err := loadPiholeConfig(*flags.Pihole)
		if err != nil {
			return fmt.Errorf("failed to load Pi-hole config: %w", err)
		}

		// Update main config with Pi-hole settings
		cfg.Pihole = *piholeConfig

		// Create Pi-hole data source using factory
		factory := interfaces.NewDataSourceFactory(webLogger)
		dataSource, err = factory.CreateDataSource(cfg)
		if err != nil {
			return fmt.Errorf("failed to create Pi-hole data source: %w", err)
		}

		webLogger.InfoFields("Pi-hole data source configured", map[string]any{
			"host":      cfg.Pihole.Host,
			"use_https": cfg.Pihole.UseHTTPS,
		})
	} else {
		// Use mock data source for testing
		webLogger.Info("Using mock data source for web interface demonstration")
		dataSource = web.NewMockDataSourceForProduction()
	}

	// Create data source adapter for web interface
	adapter, err := web.NewDataSourceAdapter(dataSource, cfg, webLogger)
	if err != nil {
		return fmt.Errorf("failed to create data source adapter: %w", err)
	}

	// Create web server configuration
	webConfig := &web.Config{
		Port:         cfg.Web.Port,
		Host:         cfg.Web.Host,
		EnableWeb:    cfg.Web.Enabled,
		ReadTimeout:  time.Duration(cfg.Web.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Web.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Web.IdleTimeout) * time.Second,
	}

	// Create and start web server
	server, err := web.NewServer(webConfig, adapter, webLogger)
	if err != nil {
		return fmt.Errorf("failed to create web server: %w", err)
	}
	
	// Create and integrate DHCP server if enabled
	if cfg.DHCP.Enabled {
		webLogger.Info("DHCP server enabled, creating DHCP server")
		
		dhcpLogger := webLogger.Component("dhcp-server")
		dhcpServer, err := createDHCPServer(cfg, dhcpLogger)
		if err != nil {
			webLogger.Error("Failed to create DHCP server: %v", err)
			// Continue without DHCP server rather than failing completely
		} else {
			// Register DHCP routes with the web server
			server.RegisterDHCPRoutes(dhcpServer)
			
			// Start DHCP server in background
			go func() {
				if err := dhcpServer.Start(ctx); err != nil {
					dhcpLogger.Error("DHCP server failed to start: %v", err)
				}
			}()
			
			// Ensure DHCP server is stopped when context is cancelled
			go func() {
				<-ctx.Done()
				dhcpLogger.Info("Stopping DHCP server")
				stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer stopCancel()
				
				if err := dhcpServer.Stop(stopCtx); err != nil {
					dhcpLogger.Error("Error stopping DHCP server: %v", err)
				}
			}()
			
			webLogger.Success("DHCP server integrated and starting")
		}
	}

	// Start server in a goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		webLogger.Success("Web interface starting on http://%s:%d", cfg.Web.Host, cfg.Web.Port)

		if cfg.Web.DaemonMode {
			webLogger.Info("Running in daemon mode - server will run until stopped")
		} else {
			webLogger.Info("Running in web mode - use Ctrl+C to stop")
		}

		err := server.Start(ctx)
		if err != nil {
			serverErrChan <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case sig := <-sigChan:
		webLogger.Info("Received shutdown signal: %v", sig)
		cancel()

		// Wait a moment for graceful shutdown
		time.Sleep(1 * time.Second)
		webLogger.Success("Web server shutdown complete")

	case err := <-serverErrChan:
		webLogger.Error("Web server error: %v", err)
		return err
	}

	return nil
}

func loadPiholeConfig(configFile string) (*types.PiholeConfig, error) {
	// This is a simplified version - in a real implementation,
	// you would parse the Pi-hole configuration file
	return &types.PiholeConfig{
		Host:        "localhost",
		Port:        80,
		APIEnabled:  true,
		APIPassword: "",
		UseHTTPS:    false,
		APITimeout:  30,
	}, nil
}

func createDHCPServer(cfg *types.Config, logger *logger.Logger) (dhcp.DHCPServer, error) {
	// Validate DHCP configuration
	if err := dhcp.ValidateDHCPConfig(&cfg.DHCP); err != nil {
		return nil, fmt.Errorf("invalid DHCP configuration: %w", err)
	}
	
	// Create DHCP server
	dhcpServer, err := dhcp.NewServer(&cfg.DHCP, logger.GetSlogger())
	if err != nil {
		return nil, fmt.Errorf("failed to create DHCP server: %w", err)
	}
	
	logger.InfoFields("DHCP server created successfully", map[string]any{
		"interface":      cfg.DHCP.Interface,
		"listen_address": cfg.DHCP.ListenAddress,
		"port":          cfg.DHCP.Port,
		"pool_start":    cfg.DHCP.Pool.StartIP,
		"pool_end":      cfg.DHCP.Pool.EndIP,
		"lease_time":    cfg.DHCP.LeaseTime,
	})
	
	return dhcpServer, nil
}
