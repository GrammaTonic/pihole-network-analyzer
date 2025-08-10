package types

// DNSRecord represents a single DNS query record
type DNSRecord struct {
	ID             int
	DateTime       string
	Domain         string
	Type           int
	IP             string
	QueryType      string
	ResponseStatus string
	ResponseTime   float64
	// Additional fields used in main code
	Client    string
	Status    string
	ReplyTime float64
	Timestamp string
}

// PiholeRecord represents a record from Pi-hole database
type PiholeRecord struct {
	ID        int
	DateTime  string
	Domain    string
	Client    string
	QueryType string
	Status    int
	// Additional fields used in main code
	Timestamp string
	HWAddr    string
	ReplyTime float64 // Response time in milliseconds
}

// ClientStats stores statistics for each client
type ClientStats struct {
	IP            string
	Hostname      string
	QueryCount    int
	Domains       map[string]int
	DomainCount   int
	MACAddress    string
	IsOnline      bool
	LastSeen      string
	TopDomains    []DomainStat
	Status        string
	UniqueQueries int
	TotalQueries  int
	FirstSeen     string
	DeviceType    string
	// Additional fields used in main code
	Client         string
	QueryTypes     map[int]int
	StatusCodes    map[int]int
	HWAddr         string
	ARPStatus      string
	TotalReplyTime float64
	AvgReplyTime   float64
	Uniquedomains  int
}

// DomainStat represents domain statistics
type DomainStat struct {
	Domain string
	Count  int
}

// ARPEntry represents an ARP table entry
type ARPEntry struct {
	IP        string
	MAC       string
	Interface string
	Type      string
	Hostname  string
	LastSeen  string
	IsOnline  bool
}

// Config represents the application configuration
type Config struct {
	OnlineOnly      bool                  `json:"online_only"`
	NoExclude       bool                  `json:"no_exclude"`
	TestMode        bool                  `json:"test_mode"`
	Quiet           bool                  `json:"quiet"`
	Pihole          PiholeConfig          `json:"pihole"`
	Output          OutputConfig          `json:"output"`
	Exclusions      ExclusionConfig       `json:"exclusions"`
	Logging         LoggingConfig         `json:"logging"`
	Web             WebConfig             `json:"web"`
	Metrics         MetricsConfig         `json:"metrics"`
	ML              MLConfig              `json:"ml"`
	NetworkAnalysis NetworkAnalysisConfig `json:"network_analysis"`
}

// PiholeConfig represents Pi-hole specific configuration
type PiholeConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`

	// API Configuration
	APIEnabled  bool   `json:"api_enabled"`
	APIPassword string `json:"api_password"`
	APITOTP     string `json:"api_totp"`
	UseHTTPS    bool   `json:"use_https"`
	APITimeout  int    `json:"api_timeout"`
}

// OutputConfig represents output formatting configuration
type OutputConfig struct {
	Colors        bool   `json:"colors"`
	Emojis        bool   `json:"emojis"`
	Verbose       bool   `json:"verbose"`
	Format        string `json:"format"`
	MaxClients    int    `json:"max_clients"`
	MaxDomains    int    `json:"max_domains_display"`
	SaveReports   bool   `json:"save_reports"`
	ReportDir     string `json:"report_dir"`
	VerboseOutput bool   `json:"verbose_output"`
}

// ExclusionConfig represents exclusion rules
type ExclusionConfig struct {
	Networks        []string `json:"networks"`
	IPs             []string `json:"ips"`
	Domains         []string `json:"domains"`
	EnableDocker    bool     `json:"enable_docker"`
	ExcludeNetworks []string `json:"exclude_networks"`
	ExcludeIPs      []string `json:"exclude_ips"`
	ExcludeHosts    []string `json:"exclude_hosts"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level         string `json:"level"`
	EnableColors  bool   `json:"enable_colors"`
	EnableEmojis  bool   `json:"enable_emojis"`
	OutputFile    string `json:"output_file"`
	ShowTimestamp bool   `json:"show_timestamp"`
	ShowCaller    bool   `json:"show_caller"`
}

