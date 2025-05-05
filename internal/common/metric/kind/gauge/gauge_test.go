package gauge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGaugeMetric(t *testing.T) {
	metricName := "test_metric"
	metricValue := 42.0

	metric := New(metricName, metricValue)

	assert.Equal(t, metricName, metric.Name(), "Metric name should match the input name")
	assert.Equal(t, metricValue, metric.GetValue(), "Metric value should match the input value")
	assert.Equal(t, MetricType, metric.Type(), "Metric type should always be 'gauge'")
}

func TestGaugeMetricStringValue(t *testing.T) {
	metric := New("test_metric", 42.0)

	expectedStringValue := "42"
	assert.Equal(t, expectedStringValue, metric.StringValue(), "StringValue should return a formatted metric value")
}

func TestGaugeMetricClone(t *testing.T) {
	original := New("test_metric", 99.9)

	clone := original.Clone()

	assert.NotSame(t, original, clone, "Clone should return a new instance of Metric")
	assert.Equal(t, original.Name(), clone.Name(), "Cloned metric should have the same name as the original")
	assert.Equal(t, original.StringValue(), clone.StringValue(), "Cloned metric should have the same value as the original")
	assert.Equal(t, original.Type(), clone.Type(), "Cloned metric should have the same type as the original")

	// Change the original to ensure the clone is independent
	original.value = 10.0
	assert.NotEqual(t, original.StringValue(), clone.StringValue(), "Changing original metric should not affect the clone")
}
