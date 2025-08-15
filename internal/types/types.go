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
	Integrations    IntegrationsConfig    `json:"integrations"`
	Alerts          AlertConfig           `json:"alerts"`
	DNS             DNSConfig             `json:"dns"`
	DHCP            DHCPConfig            `json:"dhcp"`
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

// AlertConfig represents configuration for the alert system
type AlertConfig struct {
	Enabled       bool                   `json:"enabled"`
	Rules         []AlertRule            `json:"rules"`
	Notifications NotificationConfig     `json:"notifications"`
	Storage       StorageConfig          `json:"storage"`
	Performance   AlertPerformanceConfig `json:"performance"`

	// Default settings
	DefaultSeverity string `json:"default_severity"`
	DefaultCooldown string `json:"default_cooldown"`
	MaxActiveAlerts int    `json:"max_active_alerts"`
	AlertRetention  string `json:"alert_retention"`
}

// AlertRule defines criteria for triggering alerts
type AlertRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Type        string `json:"type"`
	Severity    string `json:"severity"`

	// Rule conditions
	Conditions []AlertCondition `json:"conditions"`

	// Suppression settings
	CooldownPeriod string `json:"cooldown_period"`
	MaxAlerts      int    `json:"max_alerts"`

	// Notification settings
	Channels   []string `json:"channels"`
	Recipients []string `json:"recipients"`

	// Tags for organization
	Tags []string `json:"tags"`
}

// AlertCondition defines a condition that must be met to trigger an alert
type AlertCondition struct {
	Field      string      `json:"field"`                 // e.g., "query_count", "anomaly_score", "response_time"
	Operator   string      `json:"operator"`              // e.g., "gt", "lt", "eq", "contains"
	Value      interface{} `json:"value"`                 // Threshold value
	TimeWindow string      `json:"time_window,omitempty"` // Duration string
}

// NotificationConfig configures notification channels
type NotificationConfig struct {
	Slack SlackNotificationConfig `json:"slack"`
	Email EmailNotificationConfig `json:"email"`
}

// SlackNotificationConfig configures Slack notifications
type SlackNotificationConfig struct {
	Enabled    bool   `json:"enabled"`
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel"`
	Username   string `json:"username"`
	IconEmoji  string `json:"icon_emoji"`
	Timeout    string `json:"timeout"`
}

// EmailNotificationConfig configures email notifications
type EmailNotificationConfig struct {
	Enabled    bool     `json:"enabled"`
	SMTPHost   string   `json:"smtp_host"`
	SMTPPort   int      `json:"smtp_port"`
	Username   string   `json:"username"`
	Password   string   `json:"password"`
	From       string   `json:"from"`
	Recipients []string `json:"recipients"`
	UseTLS     bool     `json:"use_tls"`
	Timeout    string   `json:"timeout"`
}

// StorageConfig configures alert storage
type StorageConfig struct {
	Type      string `json:"type"` // "memory", "file", "database"
	Path      string `json:"path,omitempty"`
	MaxSize   int    `json:"max_size"` // Maximum number of alerts to store
	Retention string `json:"retention"`
}

// AlertPerformanceConfig configures performance settings
type AlertPerformanceConfig struct {
	MaxConcurrentNotifications int    `json:"max_concurrent_notifications"`
	NotificationTimeout        string `json:"notification_timeout"`
	BatchSize                  int    `json:"batch_size"`
	EvaluationInterval         string `json:"evaluation_interval"`
}

// IntegrationsConfig represents configuration for external integrations
type IntegrationsConfig struct {
	Enabled    bool                       `json:"enabled"`
	Grafana    GrafanaConfig              `json:"grafana"`
	Loki       LokiConfig                 `json:"loki"`
	Prometheus PrometheusExtConfig        `json:"prometheus"`
	Generic    []GenericIntegrationConfig `json:"generic"`
}