// WebConfig represents web server configuration
type WebConfig struct {
	Enabled      bool   `json:"enabled"`
	Port         int    `json:"port"`
	Host         string `json:"host"`
	DaemonMode   bool   `json:"daemon_mode"`
	ReadTimeout  int    `json:"read_timeout_seconds"`
	WriteTimeout int    `json:"write_timeout_seconds"`
	IdleTimeout  int    `json:"idle_timeout_seconds"`
}

// MetricsConfig represents metrics collection and server configuration
type MetricsConfig struct {
	Enabled               bool   `json:"enabled"`
	Port                  string `json:"port"`
	Host                  string `json:"host"`
	EnableEndpoint        bool   `json:"enable_endpoint"`
	CollectMetrics        bool   `json:"collect_metrics"`
	EnableDetailedMetrics bool   `json:"enable_detailed_metrics"`
}

// NetworkDevice represents a network device from Pi-hole API
type NetworkDevice struct {
	IP          string `json:"ip"`
	Hardware    string `json:"hardware"`
	Name        string `json:"name"`
	FirstSeen   string `json:"first_seen"`
	LastSeen    string `json:"last_seen"`
	VendorClass string `json:"vendor_class"`

	// Enhanced network analysis fields
	MAC      string `json:"mac"`       // Alias for Hardware for consistency
	Hostname string `json:"hostname"`  // Alias for Name for consistency
	Type     string `json:"type"`      // Device type classification
	IsOnline bool   `json:"is_online"` // Online status
	Vendor   string `json:"vendor"`    // Alias for VendorClass for consistency
}

// DomainAnalysis represents domain analysis data
type DomainAnalysis struct {
	TopDomains     []DomainCount  `json:"top_domains"`
	BlockedDomains []DomainCount  `json:"blocked_domains"`
	QueryTypes     map[string]int `json:"query_types"`
	TotalQueries   int            `json:"total_queries"`
	TotalBlocked   int            `json:"total_blocked"`
	BlockedPercent float64        `json:"blocked_percent"`
}

// DomainCount represents a domain with its query count
type DomainCount struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
}

// QueryPerformance represents query performance metrics
type QueryPerformance struct {
	AverageResponseTime float64 `json:"average_response_time"`
	TotalQueries        int     `json:"total_queries"`
	QueriesPerSecond    float64 `json:"queries_per_second"`
	PeakQueries         int     `json:"peak_queries"`
	SlowQueries         int     `json:"slow_queries"`
}

