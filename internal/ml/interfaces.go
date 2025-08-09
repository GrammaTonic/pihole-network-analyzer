package ml

import (
	"context"
	"time"

	"pihole-analyzer/internal/types"
)

// AnomalyDetector defines the interface for anomaly detection models
type AnomalyDetector interface {
	// Train the model with historical data
	Train(ctx context.Context, data []types.PiholeRecord) error

	// Detect anomalies in current data
	DetectAnomalies(ctx context.Context, data []types.PiholeRecord) ([]Anomaly, error)

	// Get model metadata
	GetModelInfo() ModelInfo

	// Check if model is trained
	IsTrained() bool
}

// TrendAnalyzer defines the interface for trend analysis
type TrendAnalyzer interface {
	// Analyze trends in DNS query patterns
	AnalyzeTrends(ctx context.Context, data []types.PiholeRecord, timeWindow time.Duration) (*TrendAnalysis, error)

	// Predict future query patterns
	PredictTrends(ctx context.Context, data []types.PiholeRecord, forecastWindow time.Duration) (*TrendPrediction, error)

	// Get trend metadata
	GetAnalyzerInfo() AnalyzerInfo
}

// MLEngine combines anomaly detection and trend analysis
type MLEngine interface {
	AnomalyDetector
	TrendAnalyzer

	// Initialize the ML engine
	Initialize(ctx context.Context, config MLConfig) error

	// Process data through both anomaly detection and trend analysis
	ProcessData(ctx context.Context, data []types.PiholeRecord) (*MLResults, error)

	// Get engine status
	GetStatus() EngineStatus
}

// Anomaly represents a detected anomaly
type Anomaly struct {
	ID          string                 `json:"id"`
	Type        AnomalyType            `json:"type"`
	Severity    SeverityLevel          `json:"severity"`
	Timestamp   time.Time              `json:"timestamp"`
	Description string                 `json:"description"`
	Score       float64                `json:"score"`      // Anomaly score (0-1)
	Confidence  float64                `json:"confidence"` // Model confidence (0-1)
	Metadata    map[string]interface{} `json:"metadata"`

	// Related data
	AffectedRecords []types.PiholeRecord `json:"affected_records"`
	ClientIP        string               `json:"client_ip,omitempty"`
	Domain          string               `json:"domain,omitempty"`
}

// AnomalyType defines types of anomalies
type AnomalyType string

const (
	AnomalyTypeVolumeSpike   AnomalyType = "volume_spike"
	AnomalyTypeVolumeDropout AnomalyType = "volume_dropout"
	AnomalyTypeUnusualDomain AnomalyType = "unusual_domain"
	AnomalyTypeUnusualClient AnomalyType = "unusual_client"
	AnomalyTypeQueryPattern  AnomalyType = "query_pattern"
	AnomalyTypeTimePattern   AnomalyType = "time_pattern"
	AnomalyTypeResponseTime  AnomalyType = "response_time"
	AnomalyTypeBlockedSpike  AnomalyType = "blocked_spike"
)

// SeverityLevel defines anomaly severity levels
type SeverityLevel string

const (
	SeverityLow      SeverityLevel = "low"
	SeverityMedium   SeverityLevel = "medium"
	SeverityHigh     SeverityLevel = "high"
	SeverityCritical SeverityLevel = "critical"
)

// TrendAnalysis represents trend analysis results
type TrendAnalysis struct {
	TimeWindow     time.Duration            `json:"time_window"`
	TotalQueries   int                      `json:"total_queries"`
	QueryTrend     TrendDirection           `json:"query_trend"`
	DomainTrends   []DomainTrend            `json:"domain_trends"`
	ClientTrends   []ClientTrend            `json:"client_trends"`
	HourlyPatterns map[int]float64          `json:"hourly_patterns"` // Hour -> normalized query count
	DailyPatterns  map[time.Weekday]float64 `json:"daily_patterns"`  // Day -> normalized query count
	Insights       []TrendInsight           `json:"insights"`
	Timestamp      time.Time                `json:"timestamp"`
}

// TrendPrediction represents future trend predictions
type TrendPrediction struct {
	ForecastWindow   time.Duration   `json:"forecast_window"`
	PredictedQueries []QueryForecast `json:"predicted_queries"`
	Confidence       float64         `json:"confidence"`
	Methodology      string          `json:"methodology"`
	Timestamp        time.Time       `json:"timestamp"`
}

// TrendDirection represents the direction of a trend
type TrendDirection string

const (
	TrendIncreasing TrendDirection = "increasing"
	TrendDecreasing TrendDirection = "decreasing"
	TrendStable     TrendDirection = "stable"
	TrendVolatile   TrendDirection = "volatile"
)

// DomainTrend represents trend data for a specific domain
type DomainTrend struct {
	Domain    string         `json:"domain"`
	Trend     TrendDirection `json:"trend"`
	Change    float64        `json:"change"` // Percentage change
	Queries   int            `json:"queries"`
	IsBlocked bool           `json:"is_blocked"`
}

// ClientTrend represents trend data for a specific client
type ClientTrend struct {
	ClientIP string         `json:"client_ip"`
	Trend    TrendDirection `json:"trend"`
	Change   float64        `json:"change"` // Percentage change
	Queries  int            `json:"queries"`
}

