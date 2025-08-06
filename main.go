package main

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	_ "modernc.org/sqlite"
)

// Command-line flags
var (
	onlineOnlyFlag   = flag.Bool("online-only", false, "Show only clients that are currently online (have MAC addresses in ARP table)")
	noExcludeFlag    = flag.Bool("no-exclude", false, "Disable default exclusions (Docker networks, Pi-hole host)")
	piholeFlag       = flag.String("pihole", "", "Analyze Pi-hole live data using the specified config file")
	piholeSetupFlag  = flag.Bool("pihole-setup", false, "Setup Pi-hole configuration")
	testFlag         = flag.Bool("test", false, "Run test suite with mock data")
	testModeFlag     = flag.Bool("test-mode", false, "Enable test mode for development (uses mock data)")
	configFlag       = flag.String("config", "", "Configuration file path (default: ~/.dns-analyzer/config.json)")
	showConfigFlag   = flag.Bool("show-config", false, "Show current configuration and exit")
	createConfigFlag = flag.Bool("create-config", false, "Create default configuration file and exit")
)

// DNSRecord represents a single DNS query record
type DNSRecord struct {
	ID             int
	DateTime       string
	Domain         string
	Type           int
	Status         int
	Client         string
	Forward        string
	AdditionalInfo string
	ReplyType      int
	ReplyTime      float64
	DNSSEC         int
	ListID         int
	EDE            int
}

// PiholeRecord represents a record from Pi-hole database
type PiholeRecord struct {
	Timestamp float64
	Client    string
	HWAddr    string
	Domain    string
	Status    string
}

// PiholeConfig holds SSH connection details for Pi-hole
type PiholeConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	KeyFile  string `json:"keyfile"`
	DBPath   string `json:"dbpath"`
}

// ExclusionConfig holds the exclusion rules
type ExclusionConfig struct {
	ExcludeNetworks []string `json:"exclude_networks"` // CIDR networks to exclude
	ExcludeIPs      []string `json:"exclude_ips"`      // Specific IPs to exclude
	ExcludeHosts    []string `json:"exclude_hosts"`    // Hostnames to exclude
}

// ARPEntry represents an entry from the ARP table
type ARPEntry struct {
	IP       string
	HWAddr   string
	Status   string
	IsOnline bool
}

// ClientStats holds statistics for a client
type ClientStats struct {
	Client         string
	TotalQueries   int
	Uniquedomains  int
	TotalReplyTime float64
	AvgReplyTime   float64
	Domains        map[string]int
	QueryTypes     map[int]int
	StatusCodes    map[int]int
	HWAddr         string // Added for Pi-hole data
	IsOnline       bool   // Added for ARP checking
	ARPStatus      string // Added for ARP status
	Hostname       string // Added for hostname resolution
}

// ClientStatsList for sorting
type ClientStatsList []ClientStats

func (c ClientStatsList) Len() int           { return len(c) }
func (c ClientStatsList) Less(i, j int) bool { return c[i].TotalQueries > c[j].TotalQueries }
func (c ClientStatsList) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

func main() {
	// Parse command-line flags
	flag.Parse()

	// Handle configuration-related flags first
	configPath := GetConfigPath()
	if *configFlag != "" {
		configPath = *configFlag
	}

	if *createConfigFlag {
		err := CreateDefaultConfigFile(configPath)
		if err != nil {
			log.Fatalf("Error creating config file: %v", err)
		}
		fmt.Printf("Default configuration created at: %s\n", configPath)
		return
	}

	// Load configuration
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Merge command line flags with config
	MergeFlags(config)

	if *showConfigFlag {
		ShowConfig(config)
		return
	}

	// Handle test mode first
	if *testFlag {
		RunTests()
		return
	}

	if config.TestMode {
		fmt.Println("🧪 Test Mode Enabled - Using Mock Data")
		err := InitTestMode()
		if err != nil {
			log.Fatalf("Failed to initialize test mode: %v", err)
		}
		defer CleanupTestEnvironment()
	}

	// Get non-flag arguments
	args := flag.Args()

	// Handle Pi-hole setup first
	if *piholeSetupFlag {
		setupPiholeConfig()
		return
	}

	// Handle Pi-hole analysis
	if *piholeFlag != "" {
		configFile := *piholeFlag
		fmt.Printf("Connecting to Pi-hole using config: %s\n", configFile)
		if config.OnlineOnly {
			fmt.Println("Mode: Online clients only")
		}
		if config.NoExclude {
			fmt.Println("Mode: No exclusions applied")
		}

		var clientStats map[string]*ClientStats
		var err error

		if config.TestMode {
			// Use mock Pi-hole database in test mode
			dbFile := filepath.Join("test_data", "mock_pihole.db")
			clientStats, err = analyzePiholeDatabase(dbFile)
		} else {
			clientStats, err = analyzePiholeData(configFile)
		}

		if err != nil {
			log.Fatalf("Error analyzing Pi-hole data: %v", err)
		}

		// Check ARP status for all clients
		if config.TestMode {
			err = mockCheckARPStatus(clientStats)
		} else {
			err = checkARPStatus(clientStats)
		}
		if err != nil {
			fmt.Printf("Warning: Could not check ARP status: %v\n", err)
		}

		displayResultsWithConfig(clientStats, config)
		return
	}

	// Require CSV file if no Pi-hole flags specified
	if len(args) < 1 {
		fmt.Println("DNS Usage Analyzer")
		fmt.Println("Usage:")
		fmt.Println("  dns-analyzer [flags] <csv_file>                    # Analyze CSV file")
		fmt.Println("  dns-analyzer --pihole <config_file> [flags]        # Analyze Pi-hole live data")
		fmt.Println("  dns-analyzer --pihole-setup                       # Setup Pi-hole configuration")
		fmt.Println("  dns-analyzer --test                               # Run test suite")
		fmt.Println("  dns-analyzer --test-mode [other flags]           # Use mock data for testing")
		fmt.Println("  dns-analyzer --create-config                     # Create default config file")
		fmt.Println("  dns-analyzer --show-config                       # Show current configuration")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  --online-only    Show only clients currently online (with MAC addresses in ARP)")
		fmt.Println("  --no-exclude     Disable default exclusions (Docker networks, Pi-hole host)")
		fmt.Println("  --test-mode      Enable test mode (uses mock data instead of real network)")
		fmt.Println("  --config <path>  Use specific configuration file")
		fmt.Println()
		fmt.Println("Configuration:")
		fmt.Printf("  Default config: %s\n", configPath)
		fmt.Println("  Use --create-config to generate default configuration")
		fmt.Println("  Use --show-config to view current settings")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  dns-analyzer test.csv")
		fmt.Println("  dns-analyzer --online-only --config my-config.json test.csv")
		fmt.Println("  dns-analyzer --no-exclude test.csv")
		fmt.Println("  dns-analyzer --test-mode --online-only test_data/mock_dns_data.csv")
		fmt.Println("  dns-analyzer --config my-config.json test.csv")
		os.Exit(1)
	}

	// Default: CSV analysis
	csvFile := args[0]

	// In test mode, use mock CSV file if original file doesn't exist
	if config.TestMode && csvFile == "test.csv" {
		csvFile = filepath.Join("test_data", "mock_dns_data.csv")
	}

	fmt.Printf("Analyzing DNS usage data from: %s\n", csvFile)
	fmt.Println("Processing large file, please wait...")
	if config.OnlineOnly {
		fmt.Println("Mode: Online clients only")
	}
	if config.NoExclude {
		fmt.Println("Mode: No exclusions applied")
	}

	clientStats, err := analyzeDNSDataWithConfig(csvFile, config)
	if err != nil {
		log.Fatalf("Error analyzing data: %v", err)
	}

	// Check ARP status for all clients
	if config.TestMode {
		err = mockCheckARPStatus(clientStats)
	} else {
		err = checkARPStatus(clientStats)
	}
	if err != nil {
		fmt.Printf("Warning: Could not check ARP status: %v\n", err)
	}

	displayResultsWithConfig(clientStats, config)
}

