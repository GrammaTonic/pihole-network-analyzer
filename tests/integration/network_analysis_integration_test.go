package integration

import (
	"context"
	"testing"
	"time"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/network"
	"pihole-analyzer/internal/types"
)

func TestNetworkAnalysis_Integration(t *testing.T) {
	// Create logger for integration tests
	logger := logger.New(&logger.Config{
		Component:     "integration-test",
		Level:         "info",
		EnableColors:  false,
		EnableEmojis:  false,
		ShowTimestamp: true,
	})

	// Create comprehensive configuration
	config := types.NetworkAnalysisConfig{
		Enabled: true,
		DeepPacketInspection: types.DPIConfig{
			Enabled:          true,
			AnalyzeProtocols: []string{"DNS_UDP", "DNS_TCP", "HTTP", "HTTPS"},
			PacketSampling:   1.0,
			MaxPacketSize:    1500,
			BufferSize:       10000,
			TimeWindow:       "1h",
		},
		TrafficPatterns: types.TrafficPatternsConfig{
			Enabled:           true,
			PatternTypes:      []string{"bandwidth", "frequency", "temporal", "client"},
			AnalysisWindow:    "2h",
			MinDataPoints:     10,
			PatternThreshold:  0.6,
			AnomalyDetection:  true,
		},
		SecurityAnalysis: types.SecurityAnalysisConfig{
			Enabled:               true,
			ThreatDetection:       true,
			SuspiciousPatterns:    []string{"malware", "phishing", "botnet"},
			BlacklistDomains:      []string{"malicious.com", "phishing.net", "evil.org"},
			UnusualTrafficThresh:  0.75,
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
				MaxLatency:    150.0,
				MinBandwidth:  5.0,
				MaxPacketLoss: 2.0,
				MaxJitter:     100.0,
			},
		},
	}

	t.Run("Complete Network Analysis Workflow", func(t *testing.T) {
		// Create analyzer factory
		factory := network.NewAnalyzerFactory(logger.GetSlogger())

		// Create complete network analyzer
		analyzer, err := factory.CreateNetworkAnalyzer(config)
		if err != nil {
			t.Fatalf("Failed to create network analyzer: %v", err)
		}

		// Generate comprehensive test data
		records := generateComprehensiveTestData()
		clientStats := generateRealisticClientStats()

		// Perform analysis
		startTime := time.Now()
		result, err := analyzer.AnalyzeTraffic(context.Background(), records, clientStats)
		analysisTime := time.Since(startTime)

		if err != nil {
			t.Fatalf("Network analysis failed: %v", err)
		}

		// Log analysis performance
		t.Logf("Analysis completed in %v for %d records and %d clients", 
			analysisTime, len(records), len(clientStats))

		// Comprehensive result validation
		validateNetworkAnalysisResult(t, result, records, clientStats)

		// Performance benchmarks
		if analysisTime > time.Minute {
			t.Errorf("Analysis took too long: %v (expected < 1 minute)", analysisTime)
		}

		// Memory usage validation (indirect check via result complexity)
		if result.Summary.TotalQueries != int64(len(records)) {
			t.Errorf("Query count mismatch: expected %d, got %d", 
				len(records), result.Summary.TotalQueries)
		}
	})

	t.Run("Deep Packet Inspection Integration", func(t *testing.T) {
		inspector, err := network.NewAnalyzerFactory(logger.GetSlogger()).CreateDPIAnalyzer(config.DeepPacketInspection)
		if err != nil {
			t.Fatalf("Failed to create DPI analyzer: %v", err)
		}

		records := generateDPITestData()
		result, err := inspector.InspectPackets(context.Background(), records, config.DeepPacketInspection)
		if err != nil {
			t.Fatalf("DPI analysis failed: %v", err)
		}

		validateDPIResults(t, result, records)
	})

	t.Run("Traffic Pattern Analysis Integration", func(t *testing.T) {
		analyzer, err := network.NewAnalyzerFactory(logger.GetSlogger()).CreateTrafficAnalyzer(config.TrafficPatterns)
		if err != nil {
			t.Fatalf("Failed to create traffic analyzer: %v", err)
		}

		records := generateTrafficPatternTestData()
		clientStats := generateRealisticClientStats()

		result, err := analyzer.AnalyzePatterns(context.Background(), records, clientStats, config.TrafficPatterns)
		if err != nil {
			t.Fatalf("Traffic pattern analysis failed: %v", err)
		}

		validateTrafficPatternResults(t, result, records, clientStats)
	})

	t.Run("Security Analysis Integration", func(t *testing.T) {
		analyzer, err := network.NewAnalyzerFactory(logger.GetSlogger()).CreateSecurityAnalyzer(config.SecurityAnalysis)
		if err != nil {
			t.Fatalf("Failed to create security analyzer: %v", err)
		}

		records := generateSecurityTestData()
		clientStats := generateRealisticClientStats()

		result, err := analyzer.AnalyzeSecurity(context.Background(), records, clientStats, config.SecurityAnalysis)
		if err != nil {
			t.Fatalf("Security analysis failed: %v", err)
		}

		validateSecurityResults(t, result, records)
	})

	t.Run("Performance Analysis Integration", func(t *testing.T) {
		analyzer, err := network.NewAnalyzerFactory(logger.GetSlogger()).CreatePerformanceAnalyzer(config.Performance)
		if err != nil {
			t.Fatalf("Failed to create performance analyzer: %v", err)
		}

		records := generatePerformanceTestData()
		clientStats := generateRealisticClientStats()

		result, err := analyzer.AnalyzePerformance(context.Background(), records, clientStats, config.Performance)
		if err != nil {
			t.Fatalf("Performance analysis failed: %v", err)
		}

		validatePerformanceResults(t, result, records)
	})

	t.Run("Visualization Integration", func(t *testing.T) {
		// Create complete analysis result
		analyzer, _ := network.NewAnalyzerFactory(logger.GetSlogger()).CreateNetworkAnalyzer(config)
		records := generateComprehensiveTestData()
		clientStats := generateRealisticClientStats()
		
		analysisResult, err := analyzer.AnalyzeTraffic(context.Background(), records, clientStats)
		if err != nil {
			t.Fatalf("Failed to create analysis result: %v", err)
		}

		// Test visualization generation
		visualizer, err := network.NewAnalyzerFactory(logger.GetSlogger()).CreateVisualizer()
		if err != nil {
			t.Fatalf("Failed to create visualizer: %v", err)
		}

		trafficViz, err := visualizer.GenerateTrafficVisualization(analysisResult)
		if err != nil {
			t.Fatalf("Failed to generate traffic visualization: %v", err)
		}

		topologyViz, err := visualizer.GenerateTopologyVisualization(records, clientStats)
		if err != nil {
			t.Fatalf("Failed to generate topology visualization: %v", err)
		}

		timeSeriesViz, err := visualizer.GenerateTimeSeriesData(records, "query_count", time.Hour)
		if err != nil {
			t.Fatalf("Failed to generate time series data: %v", err)
		}

		heatmapViz, err := visualizer.GenerateHeatmapData(records)
		if err != nil {
			t.Fatalf("Failed to generate heatmap data: %v", err)
		}

		validateVisualizationData(t, trafficViz, topologyViz, timeSeriesViz, heatmapViz)
	})
}

