package testutils

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"pihole-analyzer/internal/types"

	_ "modernc.org/sqlite"
)

// getStatusCodeFromString converts status strings to status codes
func getStatusCodeFromString(status string) int {
	switch status {
	case "Blocked (gravity)":
		return 1
	case "Forwarded":
		return 2
	case "Cached":
		return 3
	case "Blocked (regex/wildcard)":
		return 4
	case "Blocked (exact)":
		return 5
	case "Blocked (external)":
		return 6
	case "CNAME":
		return 7
	case "Retried":
		return 8
	case "Retried but ignored":
		return 9
	case "Already forwarded":
		return 10
	case "Already cached":
		return 11
	case "Config blocked":
		return 12
	case "Gravity blocked":
		return 13
	case "Regex blocked":
		return 14
	default:
		return 0
	}
}

// MockData contains all the mock data for testing
// MockData contains all the mock data for testing (CSV support removed)
type MockData struct {
	PiholeRecords []types.PiholeRecord
	ARPEntries    map[string]*types.ARPEntry
	Hostnames     map[string]string
}

// CreateMockData generates realistic test data (Pi-hole only, CSV support removed)
func CreateMockData() *MockData {
	mock := &MockData{
		ARPEntries: make(map[string]*types.ARPEntry),
		Hostnames:  make(map[string]string),
	}

	// Mock Pi-hole records (simulating database data)
	mock.PiholeRecords = []types.PiholeRecord{
		{Timestamp: strconv.FormatInt(time.Now().Unix()-3600, 10), Client: "192.168.2.110", HWAddr: "66:78:8b:59:bf:1a", Domain: "google.com", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix()-3500, 10), Client: "192.168.2.210", HWAddr: "32:79:b9:39:43:7c", Domain: "facebook.com", Status: 9},
		{Timestamp: strconv.FormatInt(time.Now().Unix()-3400, 10), Client: "192.168.2.110", HWAddr: "66:78:8b:59:bf:1a", Domain: "youtube.com", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix()-3300, 10), Client: "192.168.2.210", HWAddr: "32:79:b9:39:43:7c", Domain: "tracking.doubleclick.net", Status: 9},
		{Timestamp: strconv.FormatInt(time.Now().Unix()-3200, 10), Client: "192.168.2.6", HWAddr: "e0:69:95:4f:19:6d", Domain: "api.spotify.com", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix()-3100, 10), Client: "172.20.0.8", HWAddr: "02:42:ac:14:00:08", Domain: "docker.internal", Status: 3},
		{Timestamp: strconv.FormatInt(time.Now().Unix()-3000, 10), Client: "192.168.2.110", HWAddr: "66:78:8b:59:bf:1a", Domain: "github.com", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix()-2900, 10), Client: "192.168.2.210", HWAddr: "32:79:b9:39:43:7c", Domain: "ads.microsoft.com", Status: 9},
		{Timestamp: strconv.FormatInt(time.Now().Unix()-2800, 10), Client: "192.168.2.202", HWAddr: "14:da:e9:48:53:a4", Domain: "cdn.jsdelivr.net", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix()-2700, 10), Client: "192.168.2.119", HWAddr: "fc:49:2d:1a:41:16", Domain: "malware.badsite.com", Status: 9},
		{Timestamp: strconv.FormatInt(time.Now().Unix()-2600, 10), Client: "192.168.2.202", HWAddr: "14:da:e9:48:53:a4", Domain: "netflix.com", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix()-2500, 10), Client: "192.168.2.119", HWAddr: "fc:49:2d:1a:41:16", Domain: "amazon.com", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix()-2400, 10), Client: "192.168.2.110", HWAddr: "66:78:8b:59:bf:1a", Domain: "microsoft.com", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix()-2300, 10), Client: "192.168.2.210", HWAddr: "32:79:b9:39:43:7c", Domain: "telemetry.microsoft.com", Status: 9},
		{Timestamp: strconv.FormatInt(time.Now().Unix()-2200, 10), Client: "192.168.2.6", HWAddr: "e0:69:95:4f:19:6d", Domain: "pi.hole", Status: 3},
		// More Docker container traffic
		{Timestamp: strconv.FormatInt(time.Now().Unix()-2100, 10), Client: "172.20.0.8", HWAddr: "02:42:ac:14:00:08", Domain: "registry-1.docker.io", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix()-2000, 10), Client: "172.19.0.2", HWAddr: "02:42:ac:13:00:02", Domain: "hub.docker.com", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix()-1900, 10), Client: "172.20.0.8", HWAddr: "02:42:ac:14:00:08", Domain: "auth.docker.io", Status: 2},
		// Offline devices
		{Timestamp: strconv.FormatInt(time.Now().Unix()-1800, 10), Client: "192.168.2.123", HWAddr: "fc:41:16:b4:22:33", Domain: "old.device.com", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix()-1700, 10), Client: "192.168.2.156", HWAddr: "f2:6f:b8:f3:44:55", Domain: "another.old.com", Status: 3},
	}

	// Mock ARP entries (simulating current online devices)
	mock.ARPEntries = map[string]*types.ARPEntry{
		"192.168.2.110": {IP: "192.168.2.110", MAC: "66:78:8B:59:BF:1A", IsOnline: true},
		"192.168.2.210": {IP: "192.168.2.210", MAC: "32:79:B9:39:43:7C", IsOnline: true},
		"192.168.2.6":   {IP: "192.168.2.6", MAC: "E0:69:95:4F:19:6D", IsOnline: true},
		"192.168.2.202": {IP: "192.168.2.202", MAC: "14:DA:E9:48:53:A4", IsOnline: true},
		"192.168.2.119": {IP: "192.168.2.119", MAC: "FC:49:2D:1A:41:16", IsOnline: true},
		// Note: 192.168.2.123 and 192.168.2.156 are NOT in ARP table (offline)
	}

	// Mock hostname resolutions
	mock.Hostnames = map[string]string{
		"192.168.2.110": "mac.home",
		"192.168.2.210": "s21-van-marloes.home",
		"192.168.2.6":   "pi.hole",
		"192.168.2.202": "maximus.home",
		"192.168.2.119": "alexa.home",
		"192.168.2.123": "grammatonic.home",
		"192.168.2.156": "syam-s-a33.home",
	}

	return mock
}

