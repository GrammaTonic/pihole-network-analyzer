package testutils

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	_ "modernc.org/sqlite"
	"pihole-network-analyzer/internal/types"
)

// MockData contains all the mock data for testing
type MockData struct {
	DNSRecords    []types.DNSRecord
	PiholeRecords []types.PiholeRecord
	ARPEntries    map[string]*types.ARPEntry
	Hostnames     map[string]string
}

// CreateMockData generates realistic test data
func CreateMockData() *MockData {
	mock := &MockData{
		ARPEntries: make(map[string]*types.ARPEntry),
		Hostnames:  make(map[string]string),
	}

	// Mock DNS records (simulating CSV data)
	mock.DNSRecords = []types.DNSRecord{
		{ID: 1, DateTime: "2024-08-06 10:00:01", Domain: "google.com", Type: 1, Client: "192.168.2.110", ResponseTime: 0.003245},
		{ID: 2, DateTime: "2024-08-06 10:00:02", Domain: "facebook.com", Type: 1, Client: "192.168.2.210", ResponseTime: 0.002134},
		{ID: 3, DateTime: "2024-08-06 10:00:03", Domain: "youtube.com", Type: 1, Client: "192.168.2.110", ResponseTime: 0.004567},
		{ID: 4, DateTime: "2024-08-06 10:00:04", Domain: "tracking.doubleclick.net", Type: 1, Client: "192.168.2.210", ResponseTime: 0.001234},
		{ID: 5, DateTime: "2024-08-06 10:00:05", Domain: "api.spotify.com", Type: 1, Client: "192.168.2.6", ResponseTime: 0.002876},
		{ID: 6, DateTime: "2024-08-06 10:00:06", Domain: "docker.internal", Type: 1, Client: "172.20.0.8", ResponseTime: 0.000234},
		{ID: 7, DateTime: "2024-08-06 10:00:07", Domain: "github.com", Type: 1, Client: "192.168.2.110", ResponseTime: 0.003456},
		{ID: 8, DateTime: "2024-08-06 10:00:08", Domain: "ads.microsoft.com", Type: 1, Client: "192.168.2.210", ResponseTime: 0.001567},
		{ID: 9, DateTime: "2024-08-06 10:00:09", Domain: "cdn.jsdelivr.net", Type: 1, Client: "192.168.2.202", ResponseTime: 0.004123},
		{ID: 10, DateTime: "2024-08-06 10:00:10", Domain: "malware.badsite.com", Type: 1, Client: "192.168.2.119", ResponseTime: 0.000456},
		// Add more records for variety
		{ID: 11, DateTime: "2024-08-06 10:00:11", Domain: "netflix.com", Type: 1, Client: "192.168.2.202", ResponseTime: 0.005234},
		{ID: 12, DateTime: "2024-08-06 10:00:12", Domain: "amazon.com", Type: 1, Client: "192.168.2.119", ResponseTime: 0.003789},
		{ID: 13, DateTime: "2024-08-06 10:00:13", Domain: "microsoft.com", Type: 1, Client: "192.168.2.110", ResponseTime: 0.002345},
		{ID: 14, DateTime: "2024-08-06 10:00:14", Domain: "telemetry.microsoft.com", Type: 1, Client: "192.168.2.210", ResponseTime: 0.001234},
		{ID: 15, DateTime: "2024-08-06 10:00:15", Domain: "pi.hole", Type: 1, Client: "192.168.2.6", ResponseTime: 0.000123},
		// Docker container traffic
		{ID: 16, DateTime: "2024-08-06 10:00:16", Domain: "registry-1.docker.io", Type: 1, Client: "172.20.0.8", ResponseTime: 0.002456},
		{ID: 17, DateTime: "2024-08-06 10:00:17", Domain: "hub.docker.com", Type: 1, Client: "172.19.0.2", ResponseTime: 0.003123},
		{ID: 18, DateTime: "2024-08-06 10:00:18", Domain: "auth.docker.io", Type: 1, Client: "172.20.0.8", ResponseTime: 0.001789},
		// IPv6 and other addresses
		{ID: 19, DateTime: "2024-08-06 10:00:19", Domain: "ipv6.google.com", Type: 28, Client: "2001:db8::1", ResponseTime: 0.004567},
		{ID: 20, DateTime: "2024-08-06 10:00:20", Domain: "localhost", Type: 1, Client: "127.0.0.1", ResponseTime: 0.000045},
	}

	// Mock Pi-hole records (simulating database data)
	mock.PiholeRecords = []types.PiholeRecord{
		{Timestamp: strconv.FormatInt(time.Now().Unix() - 3600, 10), Client: "192.168.2.110", HWAddr: "66:78:8b:59:bf:1a", Domain: "google.com", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix() - 3500, 10), Client: "192.168.2.210", HWAddr: "32:79:b9:39:43:7c", Domain: "facebook.com", Status: 9},
		{Timestamp: strconv.FormatInt(time.Now().Unix() - 3400, 10), Client: "192.168.2.110", HWAddr: "66:78:8b:59:bf:1a", Domain: "youtube.com", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix() - 3300, 10), Client: "192.168.2.210", HWAddr: "32:79:b9:39:43:7c", Domain: "tracking.doubleclick.net", Status: 9},
		{Timestamp: strconv.FormatInt(time.Now().Unix() - 3200, 10), Client: "192.168.2.6", HWAddr: "e0:69:95:4f:19:6d", Domain: "api.spotify.com", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix() - 3100, 10), Client: "172.20.0.8", HWAddr: "02:42:ac:14:00:08", Domain: "docker.internal", Status: 3},
		{Timestamp: strconv.FormatInt(time.Now().Unix() - 3000, 10), Client: "192.168.2.110", HWAddr: "66:78:8b:59:bf:1a", Domain: "github.com", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix() - 2900, 10), Client: "192.168.2.210", HWAddr: "32:79:b9:39:43:7c", Domain: "ads.microsoft.com", Status: 9},
		{Timestamp: strconv.FormatInt(time.Now().Unix() - 2800, 10), Client: "192.168.2.202", HWAddr: "14:da:e9:48:53:a4", Domain: "cdn.jsdelivr.net", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix() - 2700, 10), Client: "192.168.2.119", HWAddr: "fc:49:2d:1a:41:16", Domain: "malware.badsite.com", Status: 9},
		{Timestamp: strconv.FormatInt(time.Now().Unix() - 2600, 10), Client: "192.168.2.202", HWAddr: "14:da:e9:48:53:a4", Domain: "netflix.com", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix() - 2500, 10), Client: "192.168.2.119", HWAddr: "fc:49:2d:1a:41:16", Domain: "amazon.com", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix() - 2400, 10), Client: "192.168.2.110", HWAddr: "66:78:8b:59:bf:1a", Domain: "microsoft.com", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix() - 2300, 10), Client: "192.168.2.210", HWAddr: "32:79:b9:39:43:7c", Domain: "telemetry.microsoft.com", Status: 9},
		{Timestamp: strconv.FormatInt(time.Now().Unix() - 2200, 10), Client: "192.168.2.6", HWAddr: "e0:69:95:4f:19:6d", Domain: "pi.hole", Status: 3},
		// More Docker container traffic
		{Timestamp: strconv.FormatInt(time.Now().Unix() - 2100, 10), Client: "172.20.0.8", HWAddr: "02:42:ac:14:00:08", Domain: "registry-1.docker.io", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix() - 2000, 10), Client: "172.19.0.2", HWAddr: "02:42:ac:13:00:02", Domain: "hub.docker.com", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix() - 1900, 10), Client: "172.20.0.8", HWAddr: "02:42:ac:14:00:08", Domain: "auth.docker.io", Status: 2},
		// Offline devices
		{Timestamp: strconv.FormatInt(time.Now().Unix() - 1800, 10), Client: "192.168.2.123", HWAddr: "fc:41:16:b4:22:33", Domain: "old.device.com", Status: 2},
		{Timestamp: strconv.FormatInt(time.Now().Unix() - 1700, 10), Client: "192.168.2.156", HWAddr: "f2:6f:b8:f3:44:55", Domain: "another.old.com", Status: 3},
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

