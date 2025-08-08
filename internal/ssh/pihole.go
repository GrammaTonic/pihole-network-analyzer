package ssh

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	_ "modernc.org/sqlite"

	"pihole-analyzer/internal/types"
)

// PiholeRecord represents a record from Pi-hole database (local definition)
type PiholeRecord struct {
	Timestamp float64
	Client    string
	HWAddr    string
	Domain    string
	Status    string
}

// AnalyzePiholeData connects to Pi-hole via SSH and analyzes the data
func AnalyzePiholeData(configFile string) (map[string]*types.ClientStats, error) {
	config, err := loadPiholeConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %v", err)
	}

	// Connect via SSH
	client, err := connectSSH(config)
	if err != nil {
		return nil, fmt.Errorf("SSH connection failed: %v", err)
	}
	defer client.Close()

	// Download Pi-hole database
	localDBPath := fmt.Sprintf("pihole-data-%d.db", time.Now().Unix())
	err = downloadFile(client, config.DBPath, localDBPath)
	if err != nil {
		return nil, fmt.Errorf("error downloading database: %v", err)
	}
	defer os.Remove(localDBPath) // Clean up

	// Analyze the database
	return AnalyzePiholeDatabase(localDBPath)
}

// connectSSH establishes SSH connection to Pi-hole server
func connectSSH(config *types.PiholeConfig) (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod

	// Try key-based authentication first
	if config.KeyFile != "" {
		key, err := os.ReadFile(config.KeyFile)
		if err == nil {
			signer, err := ssh.ParsePrivateKey(key)
			if err == nil {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
			}
		}
	}

	// Try SSH agent
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		authMethods = append(authMethods, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
		defer sshAgent.Close()
	}

	// Try password authentication
	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication methods available")
	}

	// SSH client configuration
	sshConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Note: In production, use proper host key verification
		Timeout:         30 * time.Second,
	}

	// Connect
	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	return ssh.Dial("tcp", address, sshConfig)
}

// downloadFile downloads a file from remote server via SSH
func downloadFile(client *ssh.Client, remotePath, localPath string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// Create local file
	localFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	// Use SCP to copy file
	session.Stdout = localFile
	return session.Run(fmt.Sprintf("cat %s", remotePath))
}

// AnalyzePiholeDatabase analyzes Pi-hole SQLite database
func AnalyzePiholeDatabase(dbPath string) (map[string]*types.ClientStats, error) {
	// Check if database file exists and is readable
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database file does not exist: %s", dbPath)
	}

	// Check file permissions
	file, err := os.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("cannot access database file: %v", err)
	}
	file.Close()

	db, err := sql.Open("sqlite", dbPath+"?cache=shared&mode=ro")
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	// Set connection limits for memory management
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Set SQLite pragmas for memory optimization
	_, err = db.Exec("PRAGMA cache_size = 2000") // Reduce cache size
	if err != nil {
		return nil, fmt.Errorf("error setting cache_size pragma: %v", err)
	}

	_, err = db.Exec("PRAGMA temp_store = memory") // Use memory for temp storage
	if err != nil {
		return nil, fmt.Errorf("error setting temp_store pragma: %v", err)
	}

	// Set soft heap limit (32MB for CI environments, 64MB otherwise)
	heapLimit := int64(64 * 1024 * 1024) // 64MB
	if os.Getenv("CI") != "" {
		heapLimit = 32 * 1024 * 1024 // 32MB in CI
	}
	_, err = db.Exec(fmt.Sprintf("PRAGMA soft_heap_limit = %d", heapLimit))
	if err != nil {
		return nil, fmt.Errorf("error setting soft_heap_limit pragma: %v", err)
	}

	// Test database connectivity
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("database ping failed: %v", err)
	}

	// Query DNS records from Pi-hole database
	// Use a smaller limit for test environments to avoid memory issues
	limit := 10000 // Reduced from 100000 for better memory management
	if os.Getenv("CI") != "" {
		limit = 1000 // Even smaller limit in CI environments
	}

	query := fmt.Sprintf(`
		SELECT 
			q.timestamp,
			q.client,
			COALESCE(n.name, q.client) as client_name,
			COALESCE(n.hwaddr, '') as hwaddr,
			q.domain,
			q.status
		FROM queries q
		LEFT JOIN network_addresses na ON q.client = na.ip
		LEFT JOIN network n ON na.network_id = n.id
		WHERE q.timestamp > ?
		ORDER BY q.timestamp DESC
		LIMIT %d
	`, limit)

	// Get data from last 24 hours
	yesterday := time.Now().Add(-24 * time.Hour).Unix()
	rows, err := db.Query(query, yesterday)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %v", err)
	}
	defer rows.Close()

	clientStats := make(map[string]*types.ClientStats)

	for rows.Next() {
		var record PiholeRecord
		var clientName sql.NullString
		var hwaddr sql.NullString
		var domain sql.NullString

		err := rows.Scan(
			&record.Timestamp,
			&record.Client,
			&clientName,
			&hwaddr,
			&domain,
			&record.Status,
		)
		if err != nil {
			continue // Skip invalid records
		}

		// Set optional fields
		if hwaddr.Valid {
			record.HWAddr = hwaddr.String
		}
		if domain.Valid {
			record.Domain = domain.String
		}

		updateClientStatsFromPihole(clientStats, &record)
	}

	return clientStats, nil
}

