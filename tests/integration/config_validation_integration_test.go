package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"pihole-analyzer/internal/config"
	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
	"pihole-analyzer/internal/validation"
)

// TestConfigValidationIntegration tests the complete config validation workflow
func TestConfigValidationIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	// Test 1: Valid configuration workflow
	t.Run("ValidConfigurationWorkflow", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "valid_config.json")

		// Create and save a valid config
		validConfig := validation.GetDefaultValidationConfig()
		err := config.SaveConfig(validConfig, configPath)
		if err != nil {
			t.Fatalf("Failed to save valid config: %v", err)
		}

		// Load and validate the config
		loadedConfig, err := config.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to load valid config: %v", err)
		}

		// Verify validation passes
		log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
		if !validation.ValidateAndWarn(loadedConfig, log) {
			t.Error("Valid configuration should pass validation")
		}
	})

	// Test 2: Invalid configuration with auto-fix
	t.Run("InvalidConfigurationWithAutoFix", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "invalid_config.json")

		// Create an invalid config
		invalidConfig := &types.Config{
			Pihole: types.PiholeConfig{
				Host: "", // Invalid
				Port: -1, // Invalid
			},
			Output: types.OutputConfig{
				MaxClients: 0,  // Invalid
				MaxDomains: -5, // Invalid
			},
			Logging: types.LoggingConfig{
				Level: "INVALID", // Invalid
			},
		}

		// Manually write the invalid config (bypassing validation)
		data, err := json.MarshalIndent(invalidConfig, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal invalid config: %v", err)
		}

		err = os.WriteFile(configPath, data, 0600)
		if err != nil {
			t.Fatalf("Failed to write invalid config: %v", err)
		}

		// Load config - should auto-fix with defaults
		loadedConfig, err := config.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to load invalid config: %v", err)
		}

		// Verify auto-fix worked
		if loadedConfig.Pihole.Host == "" {
			t.Error("Host should have been auto-fixed")
		}
		if loadedConfig.Pihole.Port <= 0 {
			t.Error("Port should have been auto-fixed")
		}
		if loadedConfig.Output.MaxClients <= 0 {
			t.Error("MaxClients should have been auto-fixed")
		}
		if loadedConfig.Output.MaxDomains <= 0 {
			t.Error("MaxDomains should have been auto-fixed")
		}
	})

	// Test 3: Configuration validation with structured logging
	t.Run("ConfigurationValidationLogging", func(t *testing.T) {
		// Create a logger with specific configuration
		log := logger.New(&logger.Config{
			Level:        logger.LevelDebug,
			EnableColors: false,
			EnableEmojis: false,
			Component:    "integration-test",
		})

		// Test with various configurations
		testConfigs := []struct {
			name       string
			config     *types.Config
			shouldPass bool
		}{
			{
				name:       "ValidConfig",
				config:     validation.GetDefaultValidationConfig(),
				shouldPass: true,
			},
			{
				name: "InvalidHost",
				config: &types.Config{
					Pihole: types.PiholeConfig{
						Host: "",
						Port: 80,
					},
					Output: types.OutputConfig{
						MaxClients: 20,
						MaxDomains: 10,
					},
					Logging: types.LoggingConfig{
						Level: "INFO",
					},
				},
				shouldPass: false,
			},
		}

		for _, tc := range testConfigs {
			t.Run(tc.name, func(t *testing.T) {
				validator := validation.NewValidator(log)
				result := validator.ValidateConfig(tc.config)

				if result.Valid != tc.shouldPass {
					t.Errorf("Expected validation result %v, got %v", tc.shouldPass, result.Valid)
				}

				// Test structured logging output
				if tc.shouldPass && len(result.Errors) > 0 {
					t.Errorf("Valid config should have no errors, got %d", len(result.Errors))
				}
				if !tc.shouldPass && len(result.Errors) == 0 {
					t.Error("Invalid config should have errors")
				}
			})
		}
	})

	// Test 4: Default application workflow
	t.Run("DefaultApplicationWorkflow", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "app_config.json")

		// Step 1: Create default config file
		err := config.CreateDefaultConfigFile(configPath)
		if err != nil {
			t.Fatalf("Failed to create default config file: %v", err)
		}

		// Step 2: Load config (simulating application startup)
		appConfig, err := config.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to load config during app startup: %v", err)
		}

		// Step 3: Validate config is usable
		if appConfig.Pihole.Host == "" {
			t.Error("Application config should have a valid host")
		}
		if appConfig.Pihole.Port <= 0 {
			t.Error("Application config should have a valid port")
		}
		if appConfig.Output.MaxClients <= 0 {
			t.Error("Application config should have valid max clients")
		}

		// Step 4: Show config (simulating --show-config flag)
		config.ShowConfig(appConfig)

		// Step 5: Modify and save config
		appConfig.Pihole.Host = "192.168.1.200"
		appConfig.Output.MaxClients = 100

		err = config.SaveConfig(appConfig, configPath)
		if err != nil {
			t.Fatalf("Failed to save modified config: %v", err)
		}

		// Step 6: Reload and verify changes
		reloadedConfig, err := config.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to reload modified config: %v", err)
		}

		if reloadedConfig.Pihole.Host != "192.168.1.200" {
			t.Error("Host modification was not persisted")
		}
		if reloadedConfig.Output.MaxClients != 100 {
			t.Error("MaxClients modification was not persisted")
		}
	})
}