// GrafanaConfig represents Grafana integration configuration
type GrafanaConfig struct {
	Enabled      bool   `json:"enabled"`
	URL          string `json:"url"`
	APIKey       string `json:"api_key"`
	Organization string `json:"organization"`

	// Data source configuration
	DataSource DataSourceConfig `json:"data_source"`

	// Dashboard management
	Dashboards DashboardConfig `json:"dashboards"`

	// Alert management
	AlertIntegration GrafanaAlertConfig `json:"alert_integration"`

	// Connection settings
	Timeout    int  `json:"timeout_seconds"`
	VerifyTLS  bool `json:"verify_tls"`
	RetryCount int  `json:"retry_count"`
}

// DataSourceConfig configures Grafana data source settings
type DataSourceConfig struct {
	CreateIfNotExists bool   `json:"create_if_not_exists"`
	Name              string `json:"name"`
	Type              string `json:"type"` // prometheus, loki, etc.
	URL               string `json:"url"`
	Access            string `json:"access"` // proxy, direct
	BasicAuth         bool   `json:"basic_auth"`
	Username          string `json:"username"`
	Password          string `json:"password"`
}

// DashboardConfig configures Grafana dashboard management
type DashboardConfig struct {
	AutoProvision     bool     `json:"auto_provision"`
	FolderName        string   `json:"folder_name"`
	DashboardFiles    []string `json:"dashboard_files"`
	OverwriteExisting bool     `json:"overwrite_existing"`
	Tags              []string `json:"tags"`
}

// GrafanaAlertConfig configures Grafana alerting (renamed to avoid conflict)
type GrafanaAlertConfig struct {
	Enabled              bool              `json:"enabled"`
	NotificationChannels []string          `json:"notification_channels"`
	Rules                []AlertRuleConfig `json:"rules"`
	DefaultSeverity      string            `json:"default_severity"`
}

// AlertRuleConfig represents an alert rule configuration
type AlertRuleConfig struct {
	Name        string            `json:"name"`
	Query       string            `json:"query"`
	Condition   string            `json:"condition"`
	Threshold   float64           `json:"threshold"`
	Duration    string            `json:"duration"`
	Severity    string            `json:"severity"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// LokiConfig represents Loki integration configuration
type LokiConfig struct {
	Enabled  bool   `json:"enabled"`
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
	TenantID string `json:"tenant_id"`

	// Log shipping configuration
	BatchSize    int    `json:"batch_size"`
	BatchTimeout string `json:"batch_timeout"` // duration string
	BufferSize   int    `json:"buffer_size"`

	// Label configuration
	StaticLabels  map[string]string `json:"static_labels"`
	DynamicLabels []string          `json:"dynamic_labels"`

	// Connection settings
	Timeout       int    `json:"timeout_seconds"`
	VerifyTLS     bool   `json:"verify_tls"`
	RetryCount    int    `json:"retry_count"`
	RetryInterval string `json:"retry_interval"` // duration string
}

// PrometheusExtConfig represents extended Prometheus integration configuration
type PrometheusExtConfig struct {
	Enabled          bool                   `json:"enabled"`
	PushGateway      PushGatewayConfig      `json:"push_gateway"`
	RemoteWrite      RemoteWriteConfig      `json:"remote_write"`
	ServiceDiscovery ServiceDiscoveryConfig `json:"service_discovery"`
	ExternalLabels   map[string]string      `json:"external_labels"`
}

// PushGatewayConfig configures Prometheus push gateway integration
type PushGatewayConfig struct {
	Enabled  bool   `json:"enabled"`
	URL      string `json:"url"`
	Job      string `json:"job"`
	Instance string `json:"instance"`
	Username string `json:"username"`
	Password string `json:"password"`
	Timeout  int    `json:"timeout_seconds"`
	Interval string `json:"push_interval"` // duration string
}

// RemoteWriteConfig configures Prometheus remote write
type RemoteWriteConfig struct {
	Enabled   bool              `json:"enabled"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	Username  string            `json:"username"`
	Password  string            `json:"password"`
	Timeout   int               `json:"timeout_seconds"`
	BatchSize int               `json:"batch_size"`
}

