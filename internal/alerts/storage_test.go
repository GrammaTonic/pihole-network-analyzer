package alerts

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"pihole-analyzer/internal/logger"
)

// TestMemoryStorage tests in-memory alert storage
func TestMemoryStorage(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())
	config := StorageConfig{
		Type:      "memory",
		MaxSize:   100,
		Retention: 24 * time.Hour,
	}

	storage := NewMemoryStorage(config, logger)

	// Test storing an alert
	alert := &Alert{
		ID:          "test-alert-1",
		Type:        AlertTypeAnomaly,
		Severity:    SeverityWarning,
		Status:      AlertStatusFired,
		Title:       "Test Alert",
		Description: "This is a test alert",
		Timestamp:   time.Now(),
		Source:      "test",
		Tags:        []string{"test"},
	}

	if err := storage.Store(alert); err != nil {
		t.Errorf("failed to store alert: %v", err)
	}

	// Test retrieving the alert
	retrieved, err := storage.Get(alert.ID)
	if err != nil {
		t.Errorf("failed to get alert: %v", err)
	}
	if retrieved.ID != alert.ID {
		t.Errorf("expected ID %s, got %s", alert.ID, retrieved.ID)
	}

	// Test listing alerts
	alerts, err := storage.List(map[string]interface{}{})
	if err != nil {
		t.Errorf("failed to list alerts: %v", err)
	}
	if len(alerts) != 1 {
		t.Errorf("expected 1 alert, got %d", len(alerts))
	}

	// Test updating an alert
	alert.Status = AlertStatusResolved
	now := time.Now()
	alert.ResolvedAt = &now

	if err := storage.Update(alert); err != nil {
		t.Errorf("failed to update alert: %v", err)
	}

	updated, err := storage.Get(alert.ID)
	if err != nil {
		t.Errorf("failed to get updated alert: %v", err)
	}
	if updated.Status != AlertStatusResolved {
		t.Errorf("expected status %s, got %s", AlertStatusResolved, updated.Status)
	}

	// Test getting active alerts
	activeAlerts, err := storage.GetActive()
	if err != nil {
		t.Errorf("failed to get active alerts: %v", err)
	}
	if len(activeAlerts) != 0 { // Should be 0 since we resolved the alert
		t.Errorf("expected 0 active alerts, got %d", len(activeAlerts))
	}

	// Test deleting an alert
	if err := storage.Delete(alert.ID); err != nil {
		t.Errorf("failed to delete alert: %v", err)
	}

	// Verify deletion
	_, err = storage.Get(alert.ID)
	if err == nil {
		t.Error("expected error when getting deleted alert")
	}

	// Test error cases
	if err := storage.Update(&Alert{ID: "non-existent"}); err == nil {
		t.Error("expected error when updating non-existent alert")
	}

	if err := storage.Delete("non-existent"); err == nil {
		t.Error("expected error when deleting non-existent alert")
	}
}

// TestMemoryStorageFiltering tests filtering functionality
func TestMemoryStorageFiltering(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())
	config := StorageConfig{
		Type:      "memory",
		MaxSize:   100,
		Retention: 24 * time.Hour,
	}

	storage := NewMemoryStorage(config, logger)

	// Create test alerts with different properties
	alerts := []*Alert{
		{
			ID:        "alert-1",
			Type:      AlertTypeAnomaly,
			Severity:  SeverityWarning,
			Status:    AlertStatusFired,
			Title:     "Anomaly Alert",
			Timestamp: time.Now(),
			Source:    "ml-engine",
		},
		{
			ID:        "alert-2",
			Type:      AlertTypeThreshold,
			Severity:  SeverityCritical,
			Status:    AlertStatusFired,
			Title:     "Threshold Alert",
			Timestamp: time.Now().Add(-1 * time.Hour),
			Source:    "rule-engine",
		},
		{
			ID:        "alert-3",
			Type:      AlertTypeAnomaly,
			Severity:  SeverityInfo,
			Status:    AlertStatusResolved,
			Title:     "Resolved Alert",
			Timestamp: time.Now().Add(-2 * time.Hour),
			Source:    "ml-engine",
		},
	}

	// Store all alerts
	for _, alert := range alerts {
		if err := storage.Store(alert); err != nil {
			t.Errorf("failed to store alert %s: %v", alert.ID, err)
		}
	}

	// Test filtering by type
	typeFiltered, err := storage.List(map[string]interface{}{
		"type": AlertTypeAnomaly,
	})
	if err != nil {
		t.Errorf("failed to filter by type: %v", err)
	}
	if len(typeFiltered) != 2 {
		t.Errorf("expected 2 anomaly alerts, got %d", len(typeFiltered))
	}

	// Test filtering by severity
	severityFiltered, err := storage.List(map[string]interface{}{
		"severity": SeverityCritical,
	})
	if err != nil {
		t.Errorf("failed to filter by severity: %v", err)
	}
	if len(severityFiltered) != 1 {
		t.Errorf("expected 1 critical alert, got %d", len(severityFiltered))
	}

	// Test filtering by status
	statusFiltered, err := storage.List(map[string]interface{}{
		"status": []AlertStatus{AlertStatusFired},
	})
	if err != nil {
		t.Errorf("failed to filter by status: %v", err)
	}
	if len(statusFiltered) != 2 {
		t.Errorf("expected 2 fired alerts, got %d", len(statusFiltered))
	}

	// Test filtering by source
	sourceFiltered, err := storage.List(map[string]interface{}{
		"source": "ml-engine",
	})
	if err != nil {
		t.Errorf("failed to filter by source: %v", err)
	}
	if len(sourceFiltered) != 2 {
		t.Errorf("expected 2 ml-engine alerts, got %d", len(sourceFiltered))
	}

	// Test limit
	limited, err := storage.List(map[string]interface{}{
		"limit": 2,
	})
	if err != nil {
		t.Errorf("failed to apply limit: %v", err)
	}
	if len(limited) != 2 {
		t.Errorf("expected 2 limited alerts, got %d", len(limited))
	}

	// Verify sorting (should be newest first)
	all, err := storage.List(map[string]interface{}{})
	if err != nil {
		t.Errorf("failed to list all alerts: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 total alerts, got %d", len(all))
	}
	// First alert should be the newest (alert-1)
	if all[0].ID != "alert-1" {
		t.Errorf("expected newest alert first, got %s", all[0].ID)
	}
}