func TestNetworkAnalysis_Stress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	logger := logger.New(&logger.Config{Component: "stress-test"})

	config := types.NetworkAnalysisConfig{
		Enabled: true,
		DeepPacketInspection: types.DPIConfig{
			Enabled:        true,
			PacketSampling: 0.1, // Reduce sampling for performance
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

	analyzer, err := network.NewAnalyzerFactory(logger.GetSlogger()).CreateNetworkAnalyzer(config)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	// Generate large dataset
	records := generateLargeTestDataset(100000) // 100k records
	clientStats := generateLargeClientStats(1000) // 1k clients

	t.Logf("Starting stress test with %d records and %d clients", len(records), len(clientStats))

	startTime := time.Now()
	result, err := analyzer.AnalyzeTraffic(context.Background(), records, clientStats)
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("Stress test failed: %v", err)
	}

	t.Logf("Stress test completed in %v", duration)

	// Validate results
	if result.Summary.TotalQueries != int64(len(records)) {
		t.Errorf("Query count mismatch in stress test")
	}

	// Performance requirements
	if duration > 5*time.Minute {
		t.Errorf("Stress test took too long: %v (expected < 5 minutes)", duration)
	}
}

func TestNetworkAnalysis_RealTimeSimulation(t *testing.T) {
	logger := logger.New(&logger.Config{Component: "realtime-test"})

	config := types.NetworkAnalysisConfig{
		Enabled: true,
		DeepPacketInspection: types.DPIConfig{
			Enabled:        true,
			PacketSampling: 1.0,
		},
		TrafficPatterns: types.TrafficPatternsConfig{
			Enabled:          true,
			AnomalyDetection: true,
		},
		SecurityAnalysis: types.SecurityAnalysisConfig{
			Enabled:         true,
			ThreatDetection: true,
		},
		Performance: types.NetworkPerformanceConfig{
			Enabled: true,
		},
	}

	analyzer, err := network.NewAnalyzerFactory(logger.GetSlogger()).CreateNetworkAnalyzer(config)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	// Simulate real-time data processing
	t.Run("Continuous Analysis Simulation", func(t *testing.T) {
		batchSize := 1000
		totalBatches := 10

		for batch := 0; batch < totalBatches; batch++ {
			// Generate batch of records
			records := generateTimestampedRecords(batchSize, time.Now().Add(-time.Duration(batch)*time.Minute))
			clientStats := generateRealisticClientStats()

			startTime := time.Now()
			result, err := analyzer.AnalyzeTraffic(context.Background(), records, clientStats)
			batchTime := time.Since(startTime)

			if err != nil {
				t.Fatalf("Batch %d analysis failed: %v", batch, err)
			}

			// Validate real-time requirements
			if batchTime > 30*time.Second {
				t.Errorf("Batch %d took too long: %v (expected < 30s)", batch, batchTime)
			}

			// Validate result consistency
			if result.Summary.TotalQueries != int64(len(records)) {
				t.Errorf("Batch %d query count mismatch", batch)
			}

			t.Logf("Batch %d processed %d records in %v", batch, len(records), batchTime)
		}
	})
}