// updateClientStatsFromPihole updates client statistics with Pi-hole record
func updateClientStatsFromPihole(clientStats map[string]*types.ClientStats, record *PiholeRecord) {
	client := record.Client
	if client == "" {
		return
	}

	// Initialize client stats if not exists
	if clientStats[client] == nil {
		clientStats[client] = &types.ClientStats{
			Client:      client,
			HWAddr:      record.HWAddr,
			Domains:     make(map[string]int),
			QueryTypes:  make(map[int]int),
			StatusCodes: make(map[int]int),
		}
	}

	stats := clientStats[client]
	stats.TotalQueries++

	// Convert string status to status code
	statusCode := getStatusCodeFromString(record.Status)
	stats.StatusCodes[statusCode]++

	// Track domain queries
	if record.Domain != "" {
		stats.Domains[record.Domain]++
		if stats.Domains[record.Domain] == 1 {
			stats.UniqueQueries++
		}
	}

	// Update hardware address if available
	if record.HWAddr != "" && stats.HWAddr == "" {
		stats.HWAddr = record.HWAddr
	}
}

// getStatusCodeFromString converts Pi-hole status strings to codes
func getStatusCodeFromString(status string) int {
	switch status {
	case "2", "FORWARDED":
		return 2
	case "3", "CACHED":
		return 3
	case "1", "BLOCKED":
		return 1
	case "4", "WILDCARD":
		return 4
	case "5", "BLACKLIST":
		return 5
	case "6", "EXTERNAL":
		return 6
	case "7", "CNAME":
		return 7
	default:
		// Try to parse as integer
		if code, err := strconv.Atoi(status); err == nil {
			return code
		}
		return 0
	}
}

// loadPiholeConfig loads Pi-hole configuration from file
func loadPiholeConfig(configFile string) (*types.PiholeConfig, error) {
	// This is a placeholder - in a real implementation, this would load
	// configuration from a JSON file or similar
	return &types.PiholeConfig{
		Host:     "192.168.1.100",
		Port:     22,
		Username: "pi",
		DBPath:   "/etc/pihole/pihole-FTL.db",
	}, nil
}

// SetupPiholeConfig guides user through Pi-hole configuration setup
func SetupPiholeConfig() {
	fmt.Println("Pi-hole Configuration Setup")
	fmt.Println("===========================")
	fmt.Println()
	fmt.Println("This will guide you through setting up SSH connection to your Pi-hole server.")
	fmt.Println()

	// This would be a more interactive setup in the full implementation
	fmt.Println("Please create a configuration file with the following structure:")
	fmt.Println()
	fmt.Println(`{
  "pihole": {
    "host": "192.168.1.100",
    "port": 22,
    "username": "pi",
    "password": "your_password",
    "key_file": "/path/to/ssh/key",
    "db_path": "/etc/pihole/pihole-FTL.db"
  }
}`)
	fmt.Println()
	fmt.Println("Save this as 'pihole-config.json' and run:")
	fmt.Println("./pihole-analyzer --pihole pihole-config.json")
}
