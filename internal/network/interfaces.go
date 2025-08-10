package network

import (
	"context"
	"time"

	"pihole-analyzer/internal/types"
)

// NetworkAnalyzer defines the interface for enhanced network analysis
type NetworkAnalyzer interface {
	// Initialize the analyzer with configuration
	Initialize(ctx context.Context, config types.NetworkAnalysisConfig) error

	// Analyze network traffic with Pi-hole records
	AnalyzeTraffic(ctx context.Context, records []types.PiholeRecord, clientStats map[string]*types.ClientStats) (*types.NetworkAnalysisResult, error)

	// Get analysis capabilities
	GetCapabilities() []string

	// Health check
	IsHealthy() bool
}

// DeepPacketInspector defines the interface for deep packet inspection
type DeepPacketInspector interface {
	// Analyze packets for the given time window
	InspectPackets(ctx context.Context, records []types.PiholeRecord, config types.DPIConfig) (*types.PacketAnalysisResult, error)

	// Analyze protocol distribution
	AnalyzeProtocols(records []types.PiholeRecord) (map[string]int64, error)

	// Detect packet anomalies
	DetectPacketAnomalies(ctx context.Context, records []types.PiholeRecord, baseline map[string]interface{}) ([]types.PacketAnomaly, error)

	// Get top traffic sources and destinations
	GetTrafficTopology(records []types.PiholeRecord) ([]types.IPTrafficStat, []types.IPTrafficStat, error)
}

// TrafficPatternAnalyzer defines the interface for traffic pattern analysis
type TrafficPatternAnalyzer interface {
	// Analyze traffic patterns
	AnalyzePatterns(ctx context.Context, records []types.PiholeRecord, clientStats map[string]*types.ClientStats, config types.TrafficPatternsConfig) (*types.TrafficPatternsResult, error)

	// Detect bandwidth patterns
	AnalyzeBandwidthPatterns(records []types.PiholeRecord, timeWindow time.Duration) ([]types.BandwidthPattern, error)

	// Analyze temporal patterns
	AnalyzeTemporalPatterns(records []types.PiholeRecord) ([]types.TemporalPattern, error)

	// Analyze client behavior patterns
	AnalyzeClientBehavior(clientStats map[string]*types.ClientStats, records []types.PiholeRecord) (map[string]types.ClientBehavior, error)

	// Detect traffic anomalies
	DetectTrafficAnomalies(ctx context.Context, records []types.PiholeRecord, config types.TrafficPatternsConfig) ([]types.TrafficAnomaly, error)

	// Predict traffic trends
	PredictTrafficTrends(records []types.PiholeRecord, horizon time.Duration) ([]types.TrafficTrend, error)
}

// SecurityAnalyzer defines the interface for security analysis
type SecurityAnalyzer interface {
	// Perform comprehensive security analysis
	AnalyzeSecurity(ctx context.Context, records []types.PiholeRecord, clientStats map[string]*types.ClientStats, config types.SecurityAnalysisConfig) (*types.SecurityAnalysisResult, error)

	// Detect security threats
	DetectThreats(ctx context.Context, records []types.PiholeRecord, config types.SecurityAnalysisConfig) ([]types.SecurityThreat, error)

	// Analyze suspicious activity
	AnalyzeSuspiciousActivity(records []types.PiholeRecord, clientStats map[string]*types.ClientStats) ([]types.SuspiciousActivity, error)

	// Detect DNS anomalies
	DetectDNSAnomalies(ctx context.Context, records []types.PiholeRecord) ([]types.DNSAnomaly, error)

	// Detect port scanning
	DetectPortScans(records []types.PiholeRecord) ([]types.PortScanEvent, error)

	// Detect DNS tunneling attempts
	DetectDNSTunneling(ctx context.Context, records []types.PiholeRecord) ([]types.TunnelingAttempt, error)

	// Assess threat level
	AssessThreatLevel(threats []types.SecurityThreat, suspicious []types.SuspiciousActivity) string
}

