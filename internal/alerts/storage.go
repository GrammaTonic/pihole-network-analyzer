package alerts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"pihole-analyzer/internal/logger"
)

// MemoryStorage implements in-memory alert storage
type MemoryStorage struct {
	config    StorageConfig
	logger    *logger.Logger
	alerts    map[string]*Alert
	mutex     sync.RWMutex
	maxSize   int
	retention time.Duration
}

// NewMemoryStorage creates a new memory-based storage
func NewMemoryStorage(config StorageConfig, logger *logger.Logger) *MemoryStorage {
	return &MemoryStorage{
		config:    config,
		logger:    logger.Component("memory-storage"),
		alerts:    make(map[string]*Alert),
		maxSize:   config.MaxSize,
		retention: config.Retention,
	}
}

// Store stores an alert in memory
func (s *MemoryStorage) Store(alert *Alert) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if we need to cleanup old alerts first
	if len(s.alerts) >= s.maxSize {
		s.cleanup()
	}

	s.alerts[alert.ID] = alert
	s.logger.GetSlogger().Debug("Alert stored in memory", "alert_id", alert.ID)
	return nil
}

// Get retrieves an alert by ID
func (s *MemoryStorage) Get(id string) (*Alert, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	alert, exists := s.alerts[id]
	if !exists {
		return nil, fmt.Errorf("alert %s not found", id)
	}

	return alert, nil
}

// Update updates an existing alert
func (s *MemoryStorage) Update(alert *Alert) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.alerts[alert.ID]; !exists {
		return fmt.Errorf("alert %s not found", alert.ID)
	}

	s.alerts[alert.ID] = alert
	s.logger.GetSlogger().Debug("Alert updated in memory", "alert_id", alert.ID)
	return nil
}

// Delete removes an alert
func (s *MemoryStorage) Delete(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.alerts[id]; !exists {
		return fmt.Errorf("alert %s not found", id)
	}

	delete(s.alerts, id)
	s.logger.GetSlogger().Debug("Alert deleted from memory", "alert_id", id)
	return nil
}

// List returns alerts matching the given filters
func (s *MemoryStorage) List(filters map[string]interface{}) ([]*Alert, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var alerts []*Alert
	for _, alert := range s.alerts {
		if s.matchesFilters(alert, filters) {
			alerts = append(alerts, alert)
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].Timestamp.After(alerts[j].Timestamp)
	})

	// Apply limit if specified
	if limit, ok := filters["limit"].(int); ok && limit > 0 && len(alerts) > limit {
		alerts = alerts[:limit]
	}

	return alerts, nil
}

// GetActive returns all active alerts
func (s *MemoryStorage) GetActive() ([]*Alert, error) {
	filters := map[string]interface{}{
		"status": []AlertStatus{AlertStatusFired, AlertStatusPending},
	}
	return s.List(filters)
}

// Cleanup removes old alerts based on retention policy
func (s *MemoryStorage) Cleanup() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.cleanup()
	return nil
}

// cleanup performs the actual cleanup (must be called with lock held)
func (s *MemoryStorage) cleanup() {
	// First, cleanup based on retention time
	if s.retention > 0 {
		cutoff := time.Now().Add(-s.retention)
		var toDelete []string

		for id, alert := range s.alerts {
			if alert.Timestamp.Before(cutoff) {
				toDelete = append(toDelete, id)
			}
		}

		for _, id := range toDelete {
			delete(s.alerts, id)
		}

		if len(toDelete) > 0 {
			s.logger.GetSlogger().Debug("Cleaned up old alerts", "count", len(toDelete))
		}
	}

	// Then, cleanup based on size limit
	if s.maxSize > 0 && len(s.alerts) >= s.maxSize {
		// Convert to slice for sorting
		type alertEntry struct {
			id    string
			alert *Alert
		}

		var entries []alertEntry
		for id, alert := range s.alerts {
			entries = append(entries, alertEntry{id: id, alert: alert})
		}

		// Sort by timestamp (oldest first)
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].alert.Timestamp.Before(entries[j].alert.Timestamp)
		})

		// Remove oldest alerts to get under the limit
		numToRemove := len(entries) - s.maxSize + 1 // +1 to make room for new alert
		for i := 0; i < numToRemove && i < len(entries); i++ {
			delete(s.alerts, entries[i].id)
		}

		if numToRemove > 0 {
			s.logger.GetSlogger().Debug("Cleaned up alerts due to size limit", "removed", numToRemove)
		}
	}
}

// Close closes the storage
func (s *MemoryStorage) Close() error {
	s.logger.GetSlogger().Debug("Memory storage closed")
	return nil
}