// ConnectionStatus represents the status of a data source connection
type ConnectionStatus struct {
	Connected    bool              `json:"connected"`
	LastConnect  string            `json:"last_connect"`
	LastError    string            `json:"last_error,omitempty"`
	ResponseTime float64           `json:"response_time"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Enhanced types for complete API integration and analyzer functionality

// AnalysisResult represents the comprehensive result of Pi-hole analysis
type AnalysisResult struct {
	ClientStats    map[string]*ClientStats `json:"client_stats"`
	NetworkDevices []NetworkDevice         `json:"network_devices"`
	TotalQueries   int                     `json:"total_queries"`
	UniqueClients  int                     `json:"unique_clients"`
	AnalysisMode   string                  `json:"analysis_mode"`
	DataSourceType string                  `json:"data_source_type"`
	Timestamp      string                  `json:"timestamp"`
	Performance    *QueryPerformance       `json:"performance,omitempty"`
}

// QueryParams represents parameters for querying DNS data
type QueryParams struct {
	StartTime    string `json:"start_time,omitempty"`
	EndTime      string `json:"end_time,omitempty"`
	ClientFilter string `json:"client_filter,omitempty"`
	DomainFilter string `json:"domain_filter,omitempty"`
	Limit        int    `json:"limit,omitempty"`
	StatusFilter []int  `json:"status_filter,omitempty"`
	TypeFilter   []int  `json:"type_filter,omitempty"`
}

// DataSourceInfo provides information about the current data source
type DataSourceInfo struct {
	Type        string       `json:"type"`
	IsConnected bool         `json:"is_connected"`
	Mode        string       `json:"mode"`
	Config      PiholeConfig `json:"config"`
	LastError   string       `json:"last_error,omitempty"`
}

// MLConfig represents configuration for machine learning features
type MLConfig struct {
	// Anomaly Detection Configuration
	AnomalyDetection AnomalyDetectionConfig `json:"anomaly_detection"`

	// Trend Analysis Configuration
	TrendAnalysis TrendAnalysisConfig `json:"trend_analysis"`

	// Model Training Configuration
	Training TrainingConfig `json:"training"`

	// Performance Configuration
	Performance PerformanceConfig `json:"performance"`
}

// AnomalyDetectionConfig configures anomaly detection
type AnomalyDetectionConfig struct {
	Enabled       bool               `json:"enabled"`
	Sensitivity   float64            `json:"sensitivity"`    // 0-1
	MinConfidence float64            `json:"min_confidence"` // 0-1
	WindowSize    string             `json:"window_size"`    // duration string
	AnomalyTypes  []string           `json:"anomaly_types"`
	Thresholds    map[string]float64 `json:"thresholds"`
}

// TrendAnalysisConfig configures trend analysis
type TrendAnalysisConfig struct {
	Enabled         bool    `json:"enabled"`
	AnalysisWindow  string  `json:"analysis_window"` // duration string
	ForecastWindow  string  `json:"forecast_window"` // duration string
	MinDataPoints   int     `json:"min_data_points"`
	SmoothingFactor float64 `json:"smoothing_factor"`
}

// TrainingConfig configures model training
type TrainingConfig struct {
	AutoRetrain     bool    `json:"auto_retrain"`
	RetrainInterval string  `json:"retrain_interval"` // duration string
	MinTrainingSize int     `json:"min_training_size"`
	MaxTrainingSize int     `json:"max_training_size"`
	ValidationSplit float64 `json:"validation_split"`
}

// PerformanceConfig configures performance settings
type PerformanceConfig struct {
	MaxConcurrency  int    `json:"max_concurrency"`
	TimeoutDuration string `json:"timeout_duration"` // duration string
	CacheEnabled    bool   `json:"cache_enabled"`
	CacheDuration   string `json:"cache_duration"` // duration string
	BatchSize       int    `json:"batch_size"`
}

// Enhanced Network Analysis Types

// NetworkAnalysisConfig configures enhanced network analysis features
type NetworkAnalysisConfig struct {
	Enabled              bool                     `json:"enabled"`
	DeepPacketInspection DPIConfig                `json:"deep_packet_inspection"`
	TrafficPatterns      TrafficPatternsConfig    `json:"traffic_patterns"`
	SecurityAnalysis     SecurityAnalysisConfig   `json:"security_analysis"`
	Performance          NetworkPerformanceConfig `json:"performance"`
}

// DPIConfig configures deep packet inspection
type DPIConfig struct {
	Enabled          bool     `json:"enabled"`
	AnalyzeProtocols []string `json:"analyze_protocols"` // TCP, UDP, ICMP, etc.
	PacketSampling   float64  `json:"packet_sampling"`   // 0.0-1.0 sampling rate
	MaxPacketSize    int      `json:"max_packet_size"`   // bytes
	BufferSize       int      `json:"buffer_size"`       // packets
	TimeWindow       string   `json:"time_window"`       // duration string
}

// TrafficPatternsConfig configures traffic pattern analysis
type TrafficPatternsConfig struct {
	Enabled          bool     `json:"enabled"`
	PatternTypes     []string `json:"pattern_types"`   // bandwidth, frequency, temporal
	AnalysisWindow   string   `json:"analysis_window"` // duration string
	MinDataPoints    int      `json:"min_data_points"`
	PatternThreshold float64  `json:"pattern_threshold"` // 0.0-1.0
	AnomalyDetection bool     `json:"anomaly_detection"`
}

// SecurityAnalysisConfig configures security analysis
type SecurityAnalysisConfig struct {
	Enabled               bool     `json:"enabled"`
	ThreatDetection       bool     `json:"threat_detection"`
	SuspiciousPatterns    []string `json:"suspicious_patterns"`
	BlacklistDomains      []string `json:"blacklist_domains"`
	UnusualTrafficThresh  float64  `json:"unusual_traffic_threshold"`
	PortScanDetection     bool     `json:"port_scan_detection"`
	DNSTunnelingDetection bool     `json:"dns_tunneling_detection"`
}

// NetworkPerformanceConfig configures network performance analysis
type NetworkPerformanceConfig struct {
	Enabled             bool              `json:"enabled"`
	LatencyAnalysis     bool              `json:"latency_analysis"`
	BandwidthAnalysis   bool              `json:"bandwidth_analysis"`
	ThroughputAnalysis  bool              `json:"throughput_analysis"`
	PacketLossDetection bool              `json:"packet_loss_detection"`
	JitterAnalysis      bool              `json:"jitter_analysis"`
	QualityThresholds   QualityThresholds `json:"quality_thresholds"`
}

// QualityThresholds defines network quality thresholds
type QualityThresholds struct {
	MaxLatency    float64 `json:"max_latency_ms"`          // milliseconds
	MinBandwidth  float64 `json:"min_bandwidth_mbps"`      // Mbps
	MaxPacketLoss float64 `json:"max_packet_loss_percent"` // percentage
	MaxJitter     float64 `json:"max_jitter_ms"`           // milliseconds
}

// NetworkAnalysisResult represents the result of enhanced network analysis
type NetworkAnalysisResult struct {
	Timestamp  string `json:"timestamp"`
	AnalysisID string `json:"analysis_id"`
	Duration   string `json:"duration"`

	// Deep Packet Inspection Results
	PacketAnalysis *PacketAnalysisResult `json:"packet_analysis,omitempty"`

	// Traffic Pattern Results
	TrafficPatterns *TrafficPatternsResult `json:"traffic_patterns,omitempty"`

	// Security Analysis Results
	SecurityAnalysis *SecurityAnalysisResult `json:"security_analysis,omitempty"`

	// Performance Analysis Results
	Performance *NetworkPerformanceResult `json:"performance,omitempty"`

	// Summary Statistics
	Summary *NetworkAnalysisSummary `json:"summary"`
}

// PacketAnalysisResult represents deep packet inspection results
type PacketAnalysisResult struct {
	TotalPackets           int64            `json:"total_packets"`
	AnalyzedPackets        int64            `json:"analyzed_packets"`
	ProtocolDistribution   map[string]int64 `json:"protocol_distribution"`
	PacketSizeDistribution map[string]int64 `json:"packet_size_distribution"`
	TopSourceIPs           []IPTrafficStat  `json:"top_source_ips"`
	TopDestinationIPs      []IPTrafficStat  `json:"top_destination_ips"`
	PortUsage              map[string]int64 `json:"port_usage"`
	Anomalies              []PacketAnomaly  `json:"anomalies"`
}

// TrafficPatternsResult represents traffic pattern analysis results
type TrafficPatternsResult struct {
	PatternID         string                    `json:"pattern_id"`
	DetectedPatterns  []TrafficPattern          `json:"detected_patterns"`
	BandwidthPatterns []BandwidthPattern        `json:"bandwidth_patterns"`
	TemporalPatterns  []TemporalPattern         `json:"temporal_patterns"`
	ClientBehavior    map[string]ClientBehavior `json:"client_behavior"`
	Anomalies         []TrafficAnomaly          `json:"anomalies"`
	PredictedTrends   []TrafficTrend            `json:"predicted_trends"`
}

// SecurityAnalysisResult represents security analysis results
type SecurityAnalysisResult struct {
	ThreatLevel        string               `json:"threat_level"` // LOW, MEDIUM, HIGH, CRITICAL
	DetectedThreats    []SecurityThreat     `json:"detected_threats"`
	SuspiciousActivity []SuspiciousActivity `json:"suspicious_activity"`
	BlockedConnections []BlockedConnection  `json:"blocked_connections"`
	DNSAnomalies       []DNSAnomaly         `json:"dns_anomalies"`
	PortScans          []PortScanEvent      `json:"port_scans"`
	TunnelingAttempts  []TunnelingAttempt   `json:"tunneling_attempts"`
}

// NetworkPerformanceResult represents network performance analysis results
type NetworkPerformanceResult struct {
	OverallScore      float64           `json:"overall_score"` // 0-100
	LatencyMetrics    LatencyMetrics    `json:"latency_metrics"`
	BandwidthMetrics  BandwidthMetrics  `json:"bandwidth_metrics"`
	ThroughputMetrics ThroughputMetrics `json:"throughput_metrics"`
	PacketLossMetrics PacketLossMetrics `json:"packet_loss_metrics"`
	JitterMetrics     JitterMetrics     `json:"jitter_metrics"`
	QualityAssessment QualityAssessment `json:"quality_assessment"`
}

// Supporting types for detailed analysis

// IPTrafficStat represents traffic statistics for an IP address
type IPTrafficStat struct {
	IP          string  `json:"ip"`
	Hostname    string  `json:"hostname,omitempty"`
	PacketCount int64   `json:"packet_count"`
	ByteCount   int64   `json:"byte_count"`
	Percentage  float64 `json:"percentage"`
}

// PacketAnomaly represents an anomalous packet or pattern
type PacketAnomaly struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"`
	Timestamp   string  `json:"timestamp"`
	SourceIP    string  `json:"source_ip"`
	DestIP      string  `json:"dest_ip"`
	Protocol    string  `json:"protocol"`
	Confidence  float64 `json:"confidence"`
}