func setupPiholeConfig() {
	fmt.Println("Pi-hole Configuration Setup")
	fmt.Println("============================")

	piholeConfig := PiholeConfig{}
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Pi-hole server IP/hostname: ")
	piholeConfig.Host, _ = reader.ReadString('\n')
	piholeConfig.Host = strings.TrimSpace(piholeConfig.Host)

	fmt.Print("SSH port (default 22): ")
	port, _ := reader.ReadString('\n')
	port = strings.TrimSpace(port)
	if port == "" {
		port = "22"
	}
	piholeConfig.Port = port

	fmt.Print("SSH username (default pi): ")
	piholeConfig.Username, _ = reader.ReadString('\n')
	piholeConfig.Username = strings.TrimSpace(piholeConfig.Username)
	if piholeConfig.Username == "" {
		piholeConfig.Username = "pi"
	}

	fmt.Print("Use SSH key authentication? (y/n): ")
	authChoice, _ := reader.ReadString('\n')
	authChoice = strings.TrimSpace(strings.ToLower(authChoice))

	if authChoice == "y" || authChoice == "yes" {
		fmt.Print("SSH private key file path (default ~/.ssh/id_rsa): ")
		piholeConfig.KeyFile, _ = reader.ReadString('\n')
		piholeConfig.KeyFile = strings.TrimSpace(piholeConfig.KeyFile)
		if piholeConfig.KeyFile == "" {
			homeDir, _ := os.UserHomeDir()
			piholeConfig.KeyFile = filepath.Join(homeDir, ".ssh", "id_rsa")
		}
	} else {
		fmt.Print("SSH password: ")
		piholeConfig.Password, _ = reader.ReadString('\n')
		piholeConfig.Password = strings.TrimSpace(piholeConfig.Password)
	}

	fmt.Print("Pi-hole database path (default /etc/pihole/pihole-FTL.db): ")
	piholeConfig.DBPath, _ = reader.ReadString('\n')
	piholeConfig.DBPath = strings.TrimSpace(piholeConfig.DBPath)
	if piholeConfig.DBPath == "" {
		piholeConfig.DBPath = "/etc/pihole/pihole-FTL.db"
	}

	// Load or create the main configuration
	configPath := GetConfigPath()
	config := DefaultConfig()

	// Try to load existing config
	if data, err := ioutil.ReadFile(configPath); err == nil {
		json.Unmarshal(data, config)
	}

	// Update Pi-hole configuration
	config.Pihole = piholeConfig

	// Save updated configuration
	if err := SaveConfig(config, configPath); err != nil {
		log.Fatalf("Error saving configuration: %v", err)
	}

	fmt.Printf("\nPi-hole configuration updated in: %s\n", configPath)
	fmt.Println("You can now run: dns-analyzer --config config.json test.csv")
	fmt.Println("Or use: dns-analyzer --show-config to view all settings")
}

