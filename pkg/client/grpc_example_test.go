package client

import (
	"context"
	"fmt"
	"log"

	"github.com/m1khal3v/gometheus/pkg/request"
)

func Example_grpc() {
	grpcClient, err := NewGRPC("127.0.0.1:50051")
	if err != nil {
		log.Fatalf("failed to create grpc client: %v", err)
	}

	// Пример сохранения метрики
	req := &request.SaveMetricRequest{
		MetricName: "example_metric",
		MetricType: "gauge",
		Value:      ptrFloat64(1.23),
	}
	resp, apiErr, err := grpcClient.SaveMetric(context.Background(), req)
	if err != nil {
		log.Fatalf("failed to save metric: %v", err)
	}
	if apiErr != nil {
		log.Printf("api error: %+v", apiErr)
	}
	fmt.Printf("Saved metric: %+v\n", resp)

	// Пример сохранения нескольких метрик
	requests := []request.SaveMetricRequest{
		{
			MetricName: "metric_1",
			MetricType: "gauge",
			Value:      ptrFloat64(0.89),
		},
		{
			MetricName: "metric_2",
			MetricType: "counter",
			Delta:      ptrInt64(10),
		},
	}
	batchResp, apiErr, err := grpcClient.SaveMetrics(context.Background(), requests)
	if err != nil {
		log.Fatalf("failed to save metrics: %v", err)
	}
	if apiErr != nil {
		log.Printf("api error: %+v", apiErr)
	}
	fmt.Printf("Saved metrics batch: %+v\n", batchResp)
}