// TrendInsight represents an insight from trend analysis
type TrendInsight struct {
	Type        InsightType `json:"type"`
	Description string      `json:"description"`
	Impact      string      `json:"impact"`
	Confidence  float64     `json:"confidence"`
}

// InsightType defines types of insights
type InsightType string

const (
	InsightTypePeakUsage       InsightType = "peak_usage"
	InsightTypeOffPeakUsage    InsightType = "off_peak_usage"
	InsightTypeWeekendPattern  InsightType = "weekend_pattern"
	InsightTypeSeasonalPattern InsightType = "seasonal_pattern"
	InsightTypeAnomalousClient InsightType = "anomalous_client"
	InsightTypeEmergingThreat  InsightType = "emerging_threat"
)

// QueryForecast represents a predicted query count for a time period
type QueryForecast struct {
	Timestamp          time.Time          `json:"timestamp"`
	PredictedCount     int                `json:"predicted_count"`
	ConfidenceInterval ConfidenceInterval `json:"confidence_interval"`
}

// ConfidenceInterval represents the confidence bounds for predictions
type ConfidenceInterval struct {
	Lower float64 `json:"lower"`
	Upper float64 `json:"upper"`
}

// ModelInfo provides metadata about a trained model
type ModelInfo struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	TrainedAt    time.Time              `json:"trained_at"`
	TrainingSize int                    `json:"training_size"`
	Accuracy     float64                `json:"accuracy"`
	Parameters   map[string]interface{} `json:"parameters"`
	LastUpdated  time.Time              `json:"last_updated"`
}

// AnalyzerInfo provides metadata about a trend analyzer
type AnalyzerInfo struct {
	Name       string                 `json:"name"`
	Version    string                 `json:"version"`
	Algorithms []string               `json:"algorithms"`
	Parameters map[string]interface{} `json:"parameters"`
	LastRun    time.Time              `json:"last_run"`
}

// MLResults combines anomaly detection and trend analysis results
type MLResults struct {
	Anomalies     []Anomaly        `json:"anomalies"`
	TrendAnalysis *TrendAnalysis   `json:"trend_analysis"`
	Predictions   *TrendPrediction `json:"predictions"`
	ProcessedAt   time.Time        `json:"processed_at"`
	Summary       MLSummary        `json:"summary"`
}

// MLSummary provides a high-level summary of ML analysis
type MLSummary struct {
	TotalAnomalies    int            `json:"total_anomalies"`
	HighSeverityCount int            `json:"high_severity_count"`
	TrendDirection    TrendDirection `json:"trend_direction"`
	HealthScore       float64        `json:"health_score"` // 0-100
	Recommendations   []string       `json:"recommendations"`
}

// EngineStatus represents the status of the ML engine
type EngineStatus struct {
	IsInitialized bool      `json:"is_initialized"`
	IsTrained     bool      `json:"is_trained"`
	LastTraining  time.Time `json:"last_training"`
	LastAnalysis  time.Time `json:"last_analysis"`
	Status        string    `json:"status"`
	Errors        []string  `json:"errors"`
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
	WindowSize    time.Duration      `json:"window_size"`
	AnomalyTypes  []AnomalyType      `json:"anomaly_types"`
	Thresholds    map[string]float64 `json:"thresholds"`
}

// TrendAnalysisConfig configures trend analysis
type TrendAnalysisConfig struct {
	Enabled         bool          `json:"enabled"`
	AnalysisWindow  time.Duration `json:"analysis_window"`
	ForecastWindow  time.Duration `json:"forecast_window"`
	MinDataPoints   int           `json:"min_data_points"`
	SmoothingFactor float64       `json:"smoothing_factor"`
}

// TrainingConfig configures model training
type TrainingConfig struct {
	AutoRetrain     bool          `json:"auto_retrain"`
	RetrainInterval time.Duration `json:"retrain_interval"`
	MinTrainingSize int           `json:"min_training_size"`
	MaxTrainingSize int           `json:"max_training_size"`
	ValidationSplit float64       `json:"validation_split"`
}

// PerformanceConfig configures performance settings
type PerformanceConfig struct {
	MaxConcurrency  int           `json:"max_concurrency"`
	TimeoutDuration time.Duration `json:"timeout_duration"`
	CacheEnabled    bool          `json:"cache_enabled"`
	CacheDuration   time.Duration `json:"cache_duration"`
	BatchSize       int           `json:"batch_size"`
}

// DefaultMLConfig returns a default ML configuration
func DefaultMLConfig() MLConfig {
	return MLConfig{
		AnomalyDetection: AnomalyDetectionConfig{
			Enabled:       true,
			Sensitivity:   2.0,
			MinConfidence: 0.7,
			WindowSize:    24 * time.Hour,
			AnomalyTypes: []AnomalyType{
				AnomalyTypeVolumeSpike,
				AnomalyTypeUnusualDomain,
				AnomalyTypeUnusualClient,
				AnomalyTypeTimePattern,
			},
			Thresholds: map[string]float64{
				"zscore":    2.5,
				"volume":    100,
				"frequency": 10,
			},
		},
		TrendAnalysis: TrendAnalysisConfig{
			Enabled:         true,
			AnalysisWindow:  24 * time.Hour,
			ForecastWindow:  6 * time.Hour,
			MinDataPoints:   20,
			SmoothingFactor: 0.3,
		},
	}
}