// TrafficPattern represents a detected traffic pattern
type TrafficPattern struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"`
	Description     string                 `json:"description"`
	Confidence      float64                `json:"confidence"`
	StartTime       string                 `json:"start_time"`
	EndTime         string                 `json:"end_time"`
	Frequency       float64                `json:"frequency"`
	Characteristics map[string]interface{} `json:"characteristics"`
}

// BandwidthPattern represents bandwidth usage patterns
type BandwidthPattern struct {
	TimeSlot      string  `json:"time_slot"`
	AvgBandwidth  float64 `json:"avg_bandwidth_mbps"`
	PeakBandwidth float64 `json:"peak_bandwidth_mbps"`
	Usage         float64 `json:"usage_percentage"`
	Trend         string  `json:"trend"` // increasing, decreasing, stable
}

// TemporalPattern represents time-based patterns
type TemporalPattern struct {
	Pattern     string  `json:"pattern"` // hourly, daily, weekly
	PeakHours   []int   `json:"peak_hours"`
	LowHours    []int   `json:"low_hours"`
	Regularity  float64 `json:"regularity"` // 0-1 how regular the pattern is
	Seasonality bool    `json:"seasonality"`
}

// ClientBehavior represents individual client behavior patterns
type ClientBehavior struct {
	IP            string            `json:"ip"`
	Hostname      string            `json:"hostname,omitempty"`
	BehaviorType  string            `json:"behavior_type"`
	ActivityLevel string            `json:"activity_level"` // low, normal, high
	TypicalUsage  []HourlyUsage     `json:"typical_usage"`
	Anomalies     []BehaviorAnomaly `json:"anomalies"`
	RiskScore     float64           `json:"risk_score"`
}

