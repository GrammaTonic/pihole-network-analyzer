package network

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strings"
	"time"

	"pihole-analyzer/internal/types"
)

// DefaultDeepPacketInspector implements the DeepPacketInspector interface
type DefaultDeepPacketInspector struct {
	logger *slog.Logger
}

// NewDeepPacketInspector creates a new deep packet inspector
func NewDeepPacketInspector(logger *slog.Logger) DeepPacketInspector {
	return &DefaultDeepPacketInspector{
		logger: logger,
	}
}

// InspectPackets implements DeepPacketInspector.InspectPackets
func (d *DefaultDeepPacketInspector) InspectPackets(ctx context.Context, records []types.PiholeRecord, config types.DPIConfig) (*types.PacketAnalysisResult, error) {
	d.logger.Info("Starting deep packet inspection",
		slog.Int("record_count", len(records)),
		slog.Float64("sampling_rate", config.PacketSampling),
		slog.Int("max_packet_size", config.MaxPacketSize))

	startTime := time.Now()

	// Apply sampling if configured
	sampledRecords := d.applySampling(records, config.PacketSampling)

	result := &types.PacketAnalysisResult{
		TotalPackets:           int64(len(records)),
		AnalyzedPackets:        int64(len(sampledRecords)),
		ProtocolDistribution:   make(map[string]int64),
		PacketSizeDistribution: make(map[string]int64),
		TopSourceIPs:           make([]types.IPTrafficStat, 0),
		TopDestinationIPs:      make([]types.IPTrafficStat, 0),
		PortUsage:              make(map[string]int64),
		Anomalies:              make([]types.PacketAnomaly, 0),
	}

	// Analyze protocol distribution
	protocolDist, err := d.AnalyzeProtocols(sampledRecords)
	if err != nil {
		d.logger.Error("Failed to analyze protocols", slog.String("error", err.Error()))
	} else {
		result.ProtocolDistribution = protocolDist
	}

	// Analyze packet size distribution
	result.PacketSizeDistribution = d.analyzePacketSizes(sampledRecords)

	// Get traffic topology
	sourceIPs, destIPs, err := d.GetTrafficTopology(sampledRecords)
	if err != nil {
		d.logger.Error("Failed to analyze traffic topology", slog.String("error", err.Error()))
	} else {
		result.TopSourceIPs = sourceIPs
		result.TopDestinationIPs = destIPs
	}

	// Analyze port usage
	result.PortUsage = d.analyzePortUsage(sampledRecords)

	// Detect packet anomalies
	baseline := d.createBaseline(records)
	anomalies, err := d.DetectPacketAnomalies(ctx, sampledRecords, baseline)
	if err != nil {
		d.logger.Error("Failed to detect packet anomalies", slog.String("error", err.Error()))
	} else {
		result.Anomalies = anomalies
	}

	duration := time.Since(startTime)
	d.logger.Info("Deep packet inspection completed",
		slog.String("duration", duration.String()),
		slog.Int64("analyzed_packets", result.AnalyzedPackets),
		slog.Int("protocols_detected", len(result.ProtocolDistribution)),
		slog.Int("anomalies_detected", len(result.Anomalies)))

	return result, nil
}

// AnalyzeProtocols implements DeepPacketInspector.AnalyzeProtocols
func (d *DefaultDeepPacketInspector) AnalyzeProtocols(records []types.PiholeRecord) (map[string]int64, error) {
	protocolCounts := make(map[string]int64)

	for _, record := range records {
		protocol := d.inferProtocol(record)
		protocolCounts[protocol]++
	}

	d.logger.Debug("Protocol analysis completed",
		slog.Int("protocol_types", len(protocolCounts)),
		slog.Int("records_analyzed", len(records)))

	return protocolCounts, nil
}

