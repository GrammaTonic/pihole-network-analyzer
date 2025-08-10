package ml

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"pihole-analyzer/internal/logger"
	"pihole-analyzer/internal/types"
)

// Engine implements the MLEngine interface
type Engine struct {
	config MLConfig
	logger *logger.Logger
	mu     sync.RWMutex

	// Models
	anomalyDetector AnomalyDetector
	trendAnalyzer   TrendAnalyzer

	// State
	initialized  bool
	trained      bool
	lastTraining time.Time
	lastAnalysis time.Time
	status       string
	errors       []string
}

// NewEngine creates a new ML engine instance
func NewEngine(config MLConfig, parentLogger *slog.Logger) *Engine {
	return &Engine{
		config: config,
		logger: logger.New(&logger.Config{
			Component:     "ml-engine",
			Level:         logger.LevelInfo,
			EnableColors:  true,
			EnableEmojis:  true,
			ShowTimestamp: true,
		}),
		status: "created",
		errors: make([]string, 0),
	}
}

// Initialize sets up the ML engine with the provided configuration
func (e *Engine) Initialize(ctx context.Context, config MLConfig) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.logger.Info("Initializing ML engine: anomaly_detection_enabled=%t trend_analysis_enabled=%t",
		config.AnomalyDetection.Enabled, config.TrendAnalysis.Enabled)

	e.config = config

	// Initialize anomaly detector if enabled
	if config.AnomalyDetection.Enabled {
		slogger := slog.New(slog.NewTextHandler(os.Stderr, nil)).With(slog.String("component", "anomaly-detector"))
		detector := NewStatisticalAnomalyDetector(config.AnomalyDetection, slogger)
		if err := detector.Initialize(ctx, config.AnomalyDetection); err != nil {
			e.logger.Error("Failed to initialize anomaly detector: %v", err)
			return fmt.Errorf("failed to initialize anomaly detector: %w", err)
		}
		e.anomalyDetector = detector
	}

	// Initialize trend analyzer if enabled
	if config.TrendAnalysis.Enabled {
		slogger := slog.New(slog.NewTextHandler(os.Stderr, nil)).With(slog.String("component", "trend-analyzer"))
		e.trendAnalyzer = NewTrendAnalyzer(config.TrendAnalysis, slogger)
	}

	e.initialized = true
	e.status = "initialized"
	e.logger.Info("ML engine initialized successfully")

	return nil
}

// Train trains the ML models with historical data
func (e *Engine) Train(ctx context.Context, data []types.PiholeRecord) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.initialized {
		return fmt.Errorf("ML engine not initialized")
	}

	e.logger.Info("Training ML models with %d records", len(data))

	// Train anomaly detector if enabled
	if e.config.AnomalyDetection.Enabled && e.anomalyDetector != nil {
		e.logger.Debug("Training anomaly detector")
		err := e.anomalyDetector.Train(ctx, data)
		if err != nil {
			e.logger.Error("Failed to train anomaly detector: %v", err)
			return fmt.Errorf("failed to train anomaly detector: %w", err)
		}
		e.logger.Info("Anomaly detector training completed")
	}

	e.trained = true
	e.lastTraining = time.Now()
	e.status = "trained"
	e.logger.Info("ML engine training completed successfully")

	return nil
}

// DetectAnomalies implements the AnomalyDetector interface
func (e *Engine) DetectAnomalies(ctx context.Context, data []types.PiholeRecord) ([]Anomaly, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.initialized || e.anomalyDetector == nil {
		return nil, fmt.Errorf("anomaly detector not available")
	}

	e.logger.Debug("Detecting anomalies in %d records", len(data))
	anomalies, err := e.anomalyDetector.DetectAnomalies(ctx, data)
	if err != nil {
		e.logger.Error("Anomaly detection failed: %v", err)
		return nil, fmt.Errorf("anomaly detection failed: %w", err)
	}

	e.logger.Info("Detected %d anomalies", len(anomalies))
	return anomalies, nil
}

// AnalyzeTrends implements the TrendAnalyzer interface
func (e *Engine) AnalyzeTrends(ctx context.Context, data []types.PiholeRecord, timeWindow time.Duration) (*TrendAnalysis, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.initialized || e.trendAnalyzer == nil {
		return nil, fmt.Errorf("trend analyzer not available")
	}

	e.logger.Debug("Analyzing trends in %d records", len(data))
	analysis, err := e.trendAnalyzer.AnalyzeTrends(ctx, data, timeWindow)
	if err != nil {
		e.logger.Error("Trend analysis failed: %v", err)
		return nil, fmt.Errorf("trend analysis failed: %w", err)
	}

	e.logger.Info("Trend analysis completed with %d insights", len(analysis.Insights))
	return analysis, nil
}

// PredictTrends implements the TrendAnalyzer interface
func (e *Engine) PredictTrends(ctx context.Context, data []types.PiholeRecord, forecastWindow time.Duration) (*TrendPrediction, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.initialized || e.trendAnalyzer == nil {
		return nil, fmt.Errorf("trend analyzer not available")
	}

	e.logger.Debug("Predicting trends: record_count=%d", len(data))
	prediction, err := e.trendAnalyzer.PredictTrends(ctx, data, forecastWindow)
	if err != nil {
		e.logger.Error("Trend prediction failed: %v", err)
		return nil, fmt.Errorf("trend prediction failed: %w", err)
	}

	e.logger.Info("Trend prediction completed: predicted_points=%d", len(prediction.PredictedQueries))
	return prediction, nil
}

