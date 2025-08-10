package network

import (
	"context"
	"testing"
	"time"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

func TestEnhancedNetworkAnalyzer_Initialize(t *testing.T) {
	customLogger := logger.New(&logger.Config{Component: "test"})
	slogger := customLogger.GetSlogger()
	analyzer := NewEnhancedNetworkAnalyzer(slogger)

	config := types.NetworkAnalysisConfig{
		Enabled: true,
		DeepPacketInspection: types.DPIConfig{
			Enabled:          true,
			AnalyzeProtocols: []string{"DNS_UDP", "DNS_TCP"},
			PacketSampling:   1.0,
			MaxPacketSize:    1500,
			BufferSize:       1000,
			TimeWindow:       "1h",
		},
		TrafficPatterns: types.TrafficPatternsConfig{
			Enabled:          true,
			PatternTypes:     []string{"bandwidth", "temporal", "client"},
			AnalysisWindow:   "1h",
			MinDataPoints:    10,
			PatternThreshold: 0.5,
			AnomalyDetection: true,
		},
		SecurityAnalysis: types.SecurityAnalysisConfig{
			Enabled:               true,
			ThreatDetection:       true,
			SuspiciousPatterns:    []string{"malware", "phishing"},
			BlacklistDomains:      []string{"malicious.com", "phishing.net"},
			UnusualTrafficThresh:  0.8,
			PortScanDetection:     true,
			DNSTunnelingDetection: true,
		},
		Performance: types.NetworkPerformanceConfig{
			Enabled:             true,
			LatencyAnalysis:     true,
			BandwidthAnalysis:   true,
			ThroughputAnalysis:  true,
			PacketLossDetection: true,
			JitterAnalysis:      true,
			QualityThresholds: types.QualityThresholds{
				MaxLatency:    100.0,
				MinBandwidth:  10.0,
				MaxPacketLoss: 1.0,
				MaxJitter:     50.0,
			},
		},
	}

	err := analyzer.Initialize(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to initialize analyzer: %v", err)
	}

	if !analyzer.IsHealthy() {
		t.Error("Analyzer should be healthy after initialization")
	}

	capabilities := analyzer.GetCapabilities()
	expectedCapabilities := []string{
		"basic_analysis", "deep_packet_inspection", "protocol_analysis",
		"packet_anomaly_detection", "traffic_patterns", "bandwidth_analysis",
		"temporal_patterns", "client_behavior", "security_analysis",
		"threat_detection", "dns_anomalies", "port_scan_detection",
		"performance_analysis", "latency_analysis", "throughput_analysis",
		"quality_assessment",
	}

	if len(capabilities) < len(expectedCapabilities) {
		t.Errorf("Expected at least %d capabilities, got %d", len(expectedCapabilities), len(capabilities))
	}
}

func TestEnhancedNetworkAnalyzer_AnalyzeTraffic(t *testing.T) {
	customLogger := logger.New(&logger.Config{Component: "test"})
	slogger := customLogger.GetSlogger()
	analyzer := NewEnhancedNetworkAnalyzer(slogger)

	config := types.NetworkAnalysisConfig{
		Enabled: true,
		DeepPacketInspection: types.DPIConfig{
			Enabled:        true,
			PacketSampling: 1.0,
			TimeWindow:     "1h",
		},
		TrafficPatterns: types.TrafficPatternsConfig{
			Enabled:          true,
			PatternTypes:     []string{"bandwidth", "temporal"},
			AnalysisWindow:   "1h",
			AnomalyDetection: true,
		},
		SecurityAnalysis: types.SecurityAnalysisConfig{
			Enabled:         true,
			ThreatDetection: true,
		},
		Performance: types.NetworkPerformanceConfig{
			Enabled:           true,
			LatencyAnalysis:   true,
			BandwidthAnalysis: true,
		},
	}

	err := analyzer.Initialize(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to initialize analyzer: %v", err)
	}

	// Create test data
	records := createTestRecords()
	clientStats := createTestClientStats()

	result, err := analyzer.AnalyzeTraffic(context.Background(), records, clientStats)
	if err != nil {
		t.Fatalf("Failed to analyze traffic: %v", err)
	}

	// Verify result structure
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.AnalysisID == "" {
		t.Error("Analysis ID should not be empty")
	}

	if result.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}

	if result.Duration == "" {
		t.Error("Duration should not be empty")
	}

	if result.Summary == nil {
		t.Error("Summary should not be nil")
	}

	// Verify summary data
	if result.Summary.TotalClients != len(clientStats) {
		t.Errorf("Expected %d total clients, got %d", len(clientStats), result.Summary.TotalClients)
	}

	if result.Summary.TotalQueries != int64(len(records)) {
		t.Errorf("Expected %d total queries, got %d", len(records), result.Summary.TotalQueries)
	}

	if result.Summary.HealthScore < 0 || result.Summary.HealthScore > 100 {
		t.Errorf("Health score should be between 0 and 100, got %f", result.Summary.HealthScore)
	}

	// Verify that analysis components ran
	if result.PacketAnalysis == nil {
		t.Error("Packet analysis should not be nil when DPI is enabled")
	}

	if result.TrafficPatterns == nil {
		t.Error("Traffic patterns should not be nil when traffic analysis is enabled")
	}

	if result.SecurityAnalysis == nil {
		t.Error("Security analysis should not be nil when security analysis is enabled")
	}

	if result.Performance == nil {
		t.Error("Performance analysis should not be nil when performance analysis is enabled")
	}
}