// DetectPacketAnomalies implements DeepPacketInspector.DetectPacketAnomalies
func (d *DefaultDeepPacketInspector) DetectPacketAnomalies(ctx context.Context, records []types.PiholeRecord, baseline map[string]interface{}) ([]types.PacketAnomaly, error) {
	anomalies := make([]types.PacketAnomaly, 0)

	// Extract baseline statistics
	avgQueriesPerMinute, _ := baseline["avg_queries_per_minute"].(float64)
	commonDomains, _ := baseline["common_domains"].(map[string]int)
	typicalClients, _ := baseline["typical_clients"].(map[string]int)

	// Analyze queries by time windows
	timeWindows := d.groupByTimeWindows(records, time.Minute*5)

	for timeSlot, windowRecords := range timeWindows {
		// Check for volume spikes
		queryCount := float64(len(windowRecords))
		if queryCount > avgQueriesPerMinute*3 { // 3x threshold
			anomaly := types.PacketAnomaly{
				ID:          fmt.Sprintf("volume_spike_%s", timeSlot),
				Type:        "volume_spike",
				Description: fmt.Sprintf("Query volume spike detected: %.0f queries (baseline: %.0f)", queryCount, avgQueriesPerMinute),
				Severity:    d.calculateSeverity(queryCount / avgQueriesPerMinute),
				Timestamp:   timeSlot,
				Confidence:  d.calculateConfidence(queryCount/avgQueriesPerMinute, 3.0),
			}
			anomalies = append(anomalies, anomaly)
		}

		// Check for unusual domains
		domainCounts := make(map[string]int)
		for _, record := range windowRecords {
			domainCounts[record.Domain]++
		}

		for domain, count := range domainCounts {
			if _, isCommon := commonDomains[domain]; !isCommon && count > 10 {
				anomaly := types.PacketAnomaly{
					ID:          fmt.Sprintf("unusual_domain_%s_%s", strings.ReplaceAll(domain, ".", "_"), timeSlot),
					Type:        "unusual_domain",
					Description: fmt.Sprintf("Unusual domain activity: %s (%d queries)", domain, count),
					Severity:    "MEDIUM",
					Timestamp:   timeSlot,
					Confidence:  0.75,
				}
				anomalies = append(anomalies, anomaly)
			}
		}

		// Check for unusual client behavior
		clientCounts := make(map[string]int)
		for _, record := range windowRecords {
			clientCounts[record.Client]++
		}

		for client, count := range clientCounts {
			typicalCount, hasBaseline := typicalClients[client]
			if hasBaseline && count > typicalCount*5 { // 5x typical usage
				anomaly := types.PacketAnomaly{
					ID:          fmt.Sprintf("client_spike_%s_%s", strings.ReplaceAll(client, ".", "_"), timeSlot),
					Type:        "client_behavior",
					Description: fmt.Sprintf("Unusual client activity: %s (%d queries, typical: %d)", client, count, typicalCount),
					Severity:    "MEDIUM",
					Timestamp:   timeSlot,
					SourceIP:    client,
					Confidence:  0.8,
				}
				anomalies = append(anomalies, anomaly)
			}
		}
	}

	// Detect potential DNS tunneling
	tunnelingAnomalies := d.detectDNSTunneling(records)
	anomalies = append(anomalies, tunnelingAnomalies...)

	// Detect port scanning patterns
	portScanAnomalies := d.detectPortScanning(records)
	anomalies = append(anomalies, portScanAnomalies...)

	d.logger.Info("Packet anomaly detection completed",
		slog.Int("anomalies_detected", len(anomalies)))

	return anomalies, nil
}