// HourlyUsage represents usage during specific hours
type HourlyUsage struct {
	Hour         int     `json:"hour"` // 0-23
	AvgQueries   float64 `json:"avg_queries"`
	AvgBandwidth float64 `json:"avg_bandwidth_mbps"`
}

// BehaviorAnomaly represents anomalous client behavior
type BehaviorAnomaly struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Timestamp   string  `json:"timestamp"`
	Severity    string  `json:"severity"`
	Confidence  float64 `json:"confidence"`
}

// TrafficAnomaly represents traffic-level anomalies
type TrafficAnomaly struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Timestamp   string   `json:"timestamp"`
	Duration    string   `json:"duration"`
	Affected    []string `json:"affected_clients"`
	Severity    string   `json:"severity"`
	Confidence  float64  `json:"confidence"`
}

// TrafficTrend represents predicted traffic trends
type TrafficTrend struct {
	Metric      string  `json:"metric"`
	Current     float64 `json:"current_value"`
	Predicted   float64 `json:"predicted_value"`
	Confidence  float64 `json:"confidence"`
	TimeHorizon string  `json:"time_horizon"`
	Trend       string  `json:"trend"` // increasing, decreasing, stable
}

// Security-related types

// SecurityThreat represents a detected security threat
type SecurityThreat struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Severity    string            `json:"severity"`
	Description string            `json:"description"`
	SourceIP    string            `json:"source_ip"`
	TargetIP    string            `json:"target_ip,omitempty"`
	Timestamp   string            `json:"timestamp"`
	Evidence    map[string]string `json:"evidence"`
	Confidence  float64           `json:"confidence"`
	Mitigated   bool              `json:"mitigated"`
}