// TestMemoryStorageCleanup tests cleanup functionality
func TestMemoryStorageCleanup(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())
	config := StorageConfig{
		Type:      "memory",
		MaxSize:   2, // Small size for testing
		Retention: 1 * time.Hour,
	}

	storage := NewMemoryStorage(config, logger)

	// Create old alert (should be cleaned up)
	oldAlert := &Alert{
		ID:        "old-alert",
		Type:      AlertTypeAnomaly,
		Severity:  SeverityWarning,
		Status:    AlertStatusFired,
		Title:     "Old Alert",
		Timestamp: time.Now().Add(-2 * time.Hour), // Older than retention
		Source:    "test",
	}

	// Create new alerts
	newAlert1 := &Alert{
		ID:        "new-alert-1",
		Type:      AlertTypeAnomaly,
		Severity:  SeverityWarning,
		Status:    AlertStatusFired,
		Title:     "New Alert 1",
		Timestamp: time.Now(),
		Source:    "test",
	}

	newAlert2 := &Alert{
		ID:        "new-alert-2",
		Type:      AlertTypeAnomaly,
		Severity:  SeverityWarning,
		Status:    AlertStatusFired,
		Title:     "New Alert 2",
		Timestamp: time.Now(),
		Source:    "test",
	}

	// Store old alert first
	if err := storage.Store(oldAlert); err != nil {
		t.Errorf("failed to store old alert: %v", err)
	}

	// Manual cleanup to test retention
	if err := storage.Cleanup(); err != nil {
		t.Errorf("failed to cleanup: %v", err)
	}

	// Old alert should be removed
	alerts, err := storage.List(map[string]interface{}{})
	if err != nil {
		t.Errorf("failed to list alerts after cleanup: %v", err)
	}
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts after cleanup, got %d", len(alerts))
	}

	// Store alerts to test size limit
	if err := storage.Store(newAlert1); err != nil {
		t.Errorf("failed to store new alert 1: %v", err)
	}
	if err := storage.Store(newAlert2); err != nil {
		t.Errorf("failed to store new alert 2: %v", err)
	}

	// Store third alert - should trigger size-based cleanup
	newAlert3 := &Alert{
		ID:        "new-alert-3",
		Type:      AlertTypeAnomaly,
		Severity:  SeverityWarning,
		Status:    AlertStatusFired,
		Title:     "New Alert 3",
		Timestamp: time.Now(),
		Source:    "test",
	}

	if err := storage.Store(newAlert3); err != nil {
		t.Errorf("failed to store new alert 3: %v", err)
	}

	// Should still have max 2 alerts
	alerts, err = storage.List(map[string]interface{}{})
	if err != nil {
		t.Errorf("failed to list alerts after size limit: %v", err)
	}
	if len(alerts) > 2 {
		t.Errorf("expected max 2 alerts after size limit, got %d", len(alerts))
	}
}