// GetTrafficTopology implements DeepPacketInspector.GetTrafficTopology
func (d *DefaultDeepPacketInspector) GetTrafficTopology(records []types.PiholeRecord) ([]types.IPTrafficStat, []types.IPTrafficStat, error) {
	sourceStats := make(map[string]*types.IPTrafficStat)
	destStats := make(map[string]*types.IPTrafficStat)

	for _, record := range records {
		// Source IP (client)
		if stat, exists := sourceStats[record.Client]; exists {
			stat.PacketCount++
			stat.ByteCount += d.estimatePacketSize(record)
		} else {
			sourceStats[record.Client] = &types.IPTrafficStat{
				IP:          record.Client,
				Hostname:    "", // Will be resolved later if needed
				PacketCount: 1,
				ByteCount:   d.estimatePacketSize(record),
			}
		}

		// Destination IP (inferred from domain)
		destIP := d.inferDestinationIP(record)
		if destIP != "" {
			if stat, exists := destStats[destIP]; exists {
				stat.PacketCount++
				stat.ByteCount += d.estimatePacketSize(record)
			} else {
				destStats[destIP] = &types.IPTrafficStat{
					IP:          destIP,
					PacketCount: 1,
					ByteCount:   d.estimatePacketSize(record),
				}
			}
		}
	}

	// Convert to slices and sort by packet count
	sources := make([]types.IPTrafficStat, 0, len(sourceStats))
	destinations := make([]types.IPTrafficStat, 0, len(destStats))

	totalSourcePackets := int64(0)
	for _, stat := range sourceStats {
		sources = append(sources, *stat)
		totalSourcePackets += stat.PacketCount
	}

	totalDestPackets := int64(0)
	for _, stat := range destStats {
		destinations = append(destinations, *stat)
		totalDestPackets += stat.PacketCount
	}

	// Calculate percentages
	for i := range sources {
		sources[i].Percentage = float64(sources[i].PacketCount) / float64(totalSourcePackets) * 100
	}

	for i := range destinations {
		destinations[i].Percentage = float64(destinations[i].PacketCount) / float64(totalDestPackets) * 100
	}

	// Sort by packet count
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].PacketCount > sources[j].PacketCount
	})

	sort.Slice(destinations, func(i, j int) bool {
		return destinations[i].PacketCount > destinations[j].PacketCount
	})

	// Return top 20 of each
	maxResults := 20
	if len(sources) > maxResults {
		sources = sources[:maxResults]
	}
	if len(destinations) > maxResults {
		destinations = destinations[:maxResults]
	}

	return sources, destinations, nil
}

// Helper methods

// applySampling applies packet sampling based on the configured rate
func (d *DefaultDeepPacketInspector) applySampling(records []types.PiholeRecord, samplingRate float64) []types.PiholeRecord {
	if samplingRate >= 1.0 {
		return records
	}

	if samplingRate <= 0.0 {
		return []types.PiholeRecord{}
	}

	sampled := make([]types.PiholeRecord, 0, int(float64(len(records))*samplingRate))
	step := int(1.0 / samplingRate)

	for i := 0; i < len(records); i += step {
		sampled = append(sampled, records[i])
	}

	return sampled
}

// inferProtocol infers the protocol from a DNS record
func (d *DefaultDeepPacketInspector) inferProtocol(record types.PiholeRecord) string {
	// DNS queries are typically UDP, but some use TCP
	queryType := strings.ToUpper(record.QueryType)

	switch queryType {
	case "A", "AAAA", "PTR", "CNAME", "MX", "TXT", "NS", "SOA":
		return "DNS_UDP"
	case "AXFR", "IXFR":
		return "DNS_TCP"
	default:
		if len(record.Domain) > 100 { // Large queries might use TCP
			return "DNS_TCP"
		}
		return "DNS_UDP"
	}
}

// analyzePacketSizes analyzes the distribution of packet sizes
func (d *DefaultDeepPacketInspector) analyzePacketSizes(records []types.PiholeRecord) map[string]int64 {
	sizeDistribution := make(map[string]int64)

	for _, record := range records {
		size := d.estimatePacketSize(record)
		category := d.categorizePacketSize(size)
		sizeDistribution[category]++
	}

	return sizeDistribution
}

// estimatePacketSize estimates the packet size based on the DNS record
func (d *DefaultDeepPacketInspector) estimatePacketSize(record types.PiholeRecord) int64 {
	// Base DNS header: 12 bytes
	// Question section: domain name + 4 bytes (type + class)
	// Estimated response size varies by type

	baseSize := int64(12)                       // DNS header
	domainSize := int64(len(record.Domain) + 1) // +1 for length encoding
	questionSize := domainSize + 4

	estimatedResponseSize := int64(0)
	switch strings.ToUpper(record.QueryType) {
	case "A":
		estimatedResponseSize = 16 // Answer section for A record
	case "AAAA":
		estimatedResponseSize = 28 // Answer section for AAAA record
	case "CNAME", "PTR":
		estimatedResponseSize = domainSize + 10
	case "MX":
		estimatedResponseSize = domainSize + 14
	case "TXT":
		estimatedResponseSize = domainSize + 20
	default:
		estimatedResponseSize = 20
	}

	return baseSize + questionSize + estimatedResponseSize
}