func TestDeepPacketInspector_InspectPackets(t *testing.T) {
	customLogger := logger.New(&logger.Config{Component: "test"})
	slogger := customLogger.GetSlogger()
	inspector := NewDeepPacketInspector(slogger)

	config := types.DPIConfig{
		Enabled:          true,
		AnalyzeProtocols: []string{"DNS_UDP", "DNS_TCP"},
		PacketSampling:   1.0,
		MaxPacketSize:    1500,
		BufferSize:       1000,
		TimeWindow:       "1h",
	}

	records := createTestRecords()

	result, err := inspector.InspectPackets(context.Background(), records, config)
	if err != nil {
		t.Fatalf("Failed to inspect packets: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.TotalPackets != int64(len(records)) {
		t.Errorf("Expected %d total packets, got %d", len(records), result.TotalPackets)
	}

	if result.AnalyzedPackets == 0 {
		t.Error("Analyzed packets should be greater than 0")
	}

	if len(result.ProtocolDistribution) == 0 {
		t.Error("Protocol distribution should not be empty")
	}

	if len(result.TopSourceIPs) == 0 {
		t.Error("Top source IPs should not be empty")
	}
}

func TestTrafficPatternAnalyzer_AnalyzePatterns(t *testing.T) {
	customLogger := logger.New(&logger.Config{Component: "test"})
	slogger := customLogger.GetSlogger()
	analyzer := NewTrafficPatternAnalyzer(slogger)

	config := types.TrafficPatternsConfig{
		Enabled:          true,
		PatternTypes:     []string{"bandwidth", "temporal", "client"},
		AnalysisWindow:   "1h",
		MinDataPoints:    5,
		PatternThreshold: 0.5,
		AnomalyDetection: true,
	}

	records := createTestRecords()
	clientStats := createTestClientStats()

	result, err := analyzer.AnalyzePatterns(context.Background(), records, clientStats, config)
	if err != nil {
		t.Fatalf("Failed to analyze traffic patterns: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.PatternID == "" {
		t.Error("Pattern ID should not be empty")
	}

	// Check that pattern analysis ran
	if len(result.BandwidthPatterns) == 0 {
		t.Log("No bandwidth patterns detected (this may be normal for test data)")
	}

	if len(result.TemporalPatterns) == 0 {
		t.Log("No temporal patterns detected (this may be normal for test data)")
	}

	if len(result.ClientBehavior) == 0 {
		t.Error("Client behavior analysis should produce results")
	}

	// Verify client behavior analysis
	for clientIP, behavior := range result.ClientBehavior {
		if behavior.IP != clientIP {
			t.Errorf("Client behavior IP mismatch: expected %s, got %s", clientIP, behavior.IP)
		}

		if behavior.BehaviorType == "" {
			t.Error("Behavior type should not be empty")
		}

		if behavior.ActivityLevel == "" {
			t.Error("Activity level should not be empty")
		}

		if behavior.RiskScore < 0 || behavior.RiskScore > 1 {
			t.Errorf("Risk score should be between 0 and 1, got %f", behavior.RiskScore)
		}
	}
}

func TestSecurityAnalyzer_AnalyzeSecurity(t *testing.T) {
	customLogger := logger.New(&logger.Config{Component: "test"})
	slogger := customLogger.GetSlogger()
	analyzer := NewSecurityAnalyzer(slogger)

	config := types.SecurityAnalysisConfig{
		Enabled:               true,
		ThreatDetection:       true,
		SuspiciousPatterns:    []string{"malware", "phishing"},
		BlacklistDomains:      []string{"malicious.com", "phishing.net"},
		UnusualTrafficThresh:  0.8,
		PortScanDetection:     true,
		DNSTunnelingDetection: true,
	}

	records := createTestRecords()
	clientStats := createTestClientStats()

	result, err := analyzer.AnalyzeSecurity(context.Background(), records, clientStats, config)
	if err != nil {
		t.Fatalf("Failed to analyze security: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.ThreatLevel == "" {
		t.Error("Threat level should not be empty")
	}

	// Verify threat level is valid
	validThreatLevels := []string{"LOW", "MEDIUM", "HIGH", "CRITICAL"}
	isValidThreatLevel := false
	for _, level := range validThreatLevels {
		if result.ThreatLevel == level {
			isValidThreatLevel = true
			break
		}
	}
	if !isValidThreatLevel {
		t.Errorf("Invalid threat level: %s", result.ThreatLevel)
	}

	// Test threat assessment
	threatLevel := analyzer.AssessThreatLevel(result.DetectedThreats, result.SuspiciousActivity)
	if threatLevel == "" {
		t.Error("Threat level assessment should not return empty string")
	}
}

func TestPerformanceAnalyzer_AnalyzePerformance(t *testing.T) {
	customLogger := logger.New(&logger.Config{Component: "test"})
	slogger := customLogger.GetSlogger()
	analyzer := NewPerformanceAnalyzer(slogger)

	config := types.NetworkPerformanceConfig{
		Enabled:             true,
		LatencyAnalysis:     true,
		BandwidthAnalysis:   true,
		ThroughputAnalysis:  true,
		PacketLossDetection: true,
		JitterAnalysis:      true,
		QualityThresholds: types.QualityThresholds{
			MaxLatency:    100.0,
			MinBandwidth:  10.0,
			MaxPacketLoss: 1.0,
			MaxJitter:     50.0,
		},
	}

	records := createTestRecords()
	clientStats := createTestClientStats()

	result, err := analyzer.AnalyzePerformance(context.Background(), records, clientStats, config)
	if err != nil {
		t.Fatalf("Failed to analyze performance: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.OverallScore < 0 || result.OverallScore > 100 {
		t.Errorf("Overall score should be between 0 and 100, got %f", result.OverallScore)
	}

	// Verify that performance metrics are calculated
	if result.LatencyMetrics.AvgLatency < 0 {
		t.Error("Average latency should not be negative")
	}

	if result.BandwidthMetrics.TotalBandwidth < 0 {
		t.Error("Total bandwidth should not be negative")
	}

	if result.ThroughputMetrics.QueriesPerSecond < 0 {
		t.Error("Queries per second should not be negative")
	}
}

func TestNetworkVisualizer_GenerateTrafficVisualization(t *testing.T) {
	customLogger := logger.New(&logger.Config{Component: "test"})
	slogger := customLogger.GetSlogger()
	visualizer := NewNetworkVisualizer(slogger)

	// Create a complete network analysis result
	result := &types.NetworkAnalysisResult{
		Timestamp:  time.Now().Format(time.RFC3339),
		AnalysisID: "test_analysis",
		Duration:   "1m30s",
		PacketAnalysis: &types.PacketAnalysisResult{
			TotalPackets:           1000,
			AnalyzedPackets:        1000,
			ProtocolDistribution:   map[string]int64{"DNS_UDP": 800, "DNS_TCP": 200},
			PacketSizeDistribution: map[string]int64{"small": 600, "medium": 300, "large": 100},
			TopSourceIPs:           []types.IPTrafficStat{{IP: "192.168.1.100", PacketCount: 500, ByteCount: 50000, Percentage: 50.0}},
			TopDestinationIPs:      []types.IPTrafficStat{{IP: "8.8.8.8", PacketCount: 300, ByteCount: 30000, Percentage: 30.0}},
			PortUsage:              map[string]int64{"53/DNS_UDP": 800, "53/DNS_TCP": 200},
			Anomalies:              []types.PacketAnomaly{{ID: "test_anomaly", Type: "volume_spike", Severity: "MEDIUM"}},
		},
		Summary: &types.NetworkAnalysisSummary{
			TotalClients:      5,
			ActiveClients:     3,
			TotalQueries:      1000,
			AnomaliesDetected: 1,
			ThreatLevel:       "LOW",
			OverallHealth:     "GOOD",
			HealthScore:       85.0,
			KeyInsights:       []string{"Network operating normally"},
		},
	}

	viz, err := visualizer.GenerateTrafficVisualization(result)
	if err != nil {
		t.Fatalf("Failed to generate traffic visualization: %v", err)
	}

	if viz == nil {
		t.Fatal("Visualization data should not be nil")
	}

	// Check that visualization contains expected sections
	if _, exists := viz["packet_analysis"]; !exists {
		t.Error("Visualization should contain packet_analysis section")
	}

	if _, exists := viz["summary"]; !exists {
		t.Error("Visualization should contain summary section")
	}

	if _, exists := viz["metadata"]; !exists {
		t.Error("Visualization should contain metadata section")
	}
}

func TestAnalyzerFactory_CreateAnalyzers(t *testing.T) {
	customLogger := logger.New(&logger.Config{Component: "test"})
	slogger := customLogger.GetSlogger()
	factory := NewAnalyzerFactory(slogger)

	// Test creating network analyzer
	config := types.NetworkAnalysisConfig{Enabled: true}
	networkAnalyzer, err := factory.CreateNetworkAnalyzer(config)
	if err != nil {
		t.Fatalf("Failed to create network analyzer: %v", err)
	}
	if networkAnalyzer == nil {
		t.Error("Network analyzer should not be nil")
	}

	// Test creating DPI analyzer
	dpiConfig := types.DPIConfig{Enabled: true}
	dpiAnalyzer, err := factory.CreateDPIAnalyzer(dpiConfig)
	if err != nil {
		t.Fatalf("Failed to create DPI analyzer: %v", err)
	}
	if dpiAnalyzer == nil {
		t.Error("DPI analyzer should not be nil")
	}

	// Test creating traffic analyzer
	trafficConfig := types.TrafficPatternsConfig{Enabled: true}
	trafficAnalyzer, err := factory.CreateTrafficAnalyzer(trafficConfig)
	if err != nil {
		t.Fatalf("Failed to create traffic analyzer: %v", err)
	}
	if trafficAnalyzer == nil {
		t.Error("Traffic analyzer should not be nil")
	}

	// Test creating security analyzer
	securityConfig := types.SecurityAnalysisConfig{Enabled: true}
	securityAnalyzer, err := factory.CreateSecurityAnalyzer(securityConfig)
	if err != nil {
		t.Fatalf("Failed to create security analyzer: %v", err)
	}
	if securityAnalyzer == nil {
		t.Error("Security analyzer should not be nil")
	}

	// Test creating performance analyzer
	performanceConfig := types.NetworkPerformanceConfig{Enabled: true}
	performanceAnalyzer, err := factory.CreatePerformanceAnalyzer(performanceConfig)
	if err != nil {
		t.Fatalf("Failed to create performance analyzer: %v", err)
	}
	if performanceAnalyzer == nil {
		t.Error("Performance analyzer should not be nil")
	}

	// Test creating visualizer
	visualizer, err := factory.CreateVisualizer()
	if err != nil {
		t.Fatalf("Failed to create visualizer: %v", err)
	}
	if visualizer == nil {
		t.Error("Visualizer should not be nil")
	}
}

// Helper functions to create test data

func createTestRecords() []types.PiholeRecord {
	now := time.Now()
	records := []types.PiholeRecord{
		{
			ID:        1,
			DateTime:  now.Add(-1 * time.Hour).Format(time.RFC3339),
			Domain:    "google.com",
			Client:    "192.168.1.100",
			QueryType: "A",
			Status:    0,
			Timestamp: now.Add(-1 * time.Hour).Format(time.RFC3339),
			ReplyTime: 15.5,
		},
		{
			ID:        2,
			DateTime:  now.Add(-50 * time.Minute).Format(time.RFC3339),
			Domain:    "facebook.com",
			Client:    "192.168.1.101",
			QueryType: "A",
			Status:    0,
			Timestamp: now.Add(-50 * time.Minute).Format(time.RFC3339),
			ReplyTime: 22.3,
		},
		{
			ID:        3,
			DateTime:  now.Add(-40 * time.Minute).Format(time.RFC3339),
			Domain:    "github.com",
			Client:    "192.168.1.100",
			QueryType: "AAAA",
			Status:    0,
			Timestamp: now.Add(-40 * time.Minute).Format(time.RFC3339),
			ReplyTime: 18.7,
		},
		{
			ID:        4,
			DateTime:  now.Add(-30 * time.Minute).Format(time.RFC3339),
			Domain:    "malicious.com",
			Client:    "192.168.1.102",
			QueryType: "A",
			Status:    3, // Blocked
			Timestamp: now.Add(-30 * time.Minute).Format(time.RFC3339),
			ReplyTime: 0.0, // Blocked queries have no reply time
		},
		{
			ID:        5,
			DateTime:  now.Add(-20 * time.Minute).Format(time.RFC3339),
			Domain:    "stackoverflow.com",
			Client:    "192.168.1.103",
			QueryType: "A",
			Status:    0,
			Timestamp: now.Add(-20 * time.Minute).Format(time.RFC3339),
			ReplyTime: 12.1,
		},
	}
	return records
}

func createTestClientStats() map[string]*types.ClientStats {
	return map[string]*types.ClientStats{
		"192.168.1.100": {
			IP:          "192.168.1.100",
			Hostname:    "laptop-01",
			QueryCount:  100,
			Domains:     map[string]int{"google.com": 50, "github.com": 30, "stackoverflow.com": 20},
			DomainCount: 3,
			IsOnline:    true,
			LastSeen:    time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
			FirstSeen:   time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			DeviceType:  "laptop",
		},
		"192.168.1.101": {
			IP:          "192.168.1.101",
			Hostname:    "desktop-01",
			QueryCount:  75,
			Domains:     map[string]int{"facebook.com": 40, "twitter.com": 25, "instagram.com": 10},
			DomainCount: 3,
			IsOnline:    true,
			LastSeen:    time.Now().Add(-15 * time.Minute).Format(time.RFC3339),
			FirstSeen:   time.Now().Add(-3 * time.Hour).Format(time.RFC3339),
			DeviceType:  "desktop",
		},
		"192.168.1.102": {
			IP:          "192.168.1.102",
			Hostname:    "suspicious-device",
			QueryCount:  25,
			Domains:     map[string]int{"malicious.com": 15, "phishing.net": 10},
			DomainCount: 2,
			IsOnline:    false,
			LastSeen:    time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			FirstSeen:   time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			DeviceType:  "unknown",
		},
		"192.168.1.103": {
			IP:          "192.168.1.103",
			Hostname:    "phone-01",
			QueryCount:  50,
			Domains:     map[string]int{"stackoverflow.com": 20, "reddit.com": 15, "youtube.com": 15},
			DomainCount: 3,
			IsOnline:    true,
			LastSeen:    time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			FirstSeen:   time.Now().Add(-4 * time.Hour).Format(time.RFC3339),
			DeviceType:  "mobile",
		},
	}
}

// Benchmark tests

func BenchmarkNetworkAnalyzer_AnalyzeTraffic(b *testing.B) {
	customLogger := logger.New(&logger.Config{Component: "benchmark"})
	slogger := customLogger.GetSlogger()
	analyzer := NewEnhancedNetworkAnalyzer(slogger)

	config := types.NetworkAnalysisConfig{
		Enabled: true,
		DeepPacketInspection: types.DPIConfig{
			Enabled:        true,
			PacketSampling: 0.1, // Sample 10% for performance
		},
		TrafficPatterns: types.TrafficPatternsConfig{
			Enabled:        true,
			AnalysisWindow: "1h",
		},
		SecurityAnalysis: types.SecurityAnalysisConfig{
			Enabled: true,
		},
		Performance: types.NetworkPerformanceConfig{
			Enabled: true,
		},
	}

	analyzer.Initialize(context.Background(), config)

	// Create larger test dataset for benchmarking
	records := make([]types.PiholeRecord, 10000)
	for i := 0; i < 10000; i++ {
		records[i] = types.PiholeRecord{
			ID:        i,
			Domain:    "example.com",
			Client:    "192.168.1.100",
			QueryType: "A",
			Status:    0,
			Timestamp: time.Now().Add(-time.Duration(i) * time.Second).Format(time.RFC3339),
		}
	}

	clientStats := createTestClientStats()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeTraffic(context.Background(), records, clientStats)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkDeepPacketInspector_InspectPackets(b *testing.B) {
	customLogger := logger.New(&logger.Config{Component: "benchmark"})
	slogger := customLogger.GetSlogger()
	inspector := NewDeepPacketInspector(slogger)

	config := types.DPIConfig{
		Enabled:        true,
		PacketSampling: 0.1, // Sample 10% for performance
	}

	records := make([]types.PiholeRecord, 1000)
	for i := 0; i < 1000; i++ {
		records[i] = types.PiholeRecord{
			ID:        i,
			Domain:    "example.com",
			Client:    "192.168.1.100",
			QueryType: "A",
			Timestamp: time.Now().Add(-time.Duration(i) * time.Second).Format(time.RFC3339),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := inspector.InspectPackets(context.Background(), records, config)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
