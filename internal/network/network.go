package network

import (
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"pihole-analyzer/internal/types"
)

// ARPEntry represents an entry from the ARP table (local definition)
type ARPEntry struct {
	IP       string
	HWAddr   string
	Status   string
	IsOnline bool
}

// GetLocalARPTable retrieves the local ARP table
func GetLocalARPTable() (map[string]*ARPEntry, error) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("arp", "-a")
	case "darwin":
		cmd = exec.Command("arp", "-a")
	default: // Linux and others
		cmd = exec.Command("arp", "-a")
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseARPOutput(string(output))
}

// parseARPOutput parses ARP command output
func parseARPOutput(output string) (map[string]*ARPEntry, error) {
	arpTable := make(map[string]*ARPEntry)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		entry := parseARPLine(line)
		if entry != nil && entry.IP != "" {
			arpTable[entry.IP] = entry
		}
	}

	return arpTable, nil
}

// parseARPLine parses a single line from ARP output
func parseARPLine(line string) *ARPEntry {
	// Linux/Unix format: hostname (192.168.1.1) at aa:bb:cc:dd:ee:ff [ether] on eth0
	// Windows format: 192.168.1.1          aa-bb-cc-dd-ee-ff     dynamic
	// macOS format: hostname (192.168.1.1) at aa:bb:cc:dd:ee:ff on en0 ifscope [ethernet]

	var ip, hwAddr, status string
	var isOnline bool

	switch runtime.GOOS {
	case "windows":
		// Windows format
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			ip = fields[0]
			hwAddr = fields[1]
			status = strings.Join(fields[2:], " ")
			isOnline = strings.Contains(status, "dynamic")
		}
	default:
		// Unix-like systems (Linux, macOS)
		re := regexp.MustCompile(`\(([^)]+)\)\s+at\s+([0-9a-fA-F:]{17})`)
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 3 {
			ip = matches[1]
			hwAddr = matches[2]
			status = "reachable"
			isOnline = true
		}
	}

	if ip == "" || hwAddr == "" {
		return nil
	}

	// Normalize MAC address format
	hwAddr = normalizeMACAddress(hwAddr)

	return &ARPEntry{
		IP:       ip,
		HWAddr:   hwAddr,
		Status:   status,
		IsOnline: isOnline,
	}
}

// normalizeMACAddress converts MAC address to standard format
func normalizeMACAddress(mac string) string {
	// Convert Windows format (aa-bb-cc-dd-ee-ff) to Unix format (aa:bb:cc:dd:ee:ff)
	mac = strings.ReplaceAll(mac, "-", ":")
	return strings.ToLower(mac)
}

// PingHost pings a host to check if it's reachable
func PingHost(ip string) bool {
	timeout := time.Second * 3

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("ping", "-n", "1", "-w", "3000", ip)
	case "darwin":
		cmd = exec.Command("ping", "-c", "1", "-W", "3000", ip)
	default: // Linux
		cmd = exec.Command("ping", "-c", "1", "-W", "3", ip)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		return err == nil
	case <-time.After(timeout):
		return false
	}
}

// ARPPing performs an ARP ping to check if a host is reachable
func ARPPing(ip string) bool {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("arping", "-c", "1", "-W", "1", ip)
	case "darwin":
		cmd = exec.Command("ping", "-c", "1", ip) // fallback to regular ping on macOS
	default:
		return PingHost(ip) // fallback for other systems
	}

	err := cmd.Run()
	return err == nil
}

// RefreshARPTable refreshes the ARP table by pinging client IPs
func RefreshARPTable(clientIPs []string) {
	maxConcurrent := 50
	semaphore := make(chan struct{}, maxConcurrent)

	for _, ip := range clientIPs {
		if !IsValidIPv4(ip) {
			continue
		}

		go func(ip string) {
			semaphore <- struct{}{}        // acquire
			defer func() { <-semaphore }() // release

			PingHost(ip)
		}(ip)
	}

	// Wait a bit for pings to complete
	time.Sleep(time.Second * 2)
}