// categorizePacketSize categorizes packet sizes into buckets
func (d *DefaultDeepPacketInspector) categorizePacketSize(size int64) string {
	switch {
	case size <= 64:
		return "small (≤64 bytes)"
	case size <= 256:
		return "medium (65-256 bytes)"
	case size <= 512:
		return "large (257-512 bytes)"
	case size <= 1024:
		return "jumbo (513-1024 bytes)"
	default:
		return "oversized (>1024 bytes)"
	}
}

// analyzePortUsage analyzes port usage patterns
func (d *DefaultDeepPacketInspector) analyzePortUsage(records []types.PiholeRecord) map[string]int64 {
	portUsage := make(map[string]int64)

	for _, record := range records {
		// DNS typically uses port 53
		port := "53"
		protocol := d.inferProtocol(record)

		portKey := fmt.Sprintf("%s/%s", port, protocol)
		portUsage[portKey]++
	}

	return portUsage
}

// createBaseline creates a baseline for anomaly detection
func (d *DefaultDeepPacketInspector) createBaseline(records []types.PiholeRecord) map[string]interface{} {
	baseline := make(map[string]interface{})

	// Calculate average queries per minute
	if len(records) > 0 {
		timeSpan := d.calculateTimeSpan(records)
		avgQueriesPerMinute := float64(len(records)) / timeSpan.Minutes()
		baseline["avg_queries_per_minute"] = avgQueriesPerMinute
	}

	// Identify common domains (top 10%)
	domainCounts := make(map[string]int)
	for _, record := range records {
		domainCounts[record.Domain]++
	}
	commonDomains := d.getTopDomains(domainCounts, len(domainCounts)/10)
	baseline["common_domains"] = commonDomains

	// Identify typical client usage
	clientCounts := make(map[string]int)
	for _, record := range records {
		clientCounts[record.Client]++
	}
	baseline["typical_clients"] = clientCounts

	return baseline
}

// calculateTimeSpan calculates the time span of records
func (d *DefaultDeepPacketInspector) calculateTimeSpan(records []types.PiholeRecord) time.Duration {
	if len(records) == 0 {
		return time.Hour // Default to 1 hour
	}

	var earliest, latest time.Time
	for i, record := range records {
		timestamp := parseTimestamp(record.Timestamp)
		if i == 0 {
			earliest = timestamp
			latest = timestamp
		} else {
			if timestamp.Before(earliest) {
				earliest = timestamp
			}
			if timestamp.After(latest) {
				latest = timestamp
			}
		}
	}

	duration := latest.Sub(earliest)
	if duration < time.Minute {
		duration = time.Hour // Minimum duration
	}

	return duration
}

// getTopDomains returns the top N domains by count
func (d *DefaultDeepPacketInspector) getTopDomains(domainCounts map[string]int, n int) map[string]int {
	type domainCount struct {
		domain string
		count  int
	}

	domains := make([]domainCount, 0, len(domainCounts))
	for domain, count := range domainCounts {
		domains = append(domains, domainCount{domain, count})
	}

	sort.Slice(domains, func(i, j int) bool {
		return domains[i].count > domains[j].count
	})

	if n > len(domains) {
		n = len(domains)
	}

	result := make(map[string]int)
	for i := 0; i < n; i++ {
		result[domains[i].domain] = domains[i].count
	}

	return result
}

// groupByTimeWindows groups records by time windows
func (d *DefaultDeepPacketInspector) groupByTimeWindows(records []types.PiholeRecord, windowSize time.Duration) map[string][]types.PiholeRecord {
	windows := make(map[string][]types.PiholeRecord)

	for _, record := range records {
		timestamp := parseTimestamp(record.Timestamp)
		windowKey := timestamp.Truncate(windowSize).Format(time.RFC3339)

		if _, exists := windows[windowKey]; !exists {
			windows[windowKey] = make([]types.PiholeRecord, 0)
		}
		windows[windowKey] = append(windows[windowKey], record)
	}

	return windows
}

