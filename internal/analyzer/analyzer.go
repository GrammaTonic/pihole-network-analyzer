package analyzer

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"pihole-network-analyzer/internal/colors"
	csvpkg "pihole-network-analyzer/internal/csv"
	"pihole-network-analyzer/internal/network"
	"pihole-network-analyzer/internal/types"
)

// AnalyzeDNSData performs DNS data analysis from a CSV file
func AnalyzeDNSData(filename string) (map[string]*types.ClientStats, error) {
	// Create default configuration if none provided
	defaultConfig := &types.Config{
		Exclusions: *network.GetDefaultExclusions(""),
	}
	return AnalyzeDNSDataWithConfig(filename, defaultConfig)
}

// AnalyzeDNSDataWithConfig performs DNS data analysis with specific configuration
func AnalyzeDNSDataWithConfig(filename string, config *types.Config) (map[string]*types.ClientStats, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	clientStats := make(map[string]*types.ClientStats)
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

	if !config.Quiet {
		fmt.Println(colors.ProcessingIndicator("Processing CSV records with exclusions..."))
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading record %d: %v", recordCount, err)
			continue
		}

		recordCount++
		if recordCount%10000 == 0 && !config.Quiet {
			fmt.Printf("Processed %d records...\n", recordCount)
		}

		// Parse record using CSV package
		dnsRecord, err := csvpkg.ParseRecord(record)
		if err != nil {
			continue
		}

		// Check if client should be excluded
		client := dnsRecord.Client
		if client == "" {
			continue
		}

		// Check exclusions (only once per IP to improve performance)
		if !excludedIPs[client] {
			if shouldExclude, reason := network.ShouldExcludeClient(client, "", exclusions); shouldExclude {
				excludedIPs[client] = true
				excludedCount++
				if !config.Quiet {
					fmt.Printf("Excluding client %s: %s\n", client, reason)
				}
				continue
			}
			excludedIPs[client] = false
		} else if excludedIPs[client] {
			continue
		}

		// Update client statistics using CSV package
		csvpkg.UpdateClientStats(clientStats, dnsRecord)
	}

	if !config.Quiet {
		fmt.Printf("Processed %d records, excluded %d clients\n", recordCount, excludedCount)
		fmt.Println(colors.ProcessingIndicator("Checking ARP status and resolving hostnames..."))
	}

	// Resolve hostnames and check ARP status
	network.ResolveHostnames(clientStats)
	if err := network.CheckARPStatus(clientStats); err != nil {
		log.Printf("Warning: Could not check ARP status: %v", err)
	}

	return clientStats, nil
}

// GetQueryTypeName returns human-readable query type name
func GetQueryTypeName(queryType int) string {
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
	case 35:
		return "NAPTR"
	case 39:
		return "DNAME"
	case 41:
		return "OPT"
	case 43:
		return "DS"
	case 46:
		return "RRSIG"
	case 47:
		return "NSEC"
	case 48:
		return "DNSKEY"
	case 50:
		return "NSEC3"
	case 51:
		return "NSEC3PARAM"
	case 52:
		return "TLSA"
	case 257:
		return "CAA"
	default:
		return "Unknown"
	}
}

// GetStatusName returns human-readable status name
func GetStatusName(status int) string {
	switch status {
	case 0:
		return "Unknown"
	case 1:
		return "Blocked (gravity)"
	case 2:
		return "Forwarded"
	case 3:
		return "Cached"
	case 4:
		return "Blocked (regex/wildcard)"
	case 5:
		return "Blocked (exact)"
	case 6:
		return "Blocked (external)"
	case 7:
		return "CNAME"
	case 8:
		return "Retried"
	case 9:
		return "Retried but ignored"
	case 10:
		return "Already forwarded"
	case 11:
		return "Already cached"
	case 12:
		return "Config blocked"
	case 13:
		return "Gravity blocked"
	case 14:
		return "Regex blocked"
	default:
		return "Unknown"
	}
}

// GetStatusCodeFromString converts status strings to status codes
func GetStatusCodeFromString(status string) int {
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