// IsValidIPv4 checks if a string is a valid IPv4 address
func IsValidIPv4(ip string) bool {
	if ip == "" {
		return false
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Check if it's an IPv4 address
	return parsedIP.To4() != nil
}

// ResolveHostname resolves an IP address to a hostname
func ResolveHostname(ip string) string {
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return ""
	}

	hostname := names[0]
	// Remove trailing dot if present
	if strings.HasSuffix(hostname, ".") {
		hostname = hostname[:len(hostname)-1]
	}

	return hostname
}

// ResolveHostnames resolves hostnames for all client IPs in parallel
func ResolveHostnames(clientStats map[string]*types.ClientStats) {
	maxConcurrent := 20
	semaphore := make(chan struct{}, maxConcurrent)

	for ip, stats := range clientStats {
		if !IsValidIPv4(ip) {
			continue
		}

		go func(ip string, stats *types.ClientStats) {
			semaphore <- struct{}{}        // acquire
			defer func() { <-semaphore }() // release

			hostname := ResolveHostname(ip)
			if hostname != "" {
				stats.Hostname = hostname
			}
		}(ip, stats)
	}

	// Wait for all hostname resolutions to complete
	time.Sleep(time.Second * 5)
}

// CheckARPStatus updates client statistics with ARP status information
func CheckARPStatus(clientStats map[string]*types.ClientStats) error {
	arpTable, err := GetLocalARPTable()
	if err != nil {
		return err
	}

	for ip, stats := range clientStats {
		if arpEntry, exists := arpTable[ip]; exists {
			stats.IsOnline = arpEntry.IsOnline
			stats.HWAddr = arpEntry.HWAddr
			stats.ARPStatus = arpEntry.Status
		} else {
			stats.IsOnline = false
			stats.ARPStatus = "not in ARP table"
		}
	}

	return nil
}

// IsIPInNetwork checks if an IP address is in a CIDR network
func IsIPInNetwork(ip, cidr string) bool {
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

// ShouldExcludeClient determines if a client should be excluded based on exclusion rules
func ShouldExcludeClient(ip, hostname string, exclusions *types.ExclusionConfig) (bool, string) {
	if exclusions == nil {
		return false, ""
	}

	// Check IP exclusions
	for _, excludeIP := range exclusions.ExcludeIPs {
		if ip == excludeIP {
			return true, "excluded IP"
		}
	}

	// Check network exclusions
	for _, network := range exclusions.ExcludeNetworks {
		if IsIPInNetwork(ip, network) {
			return true, "excluded network: " + network
		}
	}

	// Check hostname exclusions
	for _, excludeHost := range exclusions.ExcludeHosts {
		if strings.Contains(hostname, excludeHost) {
			return true, "excluded hostname pattern: " + excludeHost
		}
	}

	return false, ""
}

// GetDefaultExclusions returns default exclusion configuration
func GetDefaultExclusions(piholeHost string) *types.ExclusionConfig {
	exclusions := &types.ExclusionConfig{
		ExcludeNetworks: []string{
			"172.16.0.0/12",  // Docker networks
			"172.17.0.0/16",  // Docker default bridge
			"172.18.0.0/16",  // Docker custom networks
			"172.19.0.0/16",  // Docker custom networks
			"172.20.0.0/16",  // Docker custom networks
			"127.0.0.0/8",    // Loopback
			"169.254.0.0/16", // Link-local
		},
		ExcludeIPs: []string{
			"0.0.0.0",
			"255.255.255.255",
		},
		ExcludeHosts: []string{
			"localhost",
			"pi.hole",
		},
	}

	// Add Pi-hole host IP to exclusions if provided
	if piholeHost != "" && IsValidIPv4(piholeHost) {
		exclusions.ExcludeIPs = append(exclusions.ExcludeIPs, piholeHost)
	}

	return exclusions
}