// SuspiciousActivity represents suspicious network activity
type SuspiciousActivity struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	SourceIP    string   `json:"source_ip"`
	Timestamp   string   `json:"timestamp"`
	Indicators  []string `json:"indicators"`
	RiskScore   float64  `json:"risk_score"`
}

// BlockedConnection represents a blocked network connection
type BlockedConnection struct {
	SourceIP  string `json:"source_ip"`
	DestIP    string `json:"dest_ip"`
	Port      int    `json:"port"`
	Protocol  string `json:"protocol"`
	Reason    string `json:"reason"`
	Timestamp string `json:"timestamp"`
	RuleID    string `json:"rule_id,omitempty"`
}

// DNSAnomaly represents DNS-specific anomalies
type DNSAnomaly struct {
	ID         string   `json:"id"`
	Type       string   `json:"type"`
	Domain     string   `json:"domain"`
	SourceIP   string   `json:"source_ip"`
	Timestamp  string   `json:"timestamp"`
	Indicators []string `json:"indicators"`
	Confidence float64  `json:"confidence"`
}

// PortScanEvent represents detected port scanning activity
type PortScanEvent struct {
	ID        string  `json:"id"`
	SourceIP  string  `json:"source_ip"`
	TargetIP  string  `json:"target_ip"`
	Ports     []int   `json:"scanned_ports"`
	StartTime string  `json:"start_time"`
	EndTime   string  `json:"end_time"`
	ScanType  string  `json:"scan_type"`
	Intensity float64 `json:"intensity"`
}

// TunnelingAttempt represents DNS tunneling attempts
type TunnelingAttempt struct {
	ID         string   `json:"id"`
	SourceIP   string   `json:"source_ip"`
	Domain     string   `json:"domain"`
	Timestamp  string   `json:"timestamp"`
	DataSize   int64    `json:"data_size_bytes"`
	Indicators []string `json:"indicators"`
	Confidence float64  `json:"confidence"`
}

// Performance-related types