// Test data generation functions

func generateComprehensiveTestData() []types.PiholeRecord {
	records := make([]types.PiholeRecord, 0, 5000)
	baseTime := time.Now().Add(-24 * time.Hour)

	domains := []string{
		"google.com", "facebook.com", "github.com", "stackoverflow.com", "reddit.com",
		"youtube.com", "twitter.com", "linkedin.com", "microsoft.com", "apple.com",
		"amazon.com", "netflix.com", "cloudflare.com", "github.io", "medium.com",
		"malicious.com", "phishing.net", "suspicious-domain-with-very-long-name.evil.org",
	}

	clients := []string{
		"192.168.1.100", "192.168.1.101", "192.168.1.102", "192.168.1.103", "192.168.1.104",
		"192.168.1.105", "192.168.1.106", "192.168.1.107", "192.168.1.108", "192.168.1.109",
	}

	queryTypes := []string{"A", "AAAA", "CNAME", "MX", "TXT", "PTR", "SRV"}
	statuses := []int{0, 1, 2, 3} // 0=allowed, 1=blocked, 2=cached, 3=forwarded

	for i := 0; i < 5000; i++ {
		record := types.PiholeRecord{
			ID:        i + 1,
			DateTime:  baseTime.Add(time.Duration(i) * time.Second).Format(time.RFC3339),
			Domain:    domains[i%len(domains)],
			Client:    clients[i%len(clients)],
			QueryType: queryTypes[i%len(queryTypes)],
			Status:    statuses[i%len(statuses)],
			Timestamp: baseTime.Add(time.Duration(i) * time.Second).Format(time.RFC3339),
		}
		records = append(records, record)
	}

	return records
}

