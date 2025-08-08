package reporting

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"pihole-network-analyzer/internal/colors"
	"pihole-network-analyzer/internal/types"
)

// ClientStatsList is a sortable slice of ClientStats
type ClientStatsList []*types.ClientStats

func (c ClientStatsList) Len() int           { return len(c) }
func (c ClientStatsList) Less(i, j int) bool { return c[i].TotalQueries > c[j].TotalQueries }
func (c ClientStatsList) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

// DisplayResults displays the analysis results with default configuration
func DisplayResults(clientStats map[string]*types.ClientStats) {
	DisplayResultsWithConfig(clientStats, nil)
}

// DisplayResultsWithConfig displays the analysis results with specific configuration
func DisplayResultsWithConfig(clientStats map[string]*types.ClientStats, config *types.Config) {
	if len(clientStats) == 0 {
		fmt.Println(colors.Warning("No data to display"))
		return
	}

	// Convert map to slice and sort by query count
	var statsList ClientStatsList
	for _, stats := range clientStats {
		statsList = append(statsList, stats)
	}
	sort.Sort(statsList)

	// Apply configuration filters
	maxClients := 50 // default
	onlineOnly := false
	if config != nil {
		if config.Output.MaxClients > 0 {
			maxClients = config.Output.MaxClients
		}
		onlineOnly = config.OnlineOnly
	}

	// Filter online-only if requested
	if onlineOnly {
		var filteredList ClientStatsList
		for _, stats := range statsList {
			if stats.IsOnline {
				filteredList = append(filteredList, stats)
			}
		}
		statsList = filteredList
	}

	// Limit number of clients displayed
	if len(statsList) > maxClients {
		statsList = statsList[:maxClients]
	}

	// Calculate totals
	totalQueries := 0
	totalClients := len(clientStats)
	onlineClients := 0
	for _, stats := range clientStats {
		totalQueries += stats.TotalQueries
		if stats.IsOnline {
			onlineClients++
		}
	}

	// Display summary header
	fmt.Println(colors.SectionHeader("DNS Query Analysis Summary"))
	fmt.Printf("Total Clients: %s | Online: %s | Total Queries: %s\n\n",
		colors.BoldWhite(fmt.Sprintf("%d", totalClients)),
		colors.BoldGreen(fmt.Sprintf("%d", onlineClients)),
		colors.BoldCyan(fmt.Sprintf("%d", totalQueries)))

	// Display client table
	displayClientTable(statsList, config)

	// Save detailed report if requested
	if config != nil && config.Output.SaveReports {
		SaveDetailedReportWithConfig(statsList, totalQueries, config)
	}
}

// displayClientTable displays the main client statistics table
func displayClientTable(statsList ClientStatsList, config *types.Config) {
	if len(statsList) == 0 {
		return
	}

	// Table header
	fmt.Println(colors.SubSectionHeader("Client Statistics"))
	fmt.Printf("%-16s %-18s %-18s %-10s %-10s %-12s %-8s %-8s %s\n",
		"IP Address", "Hostname", "MAC Address", "Status", "Queries", "Unique", "Avg RT", "Top Domain", "Query %")

	// Table rows
	for i, stats := range statsList {
		if i >= 100 { // Reasonable limit for display
			break
		}

		displayClientRow(stats, config)
	}
}

// displayClientRow displays a single client statistics row
func displayClientRow(stats *types.ClientStats, config *types.Config) {
	// Format IP address
	ip := colors.HighlightIP(stats.Client)

	// Format hostname
	hostname := stats.Hostname
	if hostname == "" {
		hostname = "unknown"
	}
	if len(hostname) > 16 {
		hostname = hostname[:13] + "..."
	}

	// Format MAC address
	macAddr := stats.HWAddr
	if macAddr == "" {
		macAddr = "unknown"
	}
	if len(macAddr) > 16 {
		macAddr = macAddr[:13] + "..."
	}

	// Format online status
	status := colors.OnlineStatus(stats.IsOnline, stats.ARPStatus)

	// Format query counts
	totalQueries := colors.ColoredQueryCount(stats.TotalQueries)
	uniqueQueries := colors.ColoredDomainCount(stats.UniqueQueries)

	// Format average reply time
	avgReplyTime := fmt.Sprintf("%.3fms", stats.AvgReplyTime*1000)
	if stats.AvgReplyTime == 0 {
		avgReplyTime = "N/A"
	}

	// Get top domain
	topDomain := getTopDomain(stats)
	if len(topDomain) > 15 {
		topDomain = topDomain[:12] + "..."
	}
	topDomain = colors.HighlightDomain(topDomain)

	// Calculate query percentage
	// This would need total queries from all clients to be accurate
	queryPercentage := "N/A"

	fmt.Printf("%-16s %-18s %-18s %-10s %-10s %-12s %-8s %-8s %s\n",
		ip, hostname, macAddr, status, totalQueries, uniqueQueries,
		avgReplyTime, topDomain, queryPercentage)

	// Show additional details if verbose mode
	if config != nil && config.Output.Verbose {
		displayVerboseClientDetails(stats)
	}
}

