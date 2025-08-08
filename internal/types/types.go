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
	OnlineOnly bool            `json:"online_only"`
	NoExclude  bool            `json:"no_exclude"`
	TestMode   bool            `json:"test_mode"`
	Quiet      bool            `json:"quiet"`
	Pihole     PiholeConfig    `json:"pihole"`
	Output     OutputConfig    `json:"output"`
	Exclusions ExclusionConfig `json:"exclusions"`
	Logging    LoggingConfig   `json:"logging"`
}

// PiholeConfig represents Pi-hole specific configuration
type PiholeConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	DatabasePath string `json:"database_path"`
	UseSSH       bool   `json:"use_ssh"`
	SSHKeyPath   string `json:"ssh_key_path"`
	KeyFile      string `json:"key_file"`
	DBPath       string `json:"db_path"`
	KeyPath      string `json:"key_path"` // SSH key path

	// API Configuration
	APIEnabled  bool   `json:"api_enabled"`
	APIPassword string `json:"api_password"`
	APITOTP     string `json:"api_totp"`
	UseHTTPS    bool   `json:"use_https"`
	APITimeout  int    `json:"api_timeout"`

	// Migration Configuration (Phase 4)
	MigrationMode string `json:"migration_mode"` // "ssh-only", "api-first", "api-only-warn", "api-only", "auto"
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

// NetworkDevice represents a network device from Pi-hole API (Enhanced for Phase 5)
type NetworkDevice struct {
	IP          string `json:"ip"`
	Hardware    string `json:"hardware"`
	Name        string `json:"name"`
	FirstSeen   string `json:"first_seen"`
	LastSeen    string `json:"last_seen"`
	VendorClass string `json:"vendor_class"`

	// Phase 5 enhancements
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

// Phase 5: Enhanced types for complete SSH replacement and analyzer integration

// AnalysisResult represents the comprehensive result of Phase 5 analysis
type AnalysisResult struct {
	ClientStats     map[string]*ClientStats `json:"client_stats"`
	NetworkDevices  []NetworkDevice         `json:"network_devices"`
	TotalQueries    int                     `json:"total_queries"`
	UniqueClients   int                     `json:"unique_clients"`
	AnalysisMode    string                  `json:"analysis_mode"`
	DataSourceType  string                  `json:"data_source_type"`
	Timestamp       string                  `json:"timestamp"`
	Performance     *QueryPerformance       `json:"performance,omitempty"`
	MigrationStatus string                  `json:"migration_status,omitempty"`
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

// MigrationReport represents the result of migration analysis
type MigrationReport struct {
	CurrentMode     string   `json:"current_mode"`
	RecommendedMode string   `json:"recommended_mode"`
	Issues          []string `json:"issues"`
	Warnings        []string `json:"warnings"`
	NextSteps       []string `json:"next_steps"`
	ReadinessScore  int      `json:"readiness_score"`
}