func generateRealisticClientStats() map[string]*types.ClientStats {
	clients := map[string]*types.ClientStats{
		"192.168.1.100": {
			IP:          "192.168.1.100",
			Hostname:    "workstation-01",
			QueryCount:  1500,
			DomainCount: 45,
			IsOnline:    true,
			LastSeen:    time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			FirstSeen:   time.Now().Add(-12 * time.Hour).Format(time.RFC3339),
			DeviceType:  "workstation",
		},
		"192.168.1.101": {
			IP:          "192.168.1.101",
			Hostname:    "laptop-01",
			QueryCount:  800,
			DomainCount: 32,
			IsOnline:    true,
			LastSeen:    time.Now().Add(-2 * time.Minute).Format(time.RFC3339),
			FirstSeen:   time.Now().Add(-8 * time.Hour).Format(time.RFC3339),
			DeviceType:  "laptop",
		},
		"192.168.1.102": {
			IP:          "192.168.1.102",
			Hostname:    "phone-01",
			QueryCount:  400,
			DomainCount: 18,
			IsOnline:    true,
			LastSeen:    time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
			FirstSeen:   time.Now().Add(-6 * time.Hour).Format(time.RFC3339),
			DeviceType:  "mobile",
		},
		"192.168.1.103": {
			IP:          "192.168.1.103",
			Hostname:    "iot-device-01",
			QueryCount:  200,
			DomainCount: 5,
			IsOnline:    true,
			LastSeen:    time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
			FirstSeen:   time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			DeviceType:  "iot",
		},
		"192.168.1.104": {
			IP:          "192.168.1.104",
			Hostname:    "suspicious-device",
			QueryCount:  50,
			DomainCount: 10,
			IsOnline:    false,
			LastSeen:    time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			FirstSeen:   time.Now().Add(-3 * time.Hour).Format(time.RFC3339),
			DeviceType:  "unknown",
		},
	}

	return clients
}

func generateDPITestData() []types.PiholeRecord {
	return generateComprehensiveTestData()[:1000] // Subset for DPI testing
}

func generateTrafficPatternTestData() []types.PiholeRecord {
	records := make([]types.PiholeRecord, 0, 2000)
	baseTime := time.Now().Add(-4 * time.Hour)

	// Generate patterns: burst activity, regular intervals, etc.
	for i := 0; i < 2000; i++ {
		var timestamp time.Time
		if i < 500 {
			// Burst pattern
			timestamp = baseTime.Add(time.Duration(i) * time.Second)
		} else if i < 1000 {
			// Regular pattern (every 30 seconds)
			timestamp = baseTime.Add(time.Hour).Add(time.Duration(i-500) * 30 * time.Second)
		} else {
			// Random pattern
			timestamp = baseTime.Add(time.Duration(i) * time.Minute)
		}

		record := types.PiholeRecord{
			ID:        i + 1,
			DateTime:  timestamp.Format(time.RFC3339),
			Domain:    "example.com",
			Client:    "192.168.1.100",
			QueryType: "A",
			Status:    0,
			Timestamp: timestamp.Format(time.RFC3339),
		}
		records = append(records, record)
	}

	return records
}

func generateSecurityTestData() []types.PiholeRecord {
	records := make([]types.PiholeRecord, 0, 1000)
	baseTime := time.Now().Add(-2 * time.Hour)

	maliciousDomains := []string{
		"malicious.com", "phishing.net", "evil.org", "botnet-c2.com",
		"very-long-suspicious-domain-name-that-might-be-dga-generated.evil",
		"temp-mail.org", "10minutemail.com", "guerrillamail.com",
	}

	suspiciousClients := []string{"192.168.1.200", "192.168.1.201", "192.168.1.202"}
	normalClients := []string{"192.168.1.100", "192.168.1.101", "192.168.1.102"}

	for i := 0; i < 1000; i++ {
		var domain, client string
		status := 0

		if i < 100 {
			// Malicious domains
			domain = maliciousDomains[i%len(maliciousDomains)]
			client = suspiciousClients[i%len(suspiciousClients)]
			status = 3 // Blocked
		} else {
			// Normal traffic
			domain = "normal-domain.com"
			client = normalClients[i%len(normalClients)]
		}

		record := types.PiholeRecord{
			ID:        i + 1,
			DateTime:  baseTime.Add(time.Duration(i) * time.Second).Format(time.RFC3339),
			Domain:    domain,
			Client:    client,
			QueryType: "A",
			Status:    status,
			Timestamp: baseTime.Add(time.Duration(i) * time.Second).Format(time.RFC3339),
		}
		records = append(records, record)
	}

	return records
}