// PerformanceAnalyzer defines the interface for network performance analysis
type PerformanceAnalyzer interface {
	// Analyze network performance
	AnalyzePerformance(ctx context.Context, records []types.PiholeRecord, clientStats map[string]*types.ClientStats, config types.NetworkPerformanceConfig) (*types.NetworkPerformanceResult, error)

	// Analyze latency metrics
	AnalyzeLatency(records []types.PiholeRecord) (*types.LatencyMetrics, error)

	// Analyze bandwidth metrics
	AnalyzeBandwidth(records []types.PiholeRecord, clientStats map[string]*types.ClientStats) (*types.BandwidthMetrics, error)

	// Analyze throughput metrics
	AnalyzeThroughput(records []types.PiholeRecord) (*types.ThroughputMetrics, error)

	// Detect packet loss
	DetectPacketLoss(records []types.PiholeRecord) (*types.PacketLossMetrics, error)

	// Analyze jitter
	AnalyzeJitter(records []types.PiholeRecord) (*types.JitterMetrics, error)

	// Assess overall network quality
	AssessNetworkQuality(latency *types.LatencyMetrics, bandwidth *types.BandwidthMetrics,
		throughput *types.ThroughputMetrics, packetLoss *types.PacketLossMetrics,
		jitter *types.JitterMetrics, thresholds types.QualityThresholds) (*types.QualityAssessment, error)
}

// NetworkVisualizer defines the interface for network visualization data
type NetworkVisualizer interface {
	// Generate visualization data for traffic patterns
	GenerateTrafficVisualization(result *types.NetworkAnalysisResult) (map[string]interface{}, error)

	// Generate topology visualization data
	GenerateTopologyVisualization(records []types.PiholeRecord, clientStats map[string]*types.ClientStats) (map[string]interface{}, error)

	// Generate time-series data for trends
	GenerateTimeSeriesData(records []types.PiholeRecord, metric string, interval time.Duration) (map[string]interface{}, error)

	// Generate heatmap data for activity patterns
	GenerateHeatmapData(records []types.PiholeRecord) (map[string]interface{}, error)

	// Generate security dashboard data
	GenerateSecurityDashboard(securityResult *types.SecurityAnalysisResult) (map[string]interface{}, error)

	// Generate performance dashboard data
	GeneratePerformanceDashboard(performanceResult *types.NetworkPerformanceResult) (map[string]interface{}, error)
}

// AnalyzerFactory creates network analyzers
type AnalyzerFactory interface {
	// Create a complete network analyzer
	CreateNetworkAnalyzer(config types.NetworkAnalysisConfig) (NetworkAnalyzer, error)

	// Create individual analyzers
	CreateDPIAnalyzer(config types.DPIConfig) (DeepPacketInspector, error)
	CreateTrafficAnalyzer(config types.TrafficPatternsConfig) (TrafficPatternAnalyzer, error)
	CreateSecurityAnalyzer(config types.SecurityAnalysisConfig) (SecurityAnalyzer, error)
	CreatePerformanceAnalyzer(config types.NetworkPerformanceConfig) (PerformanceAnalyzer, error)
	CreateVisualizer() (NetworkVisualizer, error)
}

// AnalysisContext provides context for network analysis operations
type AnalysisContext struct {
	StartTime  time.Time
	EndTime    time.Time
	AnalysisID string
	Config     types.NetworkAnalysisConfig
	Metadata   map[string]interface{}
}

// AnalysisFilter provides filtering options for analysis
type AnalysisFilter struct {
	ClientIPs       []string
	Domains         []string
	Protocols       []string
	TimeRange       *TimeRange
	ExcludeInternal bool
	IncludeBlocked  bool
}

// TimeRange represents a time range for analysis
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// AnalysisOptions provides options for analysis behavior
type AnalysisOptions struct {
	EnableCaching      bool
	ParallelProcessing bool
	MaxConcurrency     int
	Timeout            time.Duration
	SamplingRate       float64
	DetailLevel        AnalysisDetailLevel
}

// AnalysisDetailLevel represents the level of detail for analysis
type AnalysisDetailLevel string

const (
	DetailLevelBasic         AnalysisDetailLevel = "basic"
	DetailLevelStandard      AnalysisDetailLevel = "standard"
	DetailLevelDetailed      AnalysisDetailLevel = "detailed"
	DetailLevelComprehensive AnalysisDetailLevel = "comprehensive"
)

// MetricsProvider provides metrics for monitoring
type MetricsProvider interface {
	// Record analysis metrics
	RecordAnalysisMetrics(analysisType string, duration time.Duration, recordCount int)

	// Record anomaly detection metrics
	RecordAnomalyMetrics(anomalyType string, count int, confidence float64)

	// Record performance metrics
	RecordPerformanceMetrics(metricType string, value float64)

	// Get analysis statistics
	GetAnalysisStats() map[string]interface{}
}
