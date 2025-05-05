package counter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetric_Type(t *testing.T) {
	metric := New("test_metric", 0)
	assert.Equal(t, "counter", metric.Type(), "Metric type should be 'counter'")
}

func TestMetric_Name(t *testing.T) {
	name := "test_metric"
	metric := New(name, 0)
	assert.Equal(t, name, metric.Name(), "Metric name does not match")
}

func TestMetric_StringValue(t *testing.T) {
	initialValue := int64(42)
	metric := New("test_metric", initialValue)
	assert.Equal(t, "42", metric.StringValue(), "String representation of metric value is incorrect")
}

func TestMetric_Clone(t *testing.T) {
	metric := New("test_metric", 100)
	clone := metric.Clone()

	// Проверяем, что клон вернул новую структуру
	assert.NotSame(t, metric, clone, "Clone should return a new metric instance")
	// Проверяем, что значения клона соответствуют оригиналу
	assert.Equal(t, metric.Type(), clone.Type(), "Cloned metric type does not match")
	assert.Equal(t, metric.Name(), clone.Name(), "Cloned metric name does not match")
	assert.Equal(t, metric.StringValue(), clone.StringValue(), "Cloned metric value does not match")
}

func TestMetric_GetValue(t *testing.T) {
	initialValue := int64(50)
	metric := New("test_metric", initialValue)
	assert.Equal(t, initialValue, metric.GetValue(), "Initial metric value does not match")
}

func TestMetric_Add(t *testing.T) {
	metric := New("test_metric", 10)
	metric.Add(5)
	assert.Equal(t, int64(15), metric.GetValue(), "Metric value after addition is incorrect")

	metric.Add(-3)
	assert.Equal(t, int64(12), metric.GetValue(), "Metric value after subtraction is incorrect")
}

func TestNew(t *testing.T) {
	name := "test_metric"
	initialValue := int64(77)
	metric := New(name, initialValue)

	assert.Equal(t, name, metric.Name(), "New metric name does not match")
	assert.Equal(t, initialValue, metric.GetValue(), "New metric initial value does not match")
}