// TestFileStorage tests file-based alert storage
func TestFileStorage(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())

	// Create temporary file for testing
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_alerts.json")

	config := StorageConfig{
		Type:      "file",
		Path:      tempFile,
		MaxSize:   100,
		Retention: 24 * time.Hour,
	}

	storage, err := NewFileStorage(config, logger)
	if err != nil {
		t.Fatalf("failed to create file storage: %v", err)
	}

	// Test storing an alert
	alert := &Alert{
		ID:          "test-file-alert-1",
		Type:        AlertTypeAnomaly,
		Severity:    SeverityWarning,
		Status:      AlertStatusFired,
		Title:       "Test File Alert",
		Description: "This is a test alert for file storage",
		Timestamp:   time.Now(),
		Source:      "test",
		Tags:        []string{"test", "file"},
	}

	if err := storage.Store(alert); err != nil {
		t.Errorf("failed to store alert: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Error("alert file was not created")
	}

	// Test retrieving the alert
	retrieved, err := storage.Get(alert.ID)
	if err != nil {
		t.Errorf("failed to get alert: %v", err)
	}
	if retrieved.ID != alert.ID {
		t.Errorf("expected ID %s, got %s", alert.ID, retrieved.ID)
	}

	// Test listing alerts
	alerts, err := storage.List(map[string]interface{}{})
	if err != nil {
		t.Errorf("failed to list alerts: %v", err)
	}
	if len(alerts) != 1 {
		t.Errorf("expected 1 alert, got %d", len(alerts))
	}

	// Test updating an alert
	alert.Status = AlertStatusResolved
	now := time.Now()
	alert.ResolvedAt = &now

	if err := storage.Update(alert); err != nil {
		t.Errorf("failed to update alert: %v", err)
	}

	updated, err := storage.Get(alert.ID)
	if err != nil {
		t.Errorf("failed to get updated alert: %v", err)
	}
	if updated.Status != AlertStatusResolved {
		t.Errorf("expected status %s, got %s", AlertStatusResolved, updated.Status)
	}

	// Test deleting an alert
	if err := storage.Delete(alert.ID); err != nil {
		t.Errorf("failed to delete alert: %v", err)
	}

	// Verify deletion
	_, err = storage.Get(alert.ID)
	if err == nil {
		t.Error("expected error when getting deleted alert")
	}

	// Close storage
	if err := storage.Close(); err != nil {
		t.Errorf("failed to close storage: %v", err)
	}
}

// TestFileStoragePersistence tests that file storage persists across instances
func TestFileStoragePersistence(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())

	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "persistence_test.json")

	config := StorageConfig{
		Type:      "file",
		Path:      tempFile,
		MaxSize:   100,
		Retention: 24 * time.Hour,
	}

	// Create first storage instance and store an alert
	storage1, err := NewFileStorage(config, logger)
	if err != nil {
		t.Fatalf("failed to create first storage instance: %v", err)
	}

	alert := &Alert{
		ID:        "persistence-test",
		Type:      AlertTypeAnomaly,
		Severity:  SeverityWarning,
		Status:    AlertStatusFired,
		Title:     "Persistence Test",
		Timestamp: time.Now(),
		Source:    "test",
	}

	if err := storage1.Store(alert); err != nil {
		t.Errorf("failed to store alert in first instance: %v", err)
	}

	storage1.Close()

	// Create second storage instance and verify alert is still there
	storage2, err := NewFileStorage(config, logger)
	if err != nil {
		t.Fatalf("failed to create second storage instance: %v", err)
	}

	retrieved, err := storage2.Get(alert.ID)
	if err != nil {
		t.Errorf("failed to retrieve alert from second instance: %v", err)
	}
	if retrieved.ID != alert.ID {
		t.Errorf("expected ID %s, got %s", alert.ID, retrieved.ID)
	}

	alerts, err := storage2.List(map[string]interface{}{})
	if err != nil {
		t.Errorf("failed to list alerts from second instance: %v", err)
	}
	if len(alerts) != 1 {
		t.Errorf("expected 1 alert in second instance, got %d", len(alerts))
	}

	storage2.Close()
}

// TestFileStorageErrorHandling tests file storage error handling
func TestFileStorageErrorHandling(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())

	// Test invalid file path
	invalidConfig := StorageConfig{
		Type: "file",
		Path: "", // Empty path should cause error
	}

	_, err := NewFileStorage(invalidConfig, logger)
	if err == nil {
		t.Error("expected error for empty file path")
	}

	// Test unwritable directory (if running as non-root)
	unwritableConfig := StorageConfig{
		Type: "file",
		Path: "/root/unwritable/alerts.json",
	}

	storage, err := NewFileStorage(unwritableConfig, logger)
	if err == nil && storage != nil {
		// Only test this if we couldn't create the storage
		// (this test might not work in all environments)
		alert := &Alert{
			ID:        "unwritable-test",
			Type:      AlertTypeAnomaly,
			Severity:  SeverityWarning,
			Status:    AlertStatusFired,
			Title:     "Unwritable Test",
			Timestamp: time.Now(),
			Source:    "test",
		}

		// This should fail in most cases, but might succeed in some test environments
		storage.Store(alert)
		storage.Close()
	}
}