// LatencyMetrics represents latency analysis results
type LatencyMetrics struct {
	AvgLatency   float64            `json:"avg_latency_ms"`
	MinLatency   float64            `json:"min_latency_ms"`
	MaxLatency   float64            `json:"max_latency_ms"`
	P50Latency   float64            `json:"p50_latency_ms"`
	P95Latency   float64            `json:"p95_latency_ms"`
	P99Latency   float64            `json:"p99_latency_ms"`
	PerClient    map[string]float64 `json:"per_client_latency"`
	Distribution []LatencyBucket    `json:"distribution"`
}

// LatencyBucket represents latency distribution buckets
type LatencyBucket struct {
	RangeStart float64 `json:"range_start_ms"`
	RangeEnd   float64 `json:"range_end_ms"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

// BandwidthMetrics represents bandwidth analysis results
type BandwidthMetrics struct {
	TotalBandwidth   float64             `json:"total_bandwidth_mbps"`
	AvgBandwidth     float64             `json:"avg_bandwidth_mbps"`
	PeakBandwidth    float64             `json:"peak_bandwidth_mbps"`
	PerClient        map[string]float64  `json:"per_client_bandwidth"`
	TimeDistribution []BandwidthTimeSlot `json:"time_distribution"`
}

// BandwidthTimeSlot represents bandwidth usage over time
type BandwidthTimeSlot struct {
	TimeSlot  string  `json:"time_slot"`
	Bandwidth float64 `json:"bandwidth_mbps"`
}

// ThroughputMetrics represents throughput analysis results
type ThroughputMetrics struct {
	QueriesPerSecond float64 `json:"queries_per_second"`
	PeakQPS          float64 `json:"peak_qps"`
	AvgQPS           float64 `json:"avg_qps"`
	ResponseRate     float64 `json:"response_rate_percentage"`
	ProcessingTime   float64 `json:"avg_processing_time_ms"`
}

// PacketLossMetrics represents packet loss analysis results
type PacketLossMetrics struct {
	LossPercentage float64            `json:"loss_percentage"`
	TotalLost      int64              `json:"total_lost_packets"`
	TotalSent      int64              `json:"total_sent_packets"`
	PerClient      map[string]float64 `json:"per_client_loss"`
	BurstLoss      []LossBurst        `json:"burst_loss_events"`
}

// LossBurst represents burst packet loss events
type LossBurst struct {
	StartTime   string  `json:"start_time"`
	Duration    string  `json:"duration"`
	LostPackets int64   `json:"lost_packets"`
	LossRate    float64 `json:"loss_rate_percentage"`
}

// JitterMetrics represents jitter analysis results
type JitterMetrics struct {
	AvgJitter    float64            `json:"avg_jitter_ms"`
	MaxJitter    float64            `json:"max_jitter_ms"`
	JitterStdDev float64            `json:"jitter_std_dev"`
	PerClient    map[string]float64 `json:"per_client_jitter"`
}

// QualityAssessment represents overall network quality assessment
type QualityAssessment struct {
	OverallGrade     string         `json:"overall_grade"` // A, B, C, D, F
	LatencyGrade     string         `json:"latency_grade"`
	BandwidthGrade   string         `json:"bandwidth_grade"`
	ReliabilityGrade string         `json:"reliability_grade"`
	Recommendations  []string       `json:"recommendations"`
	Issues           []QualityIssue `json:"issues"`
}

// QualityIssue represents a network quality issue
type QualityIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Resolution  string `json:"suggested_resolution"`
}

// NetworkAnalysisSummary provides high-level summary of analysis
type NetworkAnalysisSummary struct {
	TotalClients      int      `json:"total_clients"`
	ActiveClients     int      `json:"active_clients"`
	TotalQueries      int64    `json:"total_queries"`
	AnomaliesDetected int      `json:"anomalies_detected"`
	ThreatLevel       string   `json:"threat_level"`
	OverallHealth     string   `json:"overall_health"`
	HealthScore       float64  `json:"health_score"` // 0-100
	KeyInsights       []string `json:"key_insights"`
}
