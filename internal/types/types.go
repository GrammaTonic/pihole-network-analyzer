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
	OnlineOnly   bool               `json:"online_only"`
	NoExclude    bool               `json:"no_exclude"`
	TestMode     bool               `json:"test_mode"`
	Quiet        bool               `json:"quiet"`
	Pihole       PiholeConfig       `json:"pihole"`
	Output       OutputConfig       `json:"output"`
	Exclusions   ExclusionConfig    `json:"exclusions"`
	Logging      LoggingConfig      `json:"logging"`
	Web          WebConfig          `json:"web"`
	Metrics      MetricsConfig      `json:"metrics"`
	ML           MLConfig           `json:"ml"`
	Integrations IntegrationsConfig `json:"integrations"`
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
	Alerts AlertConfig `json:"alerts"`

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

// AlertConfig configures Grafana alerting
type AlertConfig struct {
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
