package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"pihole-network-analyzer/internal/types"
)

// DNSRecord represents a single DNS query record (local definition for CSV parsing)
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

// ParseRecord parses a CSV record into a DNSRecord
func ParseRecord(record []string) (*DNSRecord, error) {
	if len(record) < 5 {
		return nil, fmt.Errorf("record has insufficient fields: %d", len(record))
	}

	dnsRecord := &DNSRecord{}
	var err error

	// Parse ID
	if dnsRecord.ID, err = strconv.Atoi(record[0]); err != nil {
		return nil, fmt.Errorf("error parsing ID: %v", err)
	}

	// Parse datetime
	dnsRecord.DateTime = record[1]

	// Parse domain
	dnsRecord.Domain = record[2]

	// Parse type
	if dnsRecord.Type, err = strconv.Atoi(record[3]); err != nil {
		return nil, fmt.Errorf("error parsing type: %v", err)
	}

	// Parse status
	if dnsRecord.Status, err = strconv.Atoi(record[4]); err != nil {
		return nil, fmt.Errorf("error parsing status: %v", err)
	}

	// Parse client IP
	if len(record) > 5 {
		dnsRecord.Client = record[5]
	}

	// Parse forward (optional)
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
			dnsRecord.ReplyTime = 0.0
		}
	}

	// DNSSEC (optional)
	if len(record) > 10 && record[10] != "" {
		if dnsRecord.DNSSEC, err = strconv.Atoi(record[10]); err != nil {
			dnsRecord.DNSSEC = 0
		}
	}

	// List ID (optional)
	if len(record) > 11 && record[11] != "" {
		if dnsRecord.ListID, err = strconv.Atoi(record[11]); err != nil {
			dnsRecord.ListID = 0
		}
	}

	// EDE (optional)
	if len(record) > 12 && record[12] != "" {
		if dnsRecord.EDE, err = strconv.Atoi(record[12]); err != nil {
			dnsRecord.EDE = 0
		}
	}

	return dnsRecord, nil
}

// ReadCSVFile reads and parses a CSV file containing DNS records
func ReadCSVFile(filename string) ([]*DNSRecord, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %v", err)
	}

	var dnsRecords []*DNSRecord
	for i, record := range records {
		// Skip header row if it exists
		if i == 0 && isHeaderRow(record) {
			continue
		}

		dnsRecord, err := ParseRecord(record)
		if err != nil {
			// Log error but continue processing
			fmt.Printf("Warning: Error parsing record %d: %v\n", i+1, err)
			continue
		}

		dnsRecords = append(dnsRecords, dnsRecord)
	}

	return dnsRecords, nil
}

// isHeaderRow checks if a record looks like a header row
func isHeaderRow(record []string) bool {
	if len(record) == 0 {
		return false
	}

	// Check if first column contains non-numeric data that looks like a header
	if record[0] == "ID" || record[0] == "id" || record[0] == "timestamp" {
		return true
	}

	// Try to parse first column as integer
	if _, err := strconv.Atoi(record[0]); err != nil {
		return true // Likely a header if it's not a number
	}

	return false
}

// UpdateClientStats updates client statistics with a DNS record
func UpdateClientStats(clientStats map[string]*types.ClientStats, record *DNSRecord) {
	client := record.Client
	if client == "" {
		return
	}

	// Initialize client stats if not exists
	if clientStats[client] == nil {
		clientStats[client] = &types.ClientStats{
			Client:      client,
			Domains:     make(map[string]int),
			QueryTypes:  make(map[int]int),
			StatusCodes: make(map[int]int),
		}
	}

	stats := clientStats[client]
	stats.TotalQueries++
	stats.QueryTypes[record.Type]++
	stats.StatusCodes[record.Status]++

	// Track domain queries
	if record.Domain != "" {
		stats.Domains[record.Domain]++
		if stats.Domains[record.Domain] == 1 {
			stats.UniqueQueries++
		}
	}

	// Update reply time
	if record.ReplyTime > 0 {
		stats.TotalReplyTime += record.ReplyTime
		stats.AvgReplyTime = stats.TotalReplyTime / float64(stats.TotalQueries)
	}
}