func analyzePiholeData(configFile string) (map[string]*ClientStats, error) {
	fmt.Println("Connecting to Pi-hole server...")

	// Read configuration
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	config := PiholeConfig{}
	// Simple JSON parsing (could use json package for more robust parsing)
	lines := strings.Split(string(configData), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, `"host":`) {
			config.Host = extractJSONValue(line)
		} else if strings.Contains(line, `"port":`) {
			config.Port = extractJSONValue(line)
		} else if strings.Contains(line, `"username":`) {
			config.Username = extractJSONValue(line)
		} else if strings.Contains(line, `"password":`) {
			config.Password = extractJSONValue(line)
		} else if strings.Contains(line, `"keyfile":`) {
			config.KeyFile = extractJSONValue(line)
		} else if strings.Contains(line, `"dbpath":`) {
			config.DBPath = extractJSONValue(line)
		}
	}

	// Create SSH client
	sshConfig := &ssh.ClientConfig{
		User:            config.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // In production, use proper host key verification
		Timeout:         30 * time.Second,
	}

	// Try SSH agent first, then key file, then password
	var authMethods []ssh.AuthMethod

	// Try SSH agent (for 1Password and other SSH agent integrations)
	if sshAuthSock := os.Getenv("SSH_AUTH_SOCK"); sshAuthSock != "" {
		if agentConn, err := net.Dial("unix", sshAuthSock); err == nil {
			agentClient := agent.NewClient(agentConn)
			authMethods = append(authMethods, ssh.PublicKeysCallback(agentClient.Signers))
			// Don't close the connection here - it's needed during authentication
			defer agentConn.Close()
		}
	}

	// Try specific key file if provided
	if config.KeyFile != "" {
		key, err := ioutil.ReadFile(config.KeyFile)
		if err == nil {
			if signer, err := ssh.ParsePrivateKey(key); err == nil {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
			}
		}
	}

	// Try password if provided
	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication methods available - please configure SSH key, SSH agent, or password")
	}

	sshConfig.Auth = authMethods

	// Connect to SSH
	client, err := ssh.Dial("tcp", net.JoinHostPort(config.Host, config.Port), sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH: %v", err)
	}
	defer client.Close()

	fmt.Println("Connected to Pi-hole server successfully!")

	// Copy database file to temporary location for analysis
	tempDB := fmt.Sprintf("/tmp/pihole-analysis-%d.db", time.Now().Unix())
	copyCmd := fmt.Sprintf("cp %s %s && chmod 644 %s", config.DBPath, tempDB, tempDB)

	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	err = session.Run(copyCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to copy database: %v", err)
	}

	// Download the database file via SCP-like operation
	localDBPath := fmt.Sprintf("pihole-data-%d.db", time.Now().Unix())
	err = downloadFile(client, tempDB, localDBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to download database: %v", err)
	}

	// Clean up remote temp file
	session2, _ := client.NewSession()
	session2.Run(fmt.Sprintf("rm -f %s", tempDB))
	session2.Close()

	fmt.Printf("Database downloaded to: %s\n", localDBPath)

	// Analyze the database
	clientStats, err := analyzePiholeDatabase(localDBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze database: %v", err)
	}

	// Clean up local database file
	os.Remove(localDBPath)

	return clientStats, nil
}

func extractJSONValue(line string) string {
	// Simple JSON value extraction
	parts := strings.Split(line, ":")
	if len(parts) >= 2 {
		value := strings.Join(parts[1:], ":")
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `",`)
		return value
	}
	return ""
}

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

	// Use cat command to stream file content
	session.Stdout = localFile
	return session.Run(fmt.Sprintf("cat %s", remotePath))
}

func analyzePiholeDatabase(dbPath string) (map[string]*ClientStats, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	// Get exclusion configuration with default settings
	exclusions := getDefaultExclusions("192.168.2.6") // Default Pi-hole IP, will be updated

	// Execute the Pi-hole query
	query := `
	SELECT
		q.timestamp,
		q.client,
		COALESCE(n.hwaddr, '') as hwaddr,
		q.domain,
		CASE q.status
			WHEN 0 THEN 'Unknown'
			WHEN 1 THEN 'Blocked (gravity)'
			WHEN 2 THEN 'Forwarded'
			WHEN 3 THEN 'Cached'
			WHEN 4 THEN 'Blocked (regex/wildcard)'
			WHEN 5 THEN 'Blocked (exact)'
			WHEN 6 THEN 'Blocked (external, IP)'
			WHEN 7 THEN 'Blocked (external, NULL)'
			WHEN 8 THEN 'Blocked (external, NXDOMAIN)'
			WHEN 9 THEN 'Blocked (gravity, CNAME)'
			WHEN 10 THEN 'Blocked (regex/wildcard, CNAME)'
			WHEN 11 THEN 'Blocked (exact, CNAME)'
			ELSE 'Unknown'
		END AS status
	FROM
		queries q
	LEFT JOIN
		network_addresses na ON q.client = na.ip
	LEFT JOIN
		network n ON na.network_id = n.id
	ORDER BY q.timestamp DESC
	LIMIT 100000`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	clientStats := make(map[string]*ClientStats)
	recordCount := 0
	excludedCount := 0
	excludedIPs := make(map[string]bool) // Cache of excluded IPs to avoid repeated checks

	fmt.Println("Processing Pi-hole database records...")

	for rows.Next() {
		var record PiholeRecord
		err := rows.Scan(&record.Timestamp, &record.Client, &record.HWAddr, &record.Domain, &record.Status)
		if err != nil {
			log.Printf("Error scanning record: %v", err)
			continue
		}

		recordCount++

		// Check if this IP has already been determined to be excluded
		if excluded, exists := excludedIPs[record.Client]; exists && excluded {
			excludedCount++
			continue
		}

		// Check if this client should be excluded (only check once per IP)
		if _, exists := excludedIPs[record.Client]; !exists {
			// Only apply exclusions if not disabled by flag
			if !*noExcludeFlag {
				if shouldExclude, reason := shouldExcludeClient(record.Client, "", exclusions); shouldExclude {
					excludedIPs[record.Client] = true
					excludedCount++
					fmt.Printf("Excluded: %s\n", reason)
					continue
				} else {
					excludedIPs[record.Client] = false
				}
			} else {
				// When exclusions are disabled, mark all IPs as not excluded
				excludedIPs[record.Client] = false
			}
		}

		// Update client statistics
		updateClientStatsFromPihole(clientStats, &record)

		if recordCount%10000 == 0 {
			fmt.Printf("Processed %d records (%d excluded)...\n", recordCount, excludedCount)
		}
	}

	fmt.Printf("Total Pi-hole records processed: %d (excluded: %d)\n", recordCount, excludedCount)

	// Calculate unique domains count
	for _, stats := range clientStats {
		stats.Uniquedomains = len(stats.Domains)
	}

	return clientStats, nil
}

func updateClientStatsFromPihole(clientStats map[string]*ClientStats, record *PiholeRecord) {
	client := record.Client

	if clientStats[client] == nil {
		clientStats[client] = &ClientStats{
			Client:      client,
			Domains:     make(map[string]int),
			QueryTypes:  make(map[int]int),
			StatusCodes: make(map[int]int),
			HWAddr:      record.HWAddr,
			IsOnline:    false,
			ARPStatus:   "unknown",
			Hostname:    "",
		}
	}

	stats := clientStats[client]
	stats.TotalQueries++
	stats.Domains[record.Domain]++

	// Convert status string to status code for consistency
	statusCode := getStatusCodeFromString(record.Status)
	stats.StatusCodes[statusCode]++

	// Set HWAddr if we have it
	if record.HWAddr != "" {
		stats.HWAddr = record.HWAddr
	}
}

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
	case "Blocked (external, IP)":
		return 6
	case "Blocked (external, NULL)":
		return 7
	case "Blocked (external, NXDOMAIN)":
		return 8
	case "Blocked (gravity, CNAME)":
		return 9
	case "Blocked (regex/wildcard, CNAME)":
		return 10
	case "Blocked (exact, CNAME)":
		return 11
	default:
		return 0
	}
}

