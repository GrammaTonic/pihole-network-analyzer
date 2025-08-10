package main

import (
	"context"
	"fmt"
	"pihole-analyzer/internal/ml"
	"pihole-analyzer/internal/types"
)

func main() {
	config := ml.DefaultMLConfig().AnomalyDetection
	fmt.Printf("Threshold: %f\n", config.Thresholds["unusual_domain_threshold"])

	detector := ml.NewStatisticalAnomalyDetector(config, nil)
	ctx := context.Background()
	detector.Initialize(ctx, config)

	// Create test data with exactly 1 unusual domain with 2% frequency
	data := make([]types.PiholeRecord, 50)
	for i := 0; i < 50; i++ {
		data[i] = types.PiholeRecord{
			Domain: "suspicious.com",
		}
	}

	anomalies, _ := detector.DetectAnomalies(ctx, data)
	fmt.Printf("Found %d anomalies\n", len(anomalies))
}