// CreateMockPiholeDatabase creates a SQLite database with mock Pi-hole data
func CreateMockPiholeDatabase(filename string, mockData *MockData) error {
	// Remove file if it exists
	os.Remove(filename)

	db, err := sql.Open("sqlite", filename)
	if err != nil {
		return fmt.Errorf("error creating mock database: %v", err)
	}
	defer db.Close()

	// Create tables (simplified Pi-hole schema)
	createTablesSQL := `
	CREATE TABLE queries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp INTEGER NOT NULL,
		type INTEGER NOT NULL,
		status INTEGER NOT NULL,
		domain TEXT NOT NULL,
		client TEXT NOT NULL,
		forward TEXT
	);

	CREATE TABLE network (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hwaddr TEXT,
		interface TEXT,
		name TEXT,
		firstSeen INTEGER,
		lastQuery INTEGER,
		numQueries INTEGER,
		macVendor TEXT
	);

	CREATE TABLE network_addresses (
		network_id INTEGER,
		ip TEXT,
		lastSeen INTEGER,
		FOREIGN KEY(network_id) REFERENCES network(id)
	);
	`

	_, err = db.Exec(createTablesSQL)
	if err != nil {
		return fmt.Errorf("error creating tables: %v", err)
	}

	// Insert mock data
	// First, insert network entries
	networkInsert := `INSERT INTO network (id, hwaddr, interface, name, firstSeen, lastQuery, numQueries, macVendor) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	addressInsert := `INSERT INTO network_addresses (network_id, ip, lastSeen) VALUES (?, ?, ?)`

	networks := map[string]struct {
		id     int
		hwaddr string
		name   string
		ip     string
		vendor string
	}{
		"192.168.2.110": {1, "66:78:8b:59:bf:1a", "mac.home", "192.168.2.110", "Apple"},
		"192.168.2.210": {2, "32:79:b9:39:43:7c", "s21-van-marloes.home", "192.168.2.210", "Samsung"},
		"192.168.2.6":   {3, "e0:69:95:4f:19:6d", "pi.hole", "192.168.2.6", "Raspberry Pi"},
		"192.168.2.202": {4, "14:da:e9:48:53:a4", "maximus.home", "192.168.2.202", "Dell"},
		"192.168.2.119": {5, "fc:49:2d:1a:41:16", "alexa.home", "192.168.2.119", "Amazon"},
		"192.168.2.123": {6, "fc:41:16:b4:22:33", "grammatonic.home", "192.168.2.123", "Unknown"},
		"192.168.2.156": {7, "f2:6f:b8:f3:44:55", "syam-s-a33.home", "192.168.2.156", "Unknown"},
		"172.20.0.8":    {8, "02:42:ac:14:00:08", "", "172.20.0.8", "Docker"},
		"172.19.0.2":    {9, "02:42:ac:13:00:02", "", "172.19.0.2", "Docker"},
	}

	for _, network := range networks {
		_, err = db.Exec(networkInsert, network.id, network.hwaddr, "eth0", network.name,
			time.Now().Unix()-86400, time.Now().Unix(), 100, network.vendor)
		if err != nil {
			return fmt.Errorf("error inserting network: %v", err)
		}

		_, err = db.Exec(addressInsert, network.id, network.ip, time.Now().Unix())
		if err != nil {
			return fmt.Errorf("error inserting address: %v", err)
		}
	}

	// Insert query records
	queryInsert := `INSERT INTO queries (timestamp, type, status, domain, client, forward) VALUES (?, ?, ?, ?, ?, ?)`

	for _, record := range mockData.PiholeRecords {
		// record.Status is already an int in types.PiholeRecord
		timestamp := time.Now().Unix() // Use current time if record.Timestamp is not available
		_, err = db.Exec(queryInsert, timestamp, 1, record.Status, record.Domain, record.Client, "")
		if err != nil {
			return fmt.Errorf("error inserting query: %v", err)
		}
	}

	return nil
}

// CreateMockConfig creates a mock Pi-hole configuration file for testing
func CreateMockConfig(filename string) error {
	config := `{
  "host": "localhost",
  "port": "22",
  "username": "testuser",
  "password": "",
  "keyfile": "",
  "dbpath": "/tmp/mock-pihole.db"
}`

	return os.WriteFile(filename, []byte(config), 0600)
}

// MockARPTable simulates the ARP table for testing
func MockARPTable(mockData *MockData) (map[string]*types.ARPEntry, error) {
	return mockData.ARPEntries, nil
}

// MockHostnameResolution simulates hostname resolution for testing
func MockHostnameResolution(ip string, mockData *MockData) string {
	if hostname, exists := mockData.Hostnames[ip]; exists {
		return hostname
	}
	return ""
}

// SetupTestEnvironment creates all mock files and data for testing
func SetupTestEnvironment() (*MockData, error) {
	fmt.Println("Setting up test environment with mock data...")

	mockData := CreateMockData()

	// Create test directory
	testDir := "testing/fixtures"
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("error creating test directory: %v", err)
	}

	// Create mock Pi-hole database
	dbFile := filepath.Join(testDir, "mock_pihole.db")
	err = CreateMockPiholeDatabase(dbFile, mockData)
	if err != nil {
		return nil, fmt.Errorf("error creating mock database: %v", err)
	}

	// Create mock config
	configFile := filepath.Join(testDir, "mock_pihole_config.json")
	err = CreateMockConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("error creating mock config: %v", err)
	}

	fmt.Printf("✓ Mock Pi-hole database created: %s\n", dbFile)
	fmt.Printf("✓ Mock configuration created: %s\n", configFile)
	fmt.Println("✓ Test environment setup complete (CSV support removed)!")

	return mockData, nil
}

// CleanupTestEnvironment removes all test files
func CleanupTestEnvironment() {
	os.RemoveAll("testing/fixtures")
	fmt.Println("✓ Test environment cleaned up")
}
