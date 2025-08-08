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

	// Try enhanced analysis first
	result, err := analyzer.AnalyzePiholeDataWithMigration(ctx, config, appLogger)
	if err != nil {
		// Fallback to traditional analysis
		appLogger.Warn("⚠️  Enhanced analysis failed, using traditional method: %v", err)
		return runTraditionalAnalysis(configFile, config, appLogger)
	}

	// Display enhanced results
	if err := displayEnhancedResults(result, config, appLogger); err != nil {
		return fmt.Errorf("failed to display results: %w", err)
	}

	if !config.Quiet {
		appLogger.Info("✅ Analysis completed successfully!")
	}

	return nil
}

func runTraditionalAnalysis(configFile string, config *types.Config, appLogger *logger.Logger) error {
	// Use traditional analyzer as fallback
	clientStats, err := analyzer.AnalyzePiholeData(configFile, config)
	if err != nil {
		return fmt.Errorf("traditional analysis failed: %w", err)
	}

	if !config.Quiet {
		appLogger.Info("📊 Generating report...")
	}

	// Display traditional results
	reporting.DisplayResultsWithConfig(clientStats, config)
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

	// Display migration status if in transition mode
	if result.MigrationStatus != "" {
		appLogger.Info("🔄 Migration Status: %s", result.MigrationStatus)
	}

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