func getLocalARPTable() (map[string]*ARPEntry, error) {
	// Try different ARP commands based on OS
	var cmd *exec.Cmd

	// For macOS and Linux
	cmd = exec.Command("arp", "-a")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get ARP table: %v", err)
	}

	return parseARPOutput(string(output))
}

func parseARPOutput(output string) (map[string]*ARPEntry, error) {
	arpEntries := make(map[string]*ARPEntry)
	lines := strings.Split(output, "\n")

	// Regex patterns for different ARP output formats
	// macOS format: "hostname (192.168.1.1) at aa:bb:cc:dd:ee:ff on en0 ifscope [ethernet]"
	// Linux format: "192.168.1.1 ether aa:bb:cc:dd:ee:ff C eth0"
	macOSPattern := regexp.MustCompile(`\(([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)\) at ([a-fA-F0-9:]{17})`)
	linuxPattern := regexp.MustCompile(`([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)\s+\w+\s+([a-fA-F0-9:]{17})\s+(\w+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var ip, hwAddr, status string
		var isOnline bool

		// Try macOS format first
		if matches := macOSPattern.FindStringSubmatch(line); len(matches) >= 3 {
			ip = matches[1]
			hwAddr = strings.ToUpper(matches[2])
			status = "reachable"
			isOnline = true
		} else if matches := linuxPattern.FindStringSubmatch(line); len(matches) >= 4 {
			// Try Linux format
			ip = matches[1]
			hwAddr = strings.ToUpper(matches[2])
			status = matches[3]
			isOnline = status == "C" || status == "M" // C = complete, M = permanent
		} else {
			continue
		}

		arpEntries[ip] = &ARPEntry{
			IP:       ip,
			HWAddr:   hwAddr,
			Status:   status,
			IsOnline: isOnline,
		}
	}

	return arpEntries, nil
}

func pingHost(ip string) bool {
	// Try to ping the host to refresh ARP table
	var cmd *exec.Cmd

	// Use ping with 1 packet and 1 second timeout
	cmd = exec.Command("ping", "-c", "1", "-W", "1000", ip)

	err := cmd.Run()
	return err == nil
}

func arpPing(ip string) bool {
	// Use arping to send ARP requests directly (more reliable for ARP table population)
	// First try arping if available, fallback to regular ping
	var cmd *exec.Cmd

	// Try arping first (if installed via brew install arping)
	cmd = exec.Command("arping", "-c", "1", "-W", "1000", ip)
	err := cmd.Run()
	if err == nil {
		return true
	}

	// Fallback to regular ping
	return pingHost(ip)
}

func refreshARPTable(clientIPs []string) {
	fmt.Println("Refreshing ARP table by sending ARP requests to clients...")

	// Channel to limit concurrent pings
	semaphore := make(chan struct{}, 10) // Limit to 10 concurrent pings
	var wg sync.WaitGroup

	for _, ip := range clientIPs {
		// Skip if it's obviously not a valid IPv4 address
		if strings.Contains(ip, ":") || ip == "" || !isValidIPv4(ip) {
			continue
		}

		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			arpPing(ip) // Use ARP ping instead of regular ping
		}(ip)
	}

	// Wait for all pings to complete
	wg.Wait()
	fmt.Println("ARP table refresh completed")
}

func isValidIPv4(ip string) bool {
	// Simple IPv4 validation
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if num, err := strconv.Atoi(part); err != nil || num < 0 || num > 255 {
			return false
		}
	}
	return true
}

func resolveHostname(ip string) string {
	// Try to resolve hostname for the IP address
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return "" // Return empty string if resolution fails
	}

	// Return the first hostname, removing trailing dot if present
	hostname := names[0]
	if strings.HasSuffix(hostname, ".") {
		hostname = hostname[:len(hostname)-1]
	}
	return hostname
}

func resolveHostnames(clientStats map[string]*ClientStats) {
	fmt.Println("Resolving hostnames for clients...")

	// Channel to limit concurrent DNS lookups
	semaphore := make(chan struct{}, 5) // Limit to 5 concurrent lookups
	var wg sync.WaitGroup

	for ip, stats := range clientStats {
		// Skip if it's obviously not a valid IPv4 address
		if strings.Contains(ip, ":") || ip == "" || !isValidIPv4(ip) {
			continue
		}

		wg.Add(1)
		go func(ip string, stats *ClientStats) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			hostname := resolveHostname(ip)
			if hostname != "" {
				stats.Hostname = hostname
			}
		}(ip, stats)
	}

	// Wait for all lookups to complete
	wg.Wait()
	fmt.Println("Hostname resolution completed")
}

// getDefaultExclusions returns the default exclusion configuration
func getDefaultExclusions(piholeHost string) *ExclusionConfig {
	return &ExclusionConfig{
		ExcludeNetworks: []string{
			"172.16.0.0/12", // Docker default networks
			"127.0.0.0/8",   // Loopback
		},
		ExcludeIPs: []string{
			piholeHost, // Exclude the Pi-hole host itself
		},
		ExcludeHosts: []string{
			"pi.hole",
		},
	}
}

// isIPInNetwork checks if an IP is within a CIDR network
func isIPInNetwork(ip, cidr string) bool {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}

	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return false
	}

	return network.Contains(ipAddr)
}

// shouldExcludeClient checks if a client should be excluded based on exclusion rules
func shouldExcludeClient(ip, hostname string, exclusions *ExclusionConfig) (bool, string) {
	// Check excluded IPs
	for _, excludeIP := range exclusions.ExcludeIPs {
		if ip == excludeIP {
			return true, fmt.Sprintf("IP %s is in exclusion list", ip)
		}
	}

	// Check excluded networks
	for _, network := range exclusions.ExcludeNetworks {
		if isIPInNetwork(ip, network) {
			return true, fmt.Sprintf("IP %s is in excluded network %s", ip, network)
		}
	}

	// Check excluded hostnames
	if hostname != "" {
		for _, excludeHost := range exclusions.ExcludeHosts {
			if hostname == excludeHost {
				return true, fmt.Sprintf("Hostname %s is in exclusion list", hostname)
			}
		}
	}

	return false, ""
}

// analyzeDNSDataWithConfig analyzes DNS data using configuration settings
func analyzeDNSDataWithConfig(filename string, config *Config) (map[string]*ClientStats, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	clientStats := make(map[string]*ClientStats)
	recordCount := 0
	excludedCount := 0
	excludedIPs := make(map[string]bool)

	// Use exclusion configuration from config
	exclusions := &config.Exclusions

	// Skip header
	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading header: %v", err)
	}

	fmt.Println("Processing CSV records with exclusions...")

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading record %d: %v", recordCount, err)
			continue
		}

		if len(record) < 6 {
			continue // Skip malformed records
		}

		// Parse the record
		dnsRecord, err := parseRecord(record)
		if err != nil {
			log.Printf("Error parsing record %d: %v", recordCount, err)
			continue
		}

		recordCount++

		// Check if this IP has already been determined to be excluded
		if excluded, exists := excludedIPs[dnsRecord.Client]; exists && excluded {
			excludedCount++
			continue
		}

		// Check if this client should be excluded (only check once per IP)
		if _, exists := excludedIPs[dnsRecord.Client]; !exists {
			// Only apply exclusions if not disabled by config
			if !config.NoExclude {
				if shouldExclude, reason := shouldExcludeClient(dnsRecord.Client, "", exclusions); shouldExclude {
					excludedIPs[dnsRecord.Client] = true
					excludedCount++
					if excludedCount <= 10 {
						fmt.Printf("Excluded: %s\n", reason)
					}
					continue
				} else {
					excludedIPs[dnsRecord.Client] = false
				}
			} else {
				// When exclusions are disabled, mark all IPs as not excluded
				excludedIPs[dnsRecord.Client] = false
			}
		}

		// Update client statistics
		updateClientStats(clientStats, dnsRecord)

		if recordCount%100000 == 0 {
			fmt.Printf("Processed %d records (%d excluded)...\n", recordCount, excludedCount)
		}
	}

	fmt.Printf("Total records processed: %d (excluded: %d)\n", recordCount, excludedCount)

	// Calculate average reply times
	for _, stats := range clientStats {
		if stats.TotalQueries > 0 {
			stats.AvgReplyTime = stats.TotalReplyTime / float64(stats.TotalQueries)
			stats.Uniquedomains = len(stats.Domains)
		}
	}

	return clientStats, nil
}

func analyzeDNSData(filename string) (map[string]*ClientStats, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	clientStats := make(map[string]*ClientStats)
	recordCount := 0
	excludedCount := 0
	excludedIPs := make(map[string]bool)

	// Get exclusion configuration
	exclusions := getDefaultExclusions("192.168.2.6") // Default Pi-hole IP

	// Skip header
	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading header: %v", err)
	}

	fmt.Println("Processing CSV records with exclusions...")

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading record %d: %v", recordCount, err)
			continue
		}

		if len(record) < 6 {
			continue // Skip malformed records
		}

		// Parse the record
		dnsRecord, err := parseRecord(record)
		if err != nil {
			log.Printf("Error parsing record %d: %v", recordCount, err)
			continue
		}

		recordCount++

		// Check if this IP has already been determined to be excluded
		if excluded, exists := excludedIPs[dnsRecord.Client]; exists && excluded {
			excludedCount++
			continue
		}

		// Check if this client should be excluded (only check once per IP)
		if _, exists := excludedIPs[dnsRecord.Client]; !exists {
			// Only apply exclusions if not disabled by flag
			if !*noExcludeFlag {
				if shouldExclude, reason := shouldExcludeClient(dnsRecord.Client, "", exclusions); shouldExclude {
					excludedIPs[dnsRecord.Client] = true
					excludedCount++
					if excludedCount <= 10 {
						fmt.Printf("Excluded: %s\n", reason)
					}
					continue
				} else {
					excludedIPs[dnsRecord.Client] = false
				}
			} else {
				// When exclusions are disabled, mark all IPs as not excluded
				excludedIPs[dnsRecord.Client] = false
			}
		}

		// Update client statistics
		updateClientStats(clientStats, dnsRecord)

		if recordCount%100000 == 0 {
			fmt.Printf("Processed %d records (%d excluded)...\n", recordCount, excludedCount)
		}
	}

	fmt.Printf("Total records processed: %d (excluded: %d)\n", recordCount, excludedCount)

	// Calculate average reply times
	for _, stats := range clientStats {
		if stats.TotalQueries > 0 {
			stats.AvgReplyTime = stats.TotalReplyTime / float64(stats.TotalQueries)
			stats.Uniquedomains = len(stats.Domains)
		}
	}

	return clientStats, nil
}

func parseRecord(record []string) (*DNSRecord, error) {
	var err error
	dnsRecord := &DNSRecord{}

	// Parse ID
	if dnsRecord.ID, err = strconv.Atoi(record[0]); err != nil {
		return nil, fmt.Errorf("invalid ID: %v", err)
	}

	// DateTime (keep as string for now)
	dnsRecord.DateTime = record[1]

	// Domain
	dnsRecord.Domain = record[2]

	// Type
	if dnsRecord.Type, err = strconv.Atoi(record[3]); err != nil {
		return nil, fmt.Errorf("invalid type: %v", err)
	}

	// Status
	if dnsRecord.Status, err = strconv.Atoi(record[4]); err != nil {
		return nil, fmt.Errorf("invalid status: %v", err)
	}

	// Client
	dnsRecord.Client = record[5]

	// Forward (optional)
	if len(record) > 6 {
		dnsRecord.Forward = record[6]
	}

	// Additional info (optional)
	if len(record) > 7 {
		dnsRecord.AdditionalInfo = record[7]
	}

	// Reply type (optional)
	if len(record) > 8 && record[8] != "" {
		if dnsRecord.ReplyType, err = strconv.Atoi(record[8]); err != nil {
			dnsRecord.ReplyType = 0
		}
	}

	// Reply time (optional)
	if len(record) > 9 && record[9] != "" {
		if dnsRecord.ReplyTime, err = strconv.ParseFloat(record[9], 64); err != nil {
			dnsRecord.ReplyTime = 0
		}
	}

	return dnsRecord, nil
}

func updateClientStats(clientStats map[string]*ClientStats, record *DNSRecord) {
	client := record.Client

	if clientStats[client] == nil {
		clientStats[client] = &ClientStats{
			Client:      client,
			Domains:     make(map[string]int),
			QueryTypes:  make(map[int]int),
			StatusCodes: make(map[int]int),
			HWAddr:      "", // Not available in CSV data
			IsOnline:    false,
			ARPStatus:   "unknown",
			Hostname:    "",
		}
	}

	stats := clientStats[client]
	stats.TotalQueries++
	stats.TotalReplyTime += record.ReplyTime
	stats.Domains[record.Domain]++
	stats.QueryTypes[record.Type]++
	stats.StatusCodes[record.Status]++
}

// checkARPStatus updates the ARP status for all clients
func checkARPStatus(clientStats map[string]*ClientStats) error {
	fmt.Println("Checking ARP table for device availability...")

	// First, collect all client IPs
	var clientIPs []string
	for ip := range clientStats {
		clientIPs = append(clientIPs, ip)
	}

	// Refresh ARP table by sending ARP requests
	refreshARPTable(clientIPs)

	// Now get the updated ARP table
	arpMap, err := getLocalARPTable()
	if err != nil {
		return fmt.Errorf("failed to get ARP table: %v", err)
	}

	onlineCount := 0
	for ip, stats := range clientStats {
		if arpEntry, exists := arpMap[ip]; exists {
			stats.IsOnline = true
			stats.ARPStatus = "online"
			onlineCount++

			// If we don't have a hardware address from DNS logs but have one from ARP, use it
			if stats.HWAddr == "" && arpEntry.HWAddr != "" {
				stats.HWAddr = arpEntry.HWAddr
			}
		} else {
			stats.IsOnline = false
			stats.ARPStatus = "offline"
		}
	}

	// Resolve hostnames for all clients
	resolveHostnames(clientStats)

	fmt.Printf("Found %d/%d clients online in ARP table\n", onlineCount, len(clientStats))
	return nil
}

// displayResultsWithConfig displays results using configuration settings
func displayResultsWithConfig(clientStats map[string]*ClientStats, config *Config) {
	// Convert to slice for sorting
	var statsList ClientStatsList
	for _, stats := range clientStats {
		// If online-only is enabled, only include clients that are online
		if config.OnlineOnly {
			if stats.IsOnline {
				statsList = append(statsList, *stats)
			}
		} else {
			statsList = append(statsList, *stats)
		}
	}

	// Sort by total queries (descending)
	sort.Sort(statsList)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("DNS USAGE ANALYSIS BY CLIENT")
	if config.OnlineOnly {
		fmt.Println("(Showing only online clients)")
	}
	fmt.Println(strings.Repeat("=", 80))

	// Summary
	fmt.Printf("Total unique clients: %d\n", len(statsList))
	totalQueries := 0
	for _, stats := range statsList {
		totalQueries += stats.TotalQueries
	}
	fmt.Printf("Total DNS queries: %d\n", totalQueries)
	fmt.Println()

	// Top clients by query count (configurable)
	maxDisplay := config.Output.MaxClients
	fmt.Printf("TOP %d CLIENTS BY QUERY COUNT:\n", maxDisplay)
	fmt.Println(strings.Repeat("-", 110))
	fmt.Printf("%-25s %-20s %-10s %-12s %-12s %-8s %-8s\n", "Client IP", "Hostname", "Queries", "Domains", "Avg Reply", "% Total", "Online")
	fmt.Println(strings.Repeat("-", 110))

	if len(statsList) < maxDisplay {
		maxDisplay = len(statsList)
	}

	for i := 0; i < maxDisplay; i++ {
		stats := statsList[i]
		percentage := float64(stats.TotalQueries) / float64(totalQueries) * 100
		clientDisplay := stats.Client
		if stats.HWAddr != "" {
			clientDisplay = fmt.Sprintf("%s (%s)", stats.Client, stats.HWAddr[:12]+"...") // Show first 12 chars of MAC
		}

		// Prepare hostname display
		hostname := stats.Hostname
		if hostname == "" {
			hostname = "-"
		} else if len(hostname) > 18 {
			hostname = hostname[:15] + "..."
		}

		onlineStatus := "Unknown"
		if stats.ARPStatus != "unknown" {
			if stats.IsOnline {
				onlineStatus = "✓ Online"
			} else {
				onlineStatus = "✗ Offline"
			}
		}

		fmt.Printf("%-25s %-20s %-10d %-12d %-12.6f %-8.2f%% %-8s\n",
			clientDisplay,
			hostname,
			stats.TotalQueries,
			stats.Uniquedomains,
			stats.AvgReplyTime,
			percentage,
			onlineStatus)
	}

	// Detailed analysis for top 5 clients
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("DETAILED ANALYSIS - TOP 5 CLIENTS")
	fmt.Println(strings.Repeat("=", 80))

	maxDetailed := 5
	if len(statsList) < maxDetailed {
		maxDetailed = len(statsList)
	}

	for i := 0; i < maxDetailed; i++ {
		stats := statsList[i]
		fmt.Printf("\n%d. CLIENT: %s\n", i+1, stats.Client)
		if stats.Hostname != "" {
			fmt.Printf("   Hostname: %s\n", stats.Hostname)
		}
		if stats.HWAddr != "" {
			fmt.Printf("   Hardware Address: %s\n", stats.HWAddr)
		}

		// Show ARP status
		if stats.ARPStatus != "unknown" {
			statusIcon := "✗"
			if stats.IsOnline {
				statusIcon = "✓"
			}
			fmt.Printf("   Network Status: %s %s\n", statusIcon, strings.Title(stats.ARPStatus))
		}

		fmt.Printf("   Total Queries: %d\n", stats.TotalQueries)
		fmt.Printf("   Unique Domains: %d\n", stats.Uniquedomains)
		fmt.Printf("   Average Reply Time: %.6f seconds\n", stats.AvgReplyTime)

		// Top domains for this client (configurable)
		maxDomains := config.Output.MaxDomains
		fmt.Printf("   Top %d Domains:\n", maxDomains)
		domainList := make([]struct {
			domain string
			count  int
		}, 0, len(stats.Domains))

		for domain, count := range stats.Domains {
			domainList = append(domainList, struct {
				domain string
				count  int
			}{domain, count})
		}

		sort.Slice(domainList, func(i, j int) bool {
			return domainList[i].count > domainList[j].count
		})

		if len(domainList) < maxDomains {
			maxDomains = len(domainList)
		}

		for j := 0; j < maxDomains; j++ {
			fmt.Printf("     %s: %d queries\n", domainList[j].domain, domainList[j].count)
		}

		// Query types distribution
		if config.Output.VerboseOutput {
			fmt.Println("   Query Types:")
			for queryType, count := range stats.QueryTypes {
				queryTypeName := getQueryTypeName(queryType)
				fmt.Printf("     %s (%d): %d queries\n", queryTypeName, queryType, count)
			}

			// Status codes distribution
			fmt.Println("   Status Codes:")
			for status, count := range stats.StatusCodes {
				statusName := getStatusName(status)
				fmt.Printf("     %s (%d): %d queries\n", statusName, status, count)
			}
		}
	}

	// Save detailed report to file if configured
	if config.Output.SaveReports {
		saveDetailedReportWithConfig(statsList, totalQueries, config)
	}
}

func displayResults(clientStats map[string]*ClientStats) {
	// Convert to slice for sorting
	var statsList ClientStatsList
	for _, stats := range clientStats {
		// If online-only flag is set, only include clients that are online
		if *onlineOnlyFlag {
			if stats.IsOnline {
				statsList = append(statsList, *stats)
			}
		} else {
			statsList = append(statsList, *stats)
		}
	}

	// Sort by total queries (descending)
	sort.Sort(statsList)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("DNS USAGE ANALYSIS BY CLIENT")
	if *onlineOnlyFlag {
		fmt.Println("(Showing only online clients)")
	}
	fmt.Println(strings.Repeat("=", 80))

	// Summary
	fmt.Printf("Total unique clients: %d\n", len(statsList))
	totalQueries := 0
	for _, stats := range statsList {
		totalQueries += stats.TotalQueries
	}
	fmt.Printf("Total DNS queries: %d\n", totalQueries)
	fmt.Println()

	// Top 20 clients by query count
	fmt.Println("TOP 20 CLIENTS BY QUERY COUNT:")
	fmt.Println(strings.Repeat("-", 110))
	fmt.Printf("%-25s %-20s %-10s %-12s %-12s %-8s %-8s\n", "Client IP", "Hostname", "Queries", "Domains", "Avg Reply", "% Total", "Online")
	fmt.Println(strings.Repeat("-", 110))

	maxDisplay := 20
	if len(statsList) < maxDisplay {
		maxDisplay = len(statsList)
	}

	for i := 0; i < maxDisplay; i++ {
		stats := statsList[i]
		percentage := float64(stats.TotalQueries) / float64(totalQueries) * 100
		clientDisplay := stats.Client
		if stats.HWAddr != "" {
			clientDisplay = fmt.Sprintf("%s (%s)", stats.Client, stats.HWAddr[:12]+"...") // Show first 12 chars of MAC
		}

		// Prepare hostname display
		hostname := stats.Hostname
		if hostname == "" {
			hostname = "-"
		} else if len(hostname) > 18 {
			hostname = hostname[:15] + "..."
		}

		onlineStatus := "Unknown"
		if stats.ARPStatus != "unknown" {
			if stats.IsOnline {
				onlineStatus = "✓ Online"
			} else {
				onlineStatus = "✗ Offline"
			}
		}

		fmt.Printf("%-25s %-20s %-10d %-12d %-12.6f %-8.2f%% %-8s\n",
			clientDisplay,
			hostname,
			stats.TotalQueries,
			stats.Uniquedomains,
			stats.AvgReplyTime,
			percentage,
			onlineStatus)
	}

	// Detailed analysis for top 5 clients
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("DETAILED ANALYSIS - TOP 5 CLIENTS")
	fmt.Println(strings.Repeat("=", 80))

	maxDetailed := 5
	if len(statsList) < maxDetailed {
		maxDetailed = len(statsList)
	}

	for i := 0; i < maxDetailed; i++ {
		stats := statsList[i]
		fmt.Printf("\n%d. CLIENT: %s\n", i+1, stats.Client)
		if stats.Hostname != "" {
			fmt.Printf("   Hostname: %s\n", stats.Hostname)
		}
		if stats.HWAddr != "" {
			fmt.Printf("   Hardware Address: %s\n", stats.HWAddr)
		}

		// Show ARP status
		if stats.ARPStatus != "unknown" {
			statusIcon := "✗"
			if stats.IsOnline {
				statusIcon = "✓"
			}
			fmt.Printf("   Network Status: %s %s\n", statusIcon, strings.Title(stats.ARPStatus))
		}

		fmt.Printf("   Total Queries: %d\n", stats.TotalQueries)
		fmt.Printf("   Unique Domains: %d\n", stats.Uniquedomains)
		fmt.Printf("   Average Reply Time: %.6f seconds\n", stats.AvgReplyTime)

		// Top domains for this client
		fmt.Println("   Top 10 Domains:")
		domainList := make([]struct {
			domain string
			count  int
		}, 0, len(stats.Domains))

		for domain, count := range stats.Domains {
			domainList = append(domainList, struct {
				domain string
				count  int
			}{domain, count})
		}

		sort.Slice(domainList, func(i, j int) bool {
			return domainList[i].count > domainList[j].count
		})

		maxDomains := 10
		if len(domainList) < maxDomains {
			maxDomains = len(domainList)
		}

		for j := 0; j < maxDomains; j++ {
			fmt.Printf("     %s: %d queries\n", domainList[j].domain, domainList[j].count)
		}

		// Query types distribution
		fmt.Println("   Query Types:")
		for queryType, count := range stats.QueryTypes {
			queryTypeName := getQueryTypeName(queryType)
			fmt.Printf("     %s (%d): %d queries\n", queryTypeName, queryType, count)
		}

		// Status codes distribution
		fmt.Println("   Status Codes:")
		for status, count := range stats.StatusCodes {
			statusName := getStatusName(status)
			fmt.Printf("     %s (%d): %d queries\n", statusName, status, count)
		}
	}

	// Save detailed report to file
	saveDetailedReport(statsList, totalQueries)
}

func getQueryTypeName(queryType int) string {
	switch queryType {
	case 1:
		return "A"
	case 2:
		return "NS"
	case 5:
		return "CNAME"
	case 6:
		return "SOA"
	case 12:
		return "PTR"
	case 15:
		return "MX"
	case 16:
		return "TXT"
	case 28:
		return "AAAA"
	case 33:
		return "SRV"
	default:
		return "OTHER"
	}
}

func getStatusName(status int) string {
	switch status {
	case 1:
		return "ALLOWED"
	case 2:
		return "FORWARDED"
	case 3:
		return "CACHED"
	case 4:
		return "BLOCKED_REGEX"
	case 5:
		return "BLOCKED_EXACT"
	case 6:
		return "BLOCKED_WILDCARD"
	case 7:
		return "BLOCKED_GRAVITY"
	case 8:
		return "BLOCKED_EXTERNAL"
	case 9:
		return "BLOCKED_SPECIAL"
	case 10:
		return "BLOCKED_DATABASE"
	case 11:
		return "BLOCKED_DATABASE_ADLIST"
	case 12:
		return "RETRIED"
	case 13:
		return "RETRIED_DNSSEC"
	case 14:
		return "NOT_BLOCKED"
	case 15:
		return "UPSTREAM_ERROR"
	case 16:
		return "GRAVITY_CNAME"
	case 17:
		return "REGEX_CNAME"
	default:
		return "UNKNOWN"
	}
}

func saveDetailedReportWithConfig(statsList ClientStatsList, totalQueries int, config *Config) {
	reportDir := config.Output.ReportDir
	if reportDir == "" {
		reportDir = "."
	}

	// Ensure report directory exists
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		log.Printf("Error creating report directory: %v", err)
		return
	}

	filename := filepath.Join(reportDir, fmt.Sprintf("dns_usage_report_%s.txt", time.Now().Format("20060102_150405")))
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Error creating report file: %v", err)
		return
	}
	defer file.Close()

	fmt.Fprintf(file, "DNS USAGE ANALYSIS REPORT\n")
	fmt.Fprintf(file, "Generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, strings.Repeat("=", 80)+"\n\n")

	// Include configuration info in report
	fmt.Fprintf(file, "CONFIGURATION:\n")
	fmt.Fprintf(file, "Online Only: %t\n", config.OnlineOnly)
	fmt.Fprintf(file, "No Exclude: %t\n", config.NoExclude)
	fmt.Fprintf(file, "Test Mode: %t\n", config.TestMode)
	fmt.Fprintf(file, "Max Clients Display: %d\n", config.Output.MaxClients)
	fmt.Fprintf(file, "Max Domains Display: %d\n", config.Output.MaxDomains)
	fmt.Fprintf(file, "Verbose Output: %t\n\n", config.Output.VerboseOutput)

	fmt.Fprintf(file, "SUMMARY:\n")
	fmt.Fprintf(file, "Total unique clients: %d\n", len(statsList))
	fmt.Fprintf(file, "Total DNS queries: %d\n\n", totalQueries)

	fmt.Fprintf(file, "ALL CLIENTS (sorted by query count):\n")
	fmt.Fprintf(file, "%-20s %-10s %-15s %-15s %-10s %-10s\n", "Client IP", "Queries", "Unique Domains", "Avg Reply Time", "% of Total", "Status")
	fmt.Fprintf(file, strings.Repeat("-", 90)+"\n")

	for _, stats := range statsList {
		percentage := float64(stats.TotalQueries) / float64(totalQueries) * 100
		status := "Offline"
		if stats.IsOnline {
			status = "Online"
		}
		fmt.Fprintf(file, "%-20s %-10d %-15d %-15.6f %-10.2f%% %-10s\n",
			stats.Client,
			stats.TotalQueries,
			stats.Uniquedomains,
			stats.AvgReplyTime,
			percentage,
			status)
	}

	fmt.Printf("\nDetailed report saved to: %s\n", filename)
}

func saveDetailedReport(statsList ClientStatsList, totalQueries int) {
	filename := fmt.Sprintf("dns_usage_report_%s.txt", time.Now().Format("20060102_150405"))
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Error creating report file: %v", err)
		return
	}
	defer file.Close()

	fmt.Fprintf(file, "DNS USAGE ANALYSIS REPORT\n")
	fmt.Fprintf(file, "Generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, strings.Repeat("=", 80)+"\n\n")

	fmt.Fprintf(file, "SUMMARY:\n")
	fmt.Fprintf(file, "Total unique clients: %d\n", len(statsList))
	fmt.Fprintf(file, "Total DNS queries: %d\n\n", totalQueries)

	fmt.Fprintf(file, "ALL CLIENTS (sorted by query count):\n")
	fmt.Fprintf(file, "%-20s %-10s %-15s %-15s %-10s\n", "Client IP", "Queries", "Unique Domains", "Avg Reply Time", "% of Total")
	fmt.Fprintf(file, strings.Repeat("-", 80)+"\n")

	for _, stats := range statsList {
		percentage := float64(stats.TotalQueries) / float64(totalQueries) * 100
		fmt.Fprintf(file, "%-20s %-10d %-15d %-15.6f %-10.2f%%\n",
			stats.Client,
			stats.TotalQueries,
			stats.Uniquedomains,
			stats.AvgReplyTime,
			percentage)
	}

	fmt.Printf("\nDetailed report saved to: %s\n", filename)
}