// getTopDomain returns the most queried domain for a client
func getTopDomain(stats *types.ClientStats) string {
	if len(stats.Domains) == 0 {
		return "none"
	}

	var topDomain string
	var maxCount int

	for domain, count := range stats.Domains {
		if count > maxCount {
			maxCount = count
			topDomain = domain
		}
	}

	return topDomain
}

// displayVerboseClientDetails shows additional client information
func displayVerboseClientDetails(stats *types.ClientStats) {
	fmt.Printf("   Query Types:")
	for queryType, count := range stats.QueryTypes {
		if count > 0 {
			fmt.Printf(" %s:%d", getQueryTypeName(queryType), count)
		}
	}
	fmt.Println()

	fmt.Printf("   Status Codes:")
	for status, count := range stats.StatusCodes {
		if count > 0 {
			fmt.Printf(" %s:%d", getStatusName(status), count)
		}
	}
	fmt.Println()
}

// getQueryTypeName returns human-readable query type name
func getQueryTypeName(queryType int) string {
	switch queryType {
	case 1:
		return "A"
	case 28:
		return "AAAA"
	case 2:
		return "NS"
	case 5:
		return "CNAME"
	case 15:
		return "MX"
	case 16:
		return "TXT"
	default:
		return fmt.Sprintf("Type%d", queryType)
	}
}

// getStatusName returns human-readable status name
func getStatusName(status int) string {
	switch status {
	case 0:
		return "Unknown"
	case 1:
		return "Blocked"
	case 2:
		return "Forwarded"
	case 3:
		return "Cached"
	case 4:
		return "Regex"
	default:
		return fmt.Sprintf("Status%d", status)
	}
}

// SaveDetailedReportWithConfig saves a detailed report to file
func SaveDetailedReportWithConfig(statsList ClientStatsList, totalQueries int, config *types.Config) {
	if config == nil {
		return
	}

	reportDir := config.Output.ReportDir
	if reportDir == "" {
		reportDir = "reports"
	}

	// Create reports directory if it doesn't exist
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		fmt.Printf("Error creating reports directory: %v\n", err)
		return
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s/dns_analysis_%s.txt", reportDir, timestamp)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating report file: %v\n", err)
		return
	}
	defer file.Close()

	// Write report header
	fmt.Fprintf(file, "DNS Query Analysis Report\n")
	fmt.Fprintf(file, "Generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "Total Queries: %d\n", totalQueries)
	fmt.Fprintf(file, "Total Clients: %d\n", len(statsList))
	fmt.Fprintf(file, "\n%s\n", strings.Repeat("=", 80))

	// Write client statistics
	for i, stats := range statsList {
		fmt.Fprintf(file, "\n%d. Client: %s\n", i+1, stats.Client)
		fmt.Fprintf(file, "   Hostname: %s\n", stats.Hostname)
		fmt.Fprintf(file, "   MAC Address: %s\n", stats.HWAddr)
		fmt.Fprintf(file, "   Online: %t\n", stats.IsOnline)
		fmt.Fprintf(file, "   Total Queries: %d\n", stats.TotalQueries)
		fmt.Fprintf(file, "   Unique Domains: %d\n", stats.UniqueQueries)
		fmt.Fprintf(file, "   Average Reply Time: %.3fms\n", stats.AvgReplyTime*1000)

		// Top domains
		fmt.Fprintf(file, "   Top Domains:\n")
		topDomains := getTopDomains(stats, 10)
		for j, domain := range topDomains {
			fmt.Fprintf(file, "     %d. %s (%d queries)\n", j+1, domain.Domain, domain.Count)
		}
	}

	fmt.Printf("Detailed report saved to: %s\n", filename)
}

// DomainCount represents a domain and its query count
type DomainCount struct {
	Domain string
	Count  int
}

// getTopDomains returns the top domains for a client, sorted by query count
func getTopDomains(stats *types.ClientStats, limit int) []DomainCount {
	var domains []DomainCount
	for domain, count := range stats.Domains {
		domains = append(domains, DomainCount{Domain: domain, Count: count})
	}

	// Sort by count (descending)
	sort.Slice(domains, func(i, j int) bool {
		return domains[i].Count > domains[j].Count
	})

	// Limit results
	if len(domains) > limit {
		domains = domains[:limit]
	}

	return domains
}

// QuietPrintf prints formatted output only if not in quiet mode
func QuietPrintf(config *types.Config, format string, args ...interface{}) {
	if config != nil && config.Quiet {
		return
	}
	fmt.Printf(format, args...)
}
