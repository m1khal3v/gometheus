package rpc

import (
	"context"
	"testing"

	"github.com/m1khal3v/gometheus/internal/server/storage/kind/memory"
	"github.com/m1khal3v/gometheus/pkg/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestMetricsService_SaveMetric(t *testing.T) {
	inMemoryStorage := memory.New()

	metricsService := NewMetricsService(inMemoryStorage)

	request := &proto.SaveMetricRequest{
		MetricName: "test_metric",
		Value:      wrapperspb.Double(42),
		MetricType: "gauge",
	}

	response, err := metricsService.SaveMetric(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, "test_metric", response.GetMetricName())
	require.Equal(t, float64(42.0), response.GetValue().GetValue())
	require.Equal(t, "gauge", response.GetMetricType())

	savedMetric, err := inMemoryStorage.Get(context.Background(), "test_metric")
	require.NoError(t, err)
	require.NotNil(t, savedMetric)
	require.Equal(t, "gauge", savedMetric.Type())
	require.Equal(t, "42", savedMetric.StringValue())
}

func TestMetricsService_SaveMetrics(t *testing.T) {
	inMemoryStorage := memory.New()

	metricsService := NewMetricsService(inMemoryStorage)

	request := &proto.SaveMetricsBatchRequest{
		Metrics: []*proto.SaveMetricRequest{
			{
				MetricName: "test_metric_1",
				Delta:      wrapperspb.Int64(200),
				MetricType: "counter",
			},
			{
				MetricName: "test_metric_2",
				Value:      wrapperspb.Double(100),
				MetricType: "gauge",
			},
		},
	}

	response, err := metricsService.SaveMetrics(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Len(t, response.GetMetrics(), 2)

	// Проверяем первую метрику
	require.Equal(t, "test_metric_1", response.GetMetrics()[0].GetMetricName())
	require.Equal(t, int64(200), response.GetMetrics()[0].GetDelta().GetValue())
	require.Equal(t, "counter", response.GetMetrics()[0].GetMetricType())

	savedMetric1, err := inMemoryStorage.Get(context.Background(), "test_metric_1")
	require.NoError(t, err)
	require.NotNil(t, savedMetric1)
	require.Equal(t, "counter", savedMetric1.Type())
	require.Equal(t, "200", savedMetric1.StringValue())

	// Проверяем вторую метрику
	require.Equal(t, "test_metric_2", response.GetMetrics()[1].GetMetricName())
	require.Equal(t, float64(100.0), response.GetMetrics()[1].GetValue().GetValue())
	require.Equal(t, "gauge", response.GetMetrics()[1].GetMetricType())

	savedMetric2, err := inMemoryStorage.Get(context.Background(), "test_metric_2")
	require.NoError(t, err)
	require.NotNil(t, savedMetric2)
	require.Equal(t, "gauge", savedMetric2.Type())
	require.Equal(t, "100", savedMetric2.StringValue())
}