// ServiceDiscoveryConfig configures service discovery for Prometheus
type ServiceDiscoveryConfig struct {
	Enabled         bool              `json:"enabled"`
	Type            string            `json:"type"` // consul, k8s, static
	Endpoints       []string          `json:"endpoints"`
	RefreshInterval string            `json:"refresh_interval"` // duration string
	Labels          map[string]string `json:"labels"`
}

// GenericIntegrationConfig represents a generic monitoring platform integration
type GenericIntegrationConfig struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Enabled  bool                   `json:"enabled"`
	URL      string                 `json:"url"`
	Headers  map[string]string      `json:"headers"`
	Auth     AuthConfig             `json:"auth"`
	Settings map[string]interface{} `json:"settings"`
	Timeout  int                    `json:"timeout_seconds"`
}

// AuthConfig represents authentication configuration for integrations
type AuthConfig struct {
	Type     string `json:"type"` // basic, bearer, api_key, oauth2
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
	APIKey   string `json:"api_key"`
	Header   string `json:"header"` // header name for api_key
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

// DNS Server Configuration

// DNSConfig represents DNS server configuration
type DNSConfig struct {
	// Server settings
	Enabled    bool   `json:"enabled"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	TCPEnabled bool   `json:"tcp_enabled"`
	UDPEnabled bool   `json:"udp_enabled"`

	// Timeouts (in seconds)
	ReadTimeout  int `json:"read_timeout"`
	WriteTimeout int `json:"write_timeout"`
	IdleTimeout  int `json:"idle_timeout"`

	// Cache configuration
	Cache DNSCacheConfig `json:"cache"`

	// Forwarder configuration
	Forwarder DNSForwarderConfig `json:"forwarder"`

	// Logging
	LogQueries bool `json:"log_queries"`
	LogLevel   int  `json:"log_level"`

	// Performance
	MaxConcurrentQueries int `json:"max_concurrent_queries"`
	BufferSize           int `json:"buffer_size"`
}

// DNSCacheConfig represents DNS cache configuration
type DNSCacheConfig struct {
	Enabled         bool `json:"enabled"`
	MaxSize         int  `json:"max_size"`
	DefaultTTL      int  `json:"default_ttl"`      // seconds
	MaxTTL          int  `json:"max_ttl"`          // seconds
	MinTTL          int  `json:"min_ttl"`          // seconds
	CleanupInterval int  `json:"cleanup_interval"` // seconds

	// Cache strategy: "lru", "lfu", "ttl"
	EvictionPolicy string `json:"eviction_policy"`

	// Memory limits
	MaxMemoryMB int `json:"max_memory_mb"`
}

// DNSForwarderConfig represents DNS forwarder configuration
type DNSForwarderConfig struct {
	Enabled        bool     `json:"enabled"`
	Upstreams      []string `json:"upstreams"`
	Timeout        int      `json:"timeout"` // seconds
	Retries        int      `json:"retries"`
	HealthCheck    bool     `json:"health_check"`
	HealthInterval int      `json:"health_interval"` // seconds

	// Load balancing: "round_robin", "random", "fastest"
	LoadBalancing string `json:"load_balancing"`

	// EDNS0 support
	EDNS0Enabled bool `json:"edns0_enabled"`
	UDPSize      int  `json:"udp_size"`
}

// DHCP Server Configuration and Types

// DHCPConfig represents configuration for the DHCP server
type DHCPConfig struct {
	Enabled       bool               `json:"enabled"`
	Interface     string             `json:"interface"`      // Network interface to bind to
	ListenAddress string             `json:"listen_address"` // IP address to listen on
	Port          int                `json:"port"`           // DHCP server port (default: 67)
	Pool          DHCPPoolConfig     `json:"pool"`           // IP address pool configuration
	LeaseTime     string             `json:"lease_time"`     // Default lease duration (e.g., "24h")
	MaxLeaseTime  string             `json:"max_lease_time"` // Maximum lease duration
	RenewalTime   string             `json:"renewal_time"`   // T1 renewal time
	RebindTime    string             `json:"rebind_time"`    // T2 rebind time
	Options       DHCPOptionsConfig  `json:"options"`        // DHCP options configuration
	Reservations  []DHCPReservation  `json:"reservations"`   // Static IP reservations
	Storage       DHCPStorageConfig  `json:"storage"`        // Lease storage configuration
	Performance   DHCPPerfConfig     `json:"performance"`    // Performance settings
	Security      DHCPSecurityConfig `json:"security"`       // Security settings
}

// DHCPPoolConfig configures the IP address pool
type DHCPPoolConfig struct {
	StartIP    string   `json:"start_ip"`    // Pool start IP address
	EndIP      string   `json:"end_ip"`      // Pool end IP address
	Subnet     string   `json:"subnet"`      // Subnet in CIDR notation (e.g., "192.168.1.0/24")
	Gateway    string   `json:"gateway"`     // Default gateway IP
	DNSServers []string `json:"dns_servers"` // DNS server IP addresses
	Exclude    []string `json:"exclude"`     // IP addresses to exclude from pool
}

// DHCPOptionsConfig configures DHCP options
type DHCPOptionsConfig struct {
	Router             string         `json:"router"`               // Option 3: Router
	DomainName         string         `json:"domain_name"`          // Option 15: Domain Name
	DomainNameServer   []string       `json:"domain_name_server"`   // Option 6: DNS Servers
	NetBIOSNameServers []string       `json:"netbios_name_servers"` // Option 44: NetBIOS Name Servers
	NTPServers         []string       `json:"ntp_servers"`          // Option 42: NTP Servers
	TFTPServer         string         `json:"tftp_server"`          // Option 66: TFTP Server
	BootFileName       string         `json:"boot_file_name"`       // Option 67: Boot File Name
	MTU                int            `json:"mtu"`                  // Option 26: MTU
	CustomOptions      map[int]string `json:"custom_options"`       // Custom DHCP options
}

// DHCPReservation represents a static IP reservation
type DHCPReservation struct {
	MAC         string         `json:"mac"`         // MAC address
	IP          string         `json:"ip"`          // Reserved IP address
	Hostname    string         `json:"hostname"`    // Optional hostname
	Description string         `json:"description"` // Optional description
	Options     map[int]string `json:"options"`     // Custom options for this reservation
	Enabled     bool           `json:"enabled"`     // Whether reservation is active
}

// DHCPStorageConfig configures lease storage
type DHCPStorageConfig struct {
	Type         string `json:"type"`          // "memory", "file", "database"
	Path         string `json:"path"`          // File path or database connection string
	BackupPath   string `json:"backup_path"`   // Backup file path
	SyncInterval string `json:"sync_interval"` // How often to sync to storage
	MaxLeases    int    `json:"max_leases"`    // Maximum number of leases to store
}

// DHCPPerfConfig configures performance settings
type DHCPPerfConfig struct {
	MaxConnections       int    `json:"max_connections"`        // Maximum concurrent connections
	ReadTimeout          string `json:"read_timeout"`           // Read timeout for packets
	WriteTimeout         string `json:"write_timeout"`          // Write timeout for packets
	BufferSize           int    `json:"buffer_size"`            // UDP buffer size
	WorkerPoolSize       int    `json:"worker_pool_size"`       // Number of worker goroutines
	LeaseCleanupInterval string `json:"lease_cleanup_interval"` // How often to clean expired leases
}

// DHCPSecurityConfig configures security settings
type DHCPSecurityConfig struct {
	EnableRateLimit      bool     `json:"enable_rate_limit"`     // Enable rate limiting
	MaxRequestsPerIP     int      `json:"max_requests_per_ip"`   // Max requests per IP per minute
	AllowedClients       []string `json:"allowed_clients"`       // Allowed client MAC addresses (if set, only these can get leases)
	BlockedClients       []string `json:"blocked_clients"`       // Blocked client MAC addresses
	RequireClientID      bool     `json:"require_client_id"`     // Require DHCP client identifier
	LogAllRequests       bool     `json:"log_all_requests"`      // Log all DHCP requests
	EnableFingerprinting bool     `json:"enable_fingerprinting"` // Enable device fingerprinting
}

// DHCP Lease and Runtime Types

// DHCPLease represents an active or historical DHCP lease
type DHCPLease struct {
	ID               string            `json:"id"`                // Unique lease identifier
	IP               string            `json:"ip"`                // Assigned IP address
	MAC              string            `json:"mac"`               // Client MAC address
	Hostname         string            `json:"hostname"`          // Client hostname (if provided)
	ClientID         string            `json:"client_id"`         // DHCP client identifier
	VendorClass      string            `json:"vendor_class"`      // Vendor class identifier
	UserClass        string            `json:"user_class"`        // User class
	StartTime        string            `json:"start_time"`        // Lease start time (RFC3339)
	EndTime          string            `json:"end_time"`          // Lease expiry time (RFC3339)
	LastRenewal      string            `json:"last_renewal"`      // Last renewal time (RFC3339)
	State            DHCPLeaseState    `json:"state"`             // Current lease state
	Type             DHCPLeaseType     `json:"type"`              // Lease type (dynamic, static, etc.)
	Options          map[int]string    `json:"options"`           // DHCP options sent to client
	RequestedOptions []int             `json:"requested_options"` // Options requested by client
	Fingerprint      string            `json:"fingerprint"`       // Device fingerprint
	Metadata         map[string]string `json:"metadata"`          // Additional metadata
}

// DHCPLeaseState represents the state of a DHCP lease
type DHCPLeaseState string

const (
	LeaseStateOffered  DHCPLeaseState = "offered"  // DHCP Offer sent, waiting for Request
	LeaseStateActive   DHCPLeaseState = "active"   // Lease is active and in use
	LeaseStateExpired  DHCPLeaseState = "expired"  // Lease has expired
	LeaseStateReleased DHCPLeaseState = "released" // Client released the lease
	LeaseStateDeclined DHCPLeaseState = "declined" // Client declined the lease
	LeaseStateReserved DHCPLeaseState = "reserved" // Static reservation
)

// DHCPLeaseType represents the type of DHCP lease
type DHCPLeaseType string

const (
	LeaseTypeDynamic DHCPLeaseType = "dynamic" // Dynamically allocated
	LeaseTypeStatic  DHCPLeaseType = "static"  // Static reservation
	LeaseTypeBootP   DHCPLeaseType = "bootp"   // BootP allocation
)

// DHCPRequest represents a DHCP request from a client
type DHCPRequest struct {
	MessageType      int            `json:"message_type"`                // DHCP message type (1=DISCOVER, 3=REQUEST, etc.)
	TransactionID    uint32         `json:"transaction_id"`              // DHCP transaction ID
	ClientMAC        string         `json:"client_mac"`                  // Client MAC address
	ClientIP         string         `json:"client_ip,omitempty"`         // Client IP (for DHCP REQUEST)
	RequestedIP      string         `json:"requested_ip,omitempty"`      // Requested IP address
	ServerIdentifier string         `json:"server_identifier,omitempty"` // DHCP server identifier
	Options          map[int]string `json:"options"`                     // DHCP options from client
	RequestedOptions []int          `json:"requested_options"`           // Parameter request list
	Timestamp        string         `json:"timestamp"`                   // Request timestamp
	ClientHostname   string         `json:"client_hostname,omitempty"`   // Client hostname
	VendorClass      string         `json:"vendor_class,omitempty"`      // Vendor class identifier
	ClientID         string         `json:"client_id,omitempty"`         // Client identifier
}

// DHCPResponse represents a DHCP response to a client
type DHCPResponse struct {
	MessageType   int            `json:"message_type"`   // DHCP message type (2=OFFER, 5=ACK, etc.)
	TransactionID uint32         `json:"transaction_id"` // DHCP transaction ID
	ClientMAC     string         `json:"client_mac"`     // Client MAC address
	YourIP        string         `json:"your_ip"`        // Assigned IP address
	ServerIP      string         `json:"server_ip"`      // DHCP server IP
	Options       map[int]string `json:"options"`        // DHCP options sent to client
	LeaseTime     uint32         `json:"lease_time"`     // Lease time in seconds
	Timestamp     string         `json:"timestamp"`      // Response timestamp
}

// DHCPStatistics represents DHCP server statistics
type DHCPStatistics struct {
	TotalRequests    int64            `json:"total_requests"`   // Total DHCP requests processed
	TotalOffers      int64            `json:"total_offers"`     // Total DHCP offers sent
	TotalAcks        int64            `json:"total_acks"`       // Total DHCP ACKs sent
	TotalNaks        int64            `json:"total_naks"`       // Total DHCP NAKs sent
	TotalDeclines    int64            `json:"total_declines"`   // Total DHCP declines received
	TotalReleases    int64            `json:"total_releases"`   // Total DHCP releases received
	TotalInforms     int64            `json:"total_informs"`    // Total DHCP informs received
	ActiveLeases     int              `json:"active_leases"`    // Number of active leases
	AvailableIPs     int              `json:"available_ips"`    // Number of available IP addresses
	PoolUtilization  float64          `json:"pool_utilization"` // Pool utilization percentage
	AverageLeaseTime float64          `json:"avg_lease_time"`   // Average lease time in hours
	RequestsByType   map[string]int64 `json:"requests_by_type"` // Requests broken down by message type
	RequestsByHour   map[string]int64 `json:"requests_by_hour"` // Requests broken down by hour
	TopClients       []DHCPClientStat `json:"top_clients"`      // Top clients by request count
	RecentActivity   []DHCPActivity   `json:"recent_activity"`  // Recent DHCP activity
	ErrorCount       int64            `json:"error_count"`      // Total errors encountered
	Uptime           string           `json:"uptime"`           // Server uptime
}

// DHCPClientStat represents statistics for a DHCP client
type DHCPClientStat struct {
	MAC          string `json:"mac"`           // Client MAC address
	Hostname     string `json:"hostname"`      // Client hostname
	RequestCount int64  `json:"request_count"` // Number of requests from this client
	LastSeen     string `json:"last_seen"`     // Last seen timestamp
	CurrentIP    string `json:"current_ip"`    // Currently assigned IP
}

// DHCPActivity represents recent DHCP activity
type DHCPActivity struct {
	Timestamp   string `json:"timestamp"`   // Activity timestamp
	Type        string `json:"type"`        // Activity type (DISCOVER, REQUEST, etc.)
	ClientMAC   string `json:"client_mac"`  // Client MAC address
	IP          string `json:"ip"`          // IP address involved
	Description string `json:"description"` // Human-readable description
}

// DHCPServerStatus represents the current status of the DHCP server
type DHCPServerStatus struct {
	Running       bool           `json:"running"`        // Whether server is running
	StartTime     string         `json:"start_time"`     // Server start time
	ConfigValid   bool           `json:"config_valid"`   // Whether configuration is valid
	Interface     string         `json:"interface"`      // Network interface
	ListenAddress string         `json:"listen_address"` // Listen address
	PoolInfo      DHCPPoolInfo   `json:"pool_info"`      // Pool information
	Statistics    DHCPStatistics `json:"statistics"`     // Server statistics
	RecentErrors  []DHCPError    `json:"recent_errors"`  // Recent errors
	Version       string         `json:"version"`        // DHCP server version
}

// DHCPPoolInfo represents information about the IP address pool
type DHCPPoolInfo struct {
	StartIP         string  `json:"start_ip"`         // Pool start IP
	EndIP           string  `json:"end_ip"`           // Pool end IP
	TotalIPs        int     `json:"total_ips"`        // Total IP addresses in pool
	AllocatedIPs    int     `json:"allocated_ips"`    // Number of allocated IPs
	AvailableIPs    int     `json:"available_ips"`    // Number of available IPs
	ReservedIPs     int     `json:"reserved_ips"`     // Number of reserved IPs
	UtilizationRate float64 `json:"utilization_rate"` // Pool utilization percentage
}

// DHCPError represents a DHCP server error
type DHCPError struct {
	Timestamp string `json:"timestamp"`  // Error timestamp
	Type      string `json:"type"`       // Error type
	Message   string `json:"message"`    // Error message
	ClientMAC string `json:"client_mac"` // Client MAC (if applicable)
	Context   string `json:"context"`    // Additional context
}