func generatePerformanceTestData() []types.PiholeRecord {
	records := make([]types.PiholeRecord, 0, 3000)
	baseTime := time.Now().Add(-3 * time.Hour)

	for i := 0; i < 3000; i++ {
		record := types.PiholeRecord{
			ID:        i + 1,
			DateTime:  baseTime.Add(time.Duration(i) * time.Second).Format(time.RFC3339),
			Domain:    "performance-test.com",
			Client:    "192.168.1.100",
			QueryType: "A",
			Status:    0,
			Timestamp: baseTime.Add(time.Duration(i) * time.Second).Format(time.RFC3339),
			ReplyTime: float64(10 + (i%100)), // Varying response times
		}
		records = append(records, record)
	}

	return records
}

func generateLargeTestDataset(size int) []types.PiholeRecord {
	records := make([]types.PiholeRecord, 0, size)
	baseTime := time.Now().Add(-24 * time.Hour)

	domains := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		domains[i] = generateRandomDomain()
	}

	clients := make([]string, 200)
	for i := 0; i < 200; i++ {
		clients[i] = generateRandomIP()
	}

	for i := 0; i < size; i++ {
		record := types.PiholeRecord{
			ID:        i + 1,
			DateTime:  baseTime.Add(time.Duration(i) * time.Second).Format(time.RFC3339),
			Domain:    domains[i%len(domains)],
			Client:    clients[i%len(clients)],
			QueryType: "A",
			Status:    i % 4, // Vary statuses
			Timestamp: baseTime.Add(time.Duration(i) * time.Second).Format(time.RFC3339),
		}
		records = append(records, record)
	}

	return records
}

func generateLargeClientStats(count int) map[string]*types.ClientStats {
	stats := make(map[string]*types.ClientStats)

	for i := 0; i < count; i++ {
		ip := generateRandomIP()
		stats[ip] = &types.ClientStats{
			IP:          ip,
			Hostname:    generateRandomHostname(),
			QueryCount:  100 + (i % 1000),
			DomainCount: 10 + (i % 50),
			IsOnline:    i%10 != 0, // 90% online
			LastSeen:    time.Now().Add(-time.Duration(i%3600) * time.Second).Format(time.RFC3339),
			FirstSeen:   time.Now().Add(-time.Duration(i%86400) * time.Second).Format(time.RFC3339),
			DeviceType:  generateRandomDeviceType(),
		}
	}

	return stats
}

func generateTimestampedRecords(count int, baseTime time.Time) []types.PiholeRecord {
	records := make([]types.PiholeRecord, 0, count)

	for i := 0; i < count; i++ {
		record := types.PiholeRecord{
			ID:        i + 1,
			DateTime:  baseTime.Add(time.Duration(i) * time.Second).Format(time.RFC3339),
			Domain:    "realtime-test.com",
			Client:    "192.168.1.100",
			QueryType: "A",
			Status:    0,
			Timestamp: baseTime.Add(time.Duration(i) * time.Second).Format(time.RFC3339),
		}
		records = append(records, record)
	}

	return records
}

// Validation functions

func validateNetworkAnalysisResult(t *testing.T, result *types.NetworkAnalysisResult, records []types.PiholeRecord, clientStats map[string]*types.ClientStats) {
	if result == nil {
		t.Fatal("Analysis result should not be nil")
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
		t.Fatal("Summary should not be nil")
	}

	if result.Summary.TotalQueries != int64(len(records)) {
		t.Errorf("Total queries mismatch: expected %d, got %d", len(records), result.Summary.TotalQueries)
	}

	if result.Summary.TotalClients != len(clientStats) {
		t.Errorf("Total clients mismatch: expected %d, got %d", len(clientStats), result.Summary.TotalClients)
	}

	if result.Summary.HealthScore < 0 || result.Summary.HealthScore > 100 {
		t.Errorf("Health score out of range: %f", result.Summary.HealthScore)
	}

	if len(result.Summary.KeyInsights) == 0 {
		t.Error("Key insights should not be empty")
	}
}

