package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"pihole-analyzer/internal/analyzer"
	"pihole-analyzer/internal/cli"
	"pihole-analyzer/internal/config"
	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/metrics"
	"pihole-analyzer/internal/reporting"
	"pihole-analyzer/internal/types"
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

	// Validate CLI input
	if err := cli.ValidateInput(flags); err != nil {
		appLogger.Error("Input validation failed: %v", err)
		os.Exit(1)
	}

	// Print startup information
	cli.PrintStartupInfo(flags, cfg)

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