// calculateSeverity determines the severity based on the deviation magnitude
func (d *DefaultDeepPacketInspector) calculateSeverity(deviation float64) string {
	switch {
	case deviation >= 10:
		return "CRITICAL"
	case deviation >= 5:
		return "HIGH"
	case deviation >= 3:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

// calculateConfidence calculates confidence based on deviation and threshold
func (d *DefaultDeepPacketInspector) calculateConfidence(deviation, threshold float64) float64 {
	if deviation < threshold {
		return 0.0
	}

	confidence := math.Min(0.95, 0.5+(deviation-threshold)*0.1)
	return normalizeScore(confidence)
}

// detectDNSTunneling detects potential DNS tunneling attempts
func (d *DefaultDeepPacketInspector) detectDNSTunneling(records []types.PiholeRecord) []types.PacketAnomaly {
	anomalies := make([]types.PacketAnomaly, 0)

	// Group queries by domain and client
	domainClientQueries := make(map[string]map[string]int)

	for _, record := range records {
		if _, exists := domainClientQueries[record.Domain]; !exists {
			domainClientQueries[record.Domain] = make(map[string]int)
		}
		domainClientQueries[record.Domain][record.Client]++
	}

	// Look for suspicious patterns
	for domain, clients := range domainClientQueries {
		// Check for long domain names (potential data encoding)
		if len(domain) > 100 {
			for client, count := range clients {
				if count > 50 { // High volume of long queries
					anomaly := types.PacketAnomaly{
						ID:          fmt.Sprintf("dns_tunneling_%s_%s", strings.ReplaceAll(client, ".", "_"), strings.ReplaceAll(domain, ".", "_")),
						Type:        "dns_tunneling",
						Description: fmt.Sprintf("Potential DNS tunneling: client %s, domain %s (%d queries)", client, domain, count),
						Severity:    "HIGH",
						Timestamp:   time.Now().Format(time.RFC3339),
						SourceIP:    client,
						Confidence:  0.8,
					}
					anomalies = append(anomalies, anomaly)
				}
			}
		}
	}

	return anomalies
}

// detectPortScanning detects potential port scanning patterns
func (d *DefaultDeepPacketInspector) detectPortScanning(records []types.PiholeRecord) []types.PacketAnomaly {
	anomalies := make([]types.PacketAnomaly, 0)

	// For DNS records, port scanning would be unusual
	// We can detect rapid-fire queries to many different domains from the same client
	clientDomains := make(map[string]map[string]bool)

	for _, record := range records {
		if _, exists := clientDomains[record.Client]; !exists {
			clientDomains[record.Client] = make(map[string]bool)
		}
		clientDomains[record.Client][record.Domain] = true
	}

	// Check for clients querying many unique domains (potential reconnaissance)
	for client, domains := range clientDomains {
		if len(domains) > 100 { // Threshold for suspicious behavior
			anomaly := types.PacketAnomaly{
				ID:          fmt.Sprintf("recon_activity_%s", strings.ReplaceAll(client, ".", "_")),
				Type:        "reconnaissance",
				Description: fmt.Sprintf("Potential reconnaissance activity: client %s queried %d unique domains", client, len(domains)),
				Severity:    "MEDIUM",
				Timestamp:   time.Now().Format(time.RFC3339),
				SourceIP:    client,
				Confidence:  0.7,
			}
			anomalies = append(anomalies, anomaly)
		}
	}

	return anomalies
}

// inferDestinationIP attempts to infer destination IP from domain
func (d *DefaultDeepPacketInspector) inferDestinationIP(record types.PiholeRecord) string {
	// For Pi-hole records, we don't have actual destination IPs
	// We can categorize by domain type or use domain as identifier

	domain := record.Domain

	// Check if domain looks like an IP address
	if isIPv4(domain) {
		return domain
	}

	// For actual implementation, you might want to:
	// 1. Perform DNS resolution to get IP
	// 2. Use cached resolution results
	// 3. Categorize by domain patterns

	// For now, return empty string to indicate no specific destination IP
	return ""
}