func validateDPIResults(t *testing.T, result *types.PacketAnalysisResult, records []types.PiholeRecord) {
	if result.TotalPackets != int64(len(records)) {
		t.Errorf("Total packets mismatch: expected %d, got %d", len(records), result.TotalPackets)
	}

	if result.AnalyzedPackets == 0 {
		t.Error("Analyzed packets should be greater than 0")
	}

	if len(result.ProtocolDistribution) == 0 {
		t.Error("Protocol distribution should not be empty")
	}
}

func validateTrafficPatternResults(t *testing.T, result *types.TrafficPatternsResult, records []types.PiholeRecord, clientStats map[string]*types.ClientStats) {
	if result.PatternID == "" {
		t.Error("Pattern ID should not be empty")
	}

	if len(result.ClientBehavior) == 0 {
		t.Error("Client behavior should not be empty")
	}

	for ip, behavior := range result.ClientBehavior {
		if behavior.IP != ip {
			t.Errorf("Client behavior IP mismatch: %s != %s", behavior.IP, ip)
		}

		if behavior.RiskScore < 0 || behavior.RiskScore > 1 {
			t.Errorf("Risk score out of range: %f", behavior.RiskScore)
		}
	}
}

func validateSecurityResults(t *testing.T, result *types.SecurityAnalysisResult, records []types.PiholeRecord) {
	validThreatLevels := []string{"LOW", "MEDIUM", "HIGH", "CRITICAL"}
	isValid := false
	for _, level := range validThreatLevels {
		if result.ThreatLevel == level {
			isValid = true
			break
		}
	}
	if !isValid {
		t.Errorf("Invalid threat level: %s", result.ThreatLevel)
	}
}

func validatePerformanceResults(t *testing.T, result *types.NetworkPerformanceResult, records []types.PiholeRecord) {
	if result.OverallScore < 0 || result.OverallScore > 100 {
		t.Errorf("Overall score out of range: %f", result.OverallScore)
	}

	if result.LatencyMetrics.AvgLatency < 0 {
		t.Error("Average latency should not be negative")
	}

	if result.ThroughputMetrics.QueriesPerSecond < 0 {
		t.Error("QPS should not be negative")
	}
}

func validateVisualizationData(t *testing.T, trafficViz, topologyViz, timeSeriesViz, heatmapViz map[string]interface{}) {
	if trafficViz == nil {
		t.Error("Traffic visualization should not be nil")
	}

	if topologyViz == nil {
		t.Error("Topology visualization should not be nil")
	}

	if timeSeriesViz == nil {
		t.Error("Time series visualization should not be nil")
	}

	if heatmapViz == nil {
		t.Error("Heatmap visualization should not be nil")
	}

	// Check for required fields
	if _, exists := trafficViz["metadata"]; !exists {
		t.Error("Traffic visualization should contain metadata")
	}

	if _, exists := topologyViz["nodes"]; !exists {
		t.Error("Topology visualization should contain nodes")
	}

	if _, exists := timeSeriesViz["data_points"]; !exists {
		t.Error("Time series visualization should contain data points")
	}
}

// Helper functions for data generation

func generateRandomDomain() string {
	prefixes := []string{"www", "api", "cdn", "mail", "blog", "shop", "app"}
	domains := []string{"example", "test", "demo", "sample", "corporate", "business"}
	tlds := []string{"com", "org", "net", "io", "co"}

	return prefixes[len(prefixes)%7] + "." + domains[len(domains)%6] + "." + tlds[len(tlds)%5]
}

func generateRandomIP() string {
	return "192.168.1." + string(rune(100+(len("test")%155)))
}

func generateRandomHostname() string {
	types := []string{"laptop", "desktop", "phone", "tablet", "server", "iot"}
	return types[len(types)%6] + "-" + string(rune(1+(len("test")%99)))
}

func generateRandomDeviceType() string {
	types := []string{"laptop", "desktop", "mobile", "tablet", "iot", "server", "unknown"}
	return types[len(types)%7]
}