// matchesFilters checks if an alert matches the given filters
func (s *MemoryStorage) matchesFilters(alert *Alert, filters map[string]interface{}) bool {
	for key, value := range filters {
		switch key {
		case "status":
			if statuses, ok := value.([]AlertStatus); ok {
				found := false
				for _, status := range statuses {
					if alert.Status == status {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			} else if status, ok := value.(AlertStatus); ok {
				if alert.Status != status {
					return false
				}
			}
		case "type":
			if alertType, ok := value.(AlertType); ok && alert.Type != alertType {
				return false
			}
		case "severity":
			if severity, ok := value.(AlertSeverity); ok && alert.Severity != severity {
				return false
			}
		case "source":
			if source, ok := value.(string); ok && alert.Source != source {
				return false
			}
		case "limit":
			// Skip limit filter in matching - handled separately
			continue
		}
	}
	return true
}

// FileStorage implements file-based alert storage
type FileStorage struct {
	config   StorageConfig
	logger   *logger.Logger
	filePath string
	mutex    sync.RWMutex
}

// NewFileStorage creates a new file-based storage
func NewFileStorage(config StorageConfig, logger *logger.Logger) (*FileStorage, error) {
	if config.Path == "" {
		return nil, fmt.Errorf("file path is required for file storage")
	}

	// Ensure directory exists
	dir := filepath.Dir(config.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	return &FileStorage{
		config:   config,
		logger:   logger.Component("file-storage"),
		filePath: config.Path,
	}, nil
}

// Store stores an alert to file
func (s *FileStorage) Store(alert *Alert) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	alerts, err := s.loadAlerts()
	if err != nil {
		return fmt.Errorf("failed to load existing alerts: %w", err)
	}

	alerts = append(alerts, alert)

	// Apply size limit
	if s.config.MaxSize > 0 && len(alerts) > s.config.MaxSize {
		// Keep the most recent alerts
		sort.Slice(alerts, func(i, j int) bool {
			return alerts[i].Timestamp.After(alerts[j].Timestamp)
		})
		alerts = alerts[:s.config.MaxSize]
	}

	return s.saveAlerts(alerts)
}

// Get retrieves an alert by ID from file
func (s *FileStorage) Get(id string) (*Alert, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	alerts, err := s.loadAlerts()
	if err != nil {
		return nil, err
	}

	for _, alert := range alerts {
		if alert.ID == id {
			return alert, nil
		}
	}

	return nil, fmt.Errorf("alert %s not found", id)
}

// Update updates an existing alert in file
func (s *FileStorage) Update(alert *Alert) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	alerts, err := s.loadAlerts()
	if err != nil {
		return err
	}

	found := false
	for i, existing := range alerts {
		if existing.ID == alert.ID {
			alerts[i] = alert
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("alert %s not found", alert.ID)
	}

	return s.saveAlerts(alerts)
}

// Delete removes an alert from file
func (s *FileStorage) Delete(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	alerts, err := s.loadAlerts()
	if err != nil {
		return err
	}

	var updatedAlerts []*Alert
	found := false
	for _, alert := range alerts {
		if alert.ID != id {
			updatedAlerts = append(updatedAlerts, alert)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("alert %s not found", id)
	}

	return s.saveAlerts(updatedAlerts)
}

// List returns alerts matching the given filters from file
func (s *FileStorage) List(filters map[string]interface{}) ([]*Alert, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	alerts, err := s.loadAlerts()
	if err != nil {
		return nil, err
	}

	var filtered []*Alert
	for _, alert := range alerts {
		if s.matchesFilters(alert, filters) {
			filtered = append(filtered, alert)
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	// Apply limit if specified
	if limit, ok := filters["limit"].(int); ok && limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered, nil
}

// GetActive returns all active alerts from file
func (s *FileStorage) GetActive() ([]*Alert, error) {
	filters := map[string]interface{}{
		"status": []AlertStatus{AlertStatusFired, AlertStatusPending},
	}
	return s.List(filters)
}

// Cleanup removes old alerts from file based on retention policy
func (s *FileStorage) Cleanup() error {
	if s.config.Retention <= 0 {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	alerts, err := s.loadAlerts()
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-s.config.Retention)
	var validAlerts []*Alert

	for _, alert := range alerts {
		if alert.Timestamp.After(cutoff) {
			validAlerts = append(validAlerts, alert)
		}
	}

	if len(validAlerts) != len(alerts) {
		s.logger.GetSlogger().Debug("Cleaned up old alerts from file", "removed", len(alerts)-len(validAlerts))
		return s.saveAlerts(validAlerts)
	}

	return nil
}

// Close closes the file storage
func (s *FileStorage) Close() error {
	s.logger.GetSlogger().Debug("File storage closed")
	return nil
}

// loadAlerts loads alerts from file
func (s *FileStorage) loadAlerts() ([]*Alert, error) {
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return []*Alert{}, nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read alerts file: %w", err)
	}

	if len(data) == 0 {
		return []*Alert{}, nil
	}

	var alerts []*Alert
	if err := json.Unmarshal(data, &alerts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal alerts: %w", err)
	}

	return alerts, nil
}

// saveAlerts saves alerts to file
func (s *FileStorage) saveAlerts(alerts []*Alert) error {
	data, err := json.MarshalIndent(alerts, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal alerts: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write alerts file: %w", err)
	}

	s.logger.GetSlogger().Debug("Alerts saved to file", "count", len(alerts), "path", s.filePath)
	return nil
}

// matchesFilters checks if an alert matches the given filters (same as memory storage)
func (s *FileStorage) matchesFilters(alert *Alert, filters map[string]interface{}) bool {
	for key, value := range filters {
		switch key {
		case "status":
			if statuses, ok := value.([]AlertStatus); ok {
				found := false
				for _, status := range statuses {
					if alert.Status == status {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			} else if status, ok := value.(AlertStatus); ok {
				if alert.Status != status {
					return false
				}
			}
		case "type":
			if alertType, ok := value.(AlertType); ok && alert.Type != alertType {
				return false
			}
		case "severity":
			if severity, ok := value.(AlertSeverity); ok && alert.Severity != severity {
				return false
			}
		case "source":
			if source, ok := value.(string); ok && alert.Source != source {
				return false
			}
		case "limit":
			// Skip limit filter in matching - handled separately
			continue
		}
	}
	return true
}