// GetModelInfo implements the AnomalyDetector interface
func (e *Engine) GetModelInfo() ModelInfo {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return ModelInfo{
		Name:         "ML Engine",
		Version:      "1.0.0",
		TrainedAt:    e.lastTraining,
		TrainingSize: 0,
		Accuracy:     0.0,
		Parameters:   make(map[string]interface{}),
		LastUpdated:  e.lastAnalysis,
	}
}

// ProcessData processes data through the complete ML pipeline
func (e *Engine) ProcessData(ctx context.Context, data []types.PiholeRecord) (*MLResults, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.initialized {
		return nil, fmt.Errorf("ML engine not initialized")
	}

	e.logger.Info("Processing data through ML pipeline: record_count=%d", len(data))

	startTime := time.Now()
	results := &MLResults{
		ProcessedAt: time.Now(),
	}

	period := 24 * time.Hour

	// Detect anomalies if enabled (direct call to avoid double locking)
	if e.config.AnomalyDetection.Enabled {
		if e.trained && e.anomalyDetector != nil {
			anomalies, err := e.anomalyDetector.DetectAnomalies(ctx, data)
			if err != nil {
				e.logger.Warn("Anomaly detection failed, continuing without anomalies: %v", err)
			} else {
				results.Anomalies = anomalies
			}
		} else {
			e.logger.Warn("Anomaly detector not trained, skipping anomaly detection")
		}
	}

	// Analyze trends if enabled
	if e.config.TrendAnalysis.Enabled && e.trendAnalyzer != nil {
		trends, err := e.trendAnalyzer.AnalyzeTrends(ctx, data, period)
		if err != nil {
			e.logger.Warn("Trend analysis failed, continuing without trends: %v", err)
		} else {
			results.TrendAnalysis = trends
		}
	}

	e.lastAnalysis = time.Now()

	// Generate summary
	results.Summary = MLSummary{
		TotalAnomalies:    len(results.Anomalies),
		HighSeverityCount: countHighSeverity(results.Anomalies),
		TrendDirection:    getTrendDirection(results.TrendAnalysis),
		HealthScore:       calculateHealthScore(results),
		Recommendations:   generateRecommendations(results),
	}

	e.logger.Info("ML processing completed: processing_time=%v anomaly_count=%d",
		time.Since(startTime), len(results.Anomalies))

	if results.TrendAnalysis != nil {
		e.logger.Info("Trend analysis results: insight_count=%d", len(results.TrendAnalysis.Insights))
	}

	return results, nil
}

// IsTrained returns whether the ML engine has been trained
func (e *Engine) IsTrained() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.trained
}

// GetStatus returns the current engine status
func (e *Engine) GetStatus() EngineStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return EngineStatus{
		IsInitialized: e.initialized,
		IsTrained:     e.trained,
		LastTraining:  e.lastTraining,
		LastAnalysis:  e.lastAnalysis,
		Status:        e.status,
		Errors:        append([]string{}, e.errors...),
	}
}

// Helper functions
func countHighSeverity(anomalies []Anomaly) int {
	count := 0
	for _, a := range anomalies {
		if a.Severity == SeverityHigh {
			count++
		}
	}
	return count
}

func getTrendDirection(analysis *TrendAnalysis) TrendDirection {
	if analysis == nil {
		return TrendStable
	}
	return analysis.QueryTrend
}

func calculateHealthScore(results *MLResults) float64 {
	if results == nil {
		return 50.0
	}

	score := 100.0
	score -= float64(results.Summary.TotalAnomalies) * 2.0
	score -= float64(results.Summary.HighSeverityCount) * 10.0

	if score < 0 {
		score = 0
	}
	return score
}

func generateRecommendations(results *MLResults) []string {
	var recommendations []string

	if results.Summary.TotalAnomalies > 0 {
		recommendations = append(recommendations, "Investigate detected anomalies")
	}

	if results.Summary.HighSeverityCount > 0 {
		recommendations = append(recommendations, "Address high-severity anomalies immediately")
	}

	if results.Summary.HealthScore < 50 {
		recommendations = append(recommendations, "Network health is concerning - review security settings")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Network appears healthy")
	}

	return recommendations
}

// GetAnalyzerInfo implements the TrendAnalyzer interface
func (e *Engine) GetAnalyzerInfo() AnalyzerInfo {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return AnalyzerInfo{
		Name:       "PiHole ML Trend Analyzer",
		Version:    "1.0.0",
		Algorithms: []string{"exponential_smoothing", "pattern_detection"},
		Parameters: map[string]interface{}{
			"smoothing_factor": e.config.TrendAnalysis.SmoothingFactor,
			"min_data_points":  e.config.TrendAnalysis.MinDataPoints,
			"analysis_window":  e.config.TrendAnalysis.AnalysisWindow.String(),
		},
		LastRun: e.lastAnalysis,
	}
}
