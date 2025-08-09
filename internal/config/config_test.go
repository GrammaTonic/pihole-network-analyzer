package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
	"pihole-analyzer/internal/validation"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config == nil {
		t.Error("DefaultConfig should not return nil")
	}

	// Validate the default config
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := validation.NewValidator(log)
	result := validator.ValidateConfig(config)

	if !result.Valid {
		t.Errorf("Default configuration should be valid, got %d errors", len(result.Errors))
		for _, err := range result.Errors {
			t.Logf("Error: %s", err.Error())
		}
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	config, err := LoadConfig("/tmp/non_existent_config.json")
	if err != nil {
		t.Errorf("LoadConfig should not return error for non-existent file: %v", err)
	}
	if config == nil {
		t.Error("LoadConfig should return default config for non-existent file")
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	validConfig := validation.GetDefaultValidationConfig()
	data, err := json.MarshalIndent(validConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	err = os.WriteFile(configPath, data, 0600)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Load the config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Errorf("LoadConfig should not return error for valid file: %v", err)
	}
	if config == nil {
		t.Error("LoadConfig should return config for valid file")
	}

	// Check some values
	if config.Pihole.Host != validConfig.Pihole.Host {
		t.Errorf("Expected host %s, got %s", validConfig.Pihole.Host, config.Pihole.Host)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	// Create a temporary config file with invalid JSON
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid_config.json")

	invalidJSON := `{"pihole": {"host": "192.168.1.1", "port": 80, invalid}`
	err := os.WriteFile(configPath, []byte(invalidJSON), 0600)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Load the config
	config, err := LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig should return error for invalid JSON")
	}
	if config != nil {
		t.Error("LoadConfig should return nil config for invalid JSON")
	}
}

func TestLoadConfig_InvalidConfig(t *testing.T) {
	// Create a temporary config file with invalid configuration
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid_config.json")

	invalidConfig := &types.Config{
		Pihole: types.PiholeConfig{
			Host: "", // Invalid empty host
			Port: -1, // Invalid port
		},
		Output: types.OutputConfig{
			MaxClients: 0, // Invalid max clients
		},
	}

	data, err := json.MarshalIndent(invalidConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	err = os.WriteFile(configPath, data, 0600)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Load the config - should succeed after applying defaults
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Errorf("LoadConfig should not return error after applying defaults: %v", err)
	}
	if config == nil {
		t.Error("LoadConfig should return config after applying defaults")
	}

	// Check that defaults were applied
	if config.Pihole.Host == "" {
		t.Error("Default host should have been applied")
	}
	if config.Pihole.Port <= 0 {
		t.Error("Default port should have been applied")
	}
	if config.Output.MaxClients <= 0 {
		t.Error("Default max clients should have been applied")
	}
}

func TestSaveConfig_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.json")

	config := validation.GetDefaultValidationConfig()
	err := SaveConfig(config, configPath)
	if err != nil {
		t.Errorf("SaveConfig should not return error for valid config: %v", err)
	}

	// Check that file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file should have been created")
	}

	// Load the saved config and verify it's the same
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Errorf("Failed to load saved config: %v", err)
	}

	if loadedConfig.Pihole.Host != config.Pihole.Host {
		t.Errorf("Expected host %s, got %s", config.Pihole.Host, loadedConfig.Pihole.Host)
	}
}

func TestSaveConfig_InvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid_config.json")

	invalidConfig := &types.Config{
		Pihole: types.PiholeConfig{
			Host: "", // Invalid empty host
		},
		Output: types.OutputConfig{
			MaxClients: 0, // Invalid max clients
		},
	}

	err := SaveConfig(invalidConfig, configPath)
	if err == nil {
		t.Error("SaveConfig should return error for invalid config")
	}

	// Check that file was not created
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("Config file should not have been created for invalid config")
	}
}

func TestCreateDefaultConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "default_config.json")

	err := CreateDefaultConfigFile(configPath)
	if err != nil {
		t.Errorf("CreateDefaultConfigFile should not return error: %v", err)
	}

	// Check that file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Default config file should have been created")
	}

	// Load and validate the created config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Errorf("Failed to load created default config: %v", err)
	}

	// Should be valid
	log := logger.New(&logger.Config{EnableColors: false, EnableEmojis: false})
	validator := validation.NewValidator(log)
	result := validator.ValidateConfig(config)

	if !result.Valid {
		t.Errorf("Created default config should be valid, got %d errors", len(result.Errors))
	}
}

func TestShowConfig(t *testing.T) {
	config := validation.GetDefaultValidationConfig()

	// This test mainly checks that ShowConfig doesn't panic
	// In a real test environment, you might want to capture the output
	// and verify the logged information
	ShowConfig(config)
}

func TestMergeFlags(t *testing.T) {
	config := DefaultConfig()

	// Test merging flags
	MergeFlags(config, true, true, true, "")

	if !config.OnlineOnly {
		t.Error("OnlineOnly flag should be set")
	}
	if !config.NoExclude {
		t.Error("NoExclude flag should be set")
	}
	if !config.TestMode {
		t.Error("TestMode flag should be set")
	}
}

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()
	if path == "" {
		t.Error("GetConfigPath should return a non-empty path")
	}

	// Should contain .pihole-analyzer
	if filepath.Base(filepath.Dir(path)) != ".pihole-analyzer" {
		t.Error("Config path should be in .pihole-analyzer directory")
	}

	// Should end with config.json
	if filepath.Base(path) != "config.json" {
		t.Error("Config path should end with config.json")
	}
}

// Integration test - full workflow
func TestFullConfigWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "workflow_config.json")

	// 1. Create default config
	err := CreateDefaultConfigFile(configPath)
	if err != nil {
		t.Fatalf("Failed to create default config: %v", err)
	}

	// 2. Load the config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 3. Modify the config
	config.Pihole.Host = "192.168.1.50"
	config.Output.MaxClients = 50

	// 4. Save the modified config
	err = SaveConfig(config, configPath)
	if err != nil {
		t.Fatalf("Failed to save modified config: %v", err)
	}

	// 5. Load the config again and verify changes
	config2, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load modified config: %v", err)
	}

	if config2.Pihole.Host != "192.168.1.50" {
		t.Error("Host modification was not persisted")
	}
	if config2.Output.MaxClients != 50 {
		t.Error("MaxClients modification was not persisted")
	}

	// 6. Display the config (just to test it doesn't panic)
	ShowConfig(config2)
}

// Benchmark tests
func BenchmarkLoadConfig(b *testing.B) {
	tmpDir := b.TempDir()
	configPath := filepath.Join(tmpDir, "bench_config.json")

	// Create a test config file
	config := validation.GetDefaultValidationConfig()
	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(configPath, data, 0600)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LoadConfig(configPath)
	}
}

func BenchmarkSaveConfig(b *testing.B) {
	tmpDir := b.TempDir()
	config := validation.GetDefaultValidationConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		configPath := filepath.Join(tmpDir, fmt.Sprintf("bench_config_%d.json", i))
		SaveConfig(config, configPath)
	}
}