// CreateMockCSVFile creates a CSV file with mock DNS data
func CreateMockCSVFile(filename string, mockData *MockData) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating mock CSV file: %v", err)
	}
	defer file.Close()

	// Write CSV header
	_, err = file.WriteString("ID,DateTime,Domain,Type,Status,Client,Forward,AdditionalInfo,ReplyType,ReplyTime,DNSSEC,ListID,EDE\n")
	if err != nil {
		return err
	}

	// Write mock data records
	for _, record := range mockData.types.DNSRecords {
		line := fmt.Sprintf("%d,%s,%s,%d,%d,%s,,%s,%d,%.6f,0,0,0\n",
			record.ID,
			record.DateTime,
			record.Domain,
			record.Type,
			record.Status,
			record.Client,
			record.AdditionalInfo,
			record.ReplyType,
			record.ReplyTime,
		)
		_, err = file.WriteString(line)
		if err != nil {
			return err
		}
	}

	return nil
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

	for _, record := range mockData.types.PiholeRecords {
		statusCode := getStatusCodeFromString(record.Status)
		_, err = db.Exec(queryInsert, int64(record.Timestamp), 1, statusCode, record.Domain, record.Client, "")
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
	testDir := "test_data"
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("error creating test directory: %v", err)
	}

	// Create mock CSV file
	csvFile := filepath.Join(testDir, "mock_dns_data.csv")
	err = CreateMockCSVFile(csvFile, mockData)
	if err != nil {
		return nil, fmt.Errorf("error creating mock CSV: %v", err)
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

	fmt.Printf("✓ Mock CSV file created: %s\n", csvFile)
	fmt.Printf("✓ Mock Pi-hole database created: %s\n", dbFile)
	fmt.Printf("✓ Mock configuration created: %s\n", configFile)
	fmt.Println("✓ Test environment setup complete!")

	return mockData, nil
}

// CleanupTestEnvironment removes all test files
func CleanupTestEnvironment() {
	os.RemoveAll("test_data")
	fmt.Println("✓ Test environment cleaned up")
}