// TestValidationErrorHandling tests error handling in validation
func TestValidationErrorHandling(t *testing.T) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := validation.NewValidator(log)

	// Test with various invalid configurations
	testCases := []struct {
		name           string
		config         *types.Config
		expectedErrors []string
	}{
		{
			name: "MultipleErrors",
			config: &types.Config{
				Pihole: types.PiholeConfig{
					Host: "",    // Error: empty host
					Port: 70000, // Error: invalid port
				},
				Output: types.OutputConfig{
					MaxClients: -5,        // Error: negative max clients
					MaxDomains: 0,         // Error: zero max domains
					Format:     "invalid", // Error: invalid format
				},
				Exclusions: types.ExclusionConfig{
					ExcludeNetworks: []string{"invalid_cidr"},    // Error: invalid CIDR
					ExcludeIPs:      []string{"999.999.999.999"}, // Error: invalid IP
				},
				Logging: types.LoggingConfig{
					Level: "INVALID", // Error: invalid log level
				},
			},
			expectedErrors: []string{
				"pihole.host",
				"pihole.port",
				"output.max_clients",
				"output.max_domains_display",
				"output.format",
				"exclusions.exclude_networks[0]",
				"exclusions.exclude_ips[0]",
				"logging.level",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.ValidateConfig(tc.config)

			if result.Valid {
				t.Error("Configuration with multiple errors should be invalid")
			}

			if len(result.Errors) != len(tc.expectedErrors) {
				t.Errorf("Expected %d errors, got %d", len(tc.expectedErrors), len(result.Errors))
			}

			// Check that all expected error fields are present
			errorFields := make(map[string]bool)
			for _, err := range result.Errors {
				errorFields[err.Field] = true
			}

			for _, expectedField := range tc.expectedErrors {
				if !errorFields[expectedField] {
					t.Errorf("Expected error for field %s, but not found", expectedField)
				}
			}
		})
	}
}

// TestValidationPerformance tests validation performance
func TestValidationPerformance(t *testing.T) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := validation.NewValidator(log)
	config := validation.GetDefaultValidationConfig()

	// Run validation multiple times to test performance
	iterations := 1000

	for i := 0; i < iterations; i++ {
		result := validator.ValidateConfig(config)
		if !result.Valid {
			t.Fatalf("Valid config should always pass validation, failed at iteration %d", i)
		}
	}
}

// TestConfigMergeFlags tests flag merging functionality
func TestConfigMergeFlags(t *testing.T) {
	// Test merging various flag combinations
	testCases := []struct {
		name         string
		onlineOnly   bool
		noExclude    bool
		testMode     bool
		piholeConfig string
	}{
		{
			name:       "AllFlagsTrue",
			onlineOnly: true,
			noExclude:  true,
			testMode:   true,
		},
		{
			name:       "AllFlagsFalse",
			onlineOnly: false,
			noExclude:  false,
			testMode:   false,
		},
		{
			name:       "MixedFlags",
			onlineOnly: true,
			noExclude:  false,
			testMode:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testConfig := config.DefaultConfig()
			config.MergeFlags(testConfig, tc.onlineOnly, tc.noExclude, tc.testMode, tc.piholeConfig)

			if testConfig.OnlineOnly != tc.onlineOnly {
				t.Errorf("Expected OnlineOnly %v, got %v", tc.onlineOnly, testConfig.OnlineOnly)
			}
			if testConfig.NoExclude != tc.noExclude {
				t.Errorf("Expected NoExclude %v, got %v", tc.noExclude, testConfig.NoExclude)
			}
			if testConfig.TestMode != tc.testMode {
				t.Errorf("Expected TestMode %v, got %v", tc.testMode, testConfig.TestMode)
			}
		})
	}
}

// BenchmarkConfigValidation benchmarks the validation process
func BenchmarkConfigValidation(b *testing.B) {
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := validation.NewValidator(log)
	config := validation.GetDefaultValidationConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateConfig(config)
	}
}

// BenchmarkConfigLoadSave benchmarks config loading and saving
func BenchmarkConfigLoadSave(b *testing.B) {
	tmpDir := b.TempDir()
	configPath := filepath.Join(tmpDir, "bench_config.json")
	testConfig := validation.GetDefaultValidationConfig()

	// Initial save
	config.SaveConfig(testConfig, configPath)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loadedConfig, err := config.LoadConfig(configPath)
		if err != nil {
			b.Fatalf("Failed to load config: %v", err)
		}

		err = config.SaveConfig(loadedConfig, configPath)
		if err != nil {
			b.Fatalf("Failed to save config: %v", err)
		}
	}
}
