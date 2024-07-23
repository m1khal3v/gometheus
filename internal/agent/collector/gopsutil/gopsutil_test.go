package gopsutil

import (
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNew(t *testing.T) {
	cpuCount, err := cpu.Counts(true)
	require.NoError(t, err)
	tests := []struct {
		name    string
		metrics map[string]string
		want    *Collector
		wantErr error
	}{
		{
			name: "valid",
			metrics: map[string]string{
				MetricFreeMemory:     "Free",
				MetricTotalMemory:    "Total",
				MetricCPUUtilization: "CPU",
			},
			want: &Collector{
				metrics: map[string]string{
					MetricFreeMemory:     "Free",
					MetricTotalMemory:    "Total",
					MetricCPUUtilization: "CPU",
				},
				channelSize: uint16(2 + cpuCount),
			},
			wantErr: nil,
		},
		{
			name: "invalid metric",
			metrics: map[string]string{
				"Invalid": "Metric",
			},
			wantErr: newErrInvalidMetricName("Invalid"),
		},
		{
			name:    "empty metrics",
			metrics: map[string]string{},
			wantErr: ErrEmptyMetrics,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.metrics)
			if tt.wantErr == nil {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tt.want, got)
			} else {
				assert.Nil(t, got)
				assert.ErrorAs(t, err, &tt.wantErr)
			}
		})
	}
}

func TestCollector_Collect(t *testing.T) {
	tests := []struct {
		name    string
		metrics map[string]string
	}{
		{
			name: "valid 1",
			metrics: map[string]string{
				MetricFreeMemory: "Free",
			},
		},
		{
			name: "valid 2",
			metrics: map[string]string{
				MetricTotalMemory: "Total",
			},
		},
		{
			name: "valid 3",
			metrics: map[string]string{
				MetricCPUUtilization: "CPU",
			},
		},
		{
			name: "valid 4",
			metrics: map[string]string{
				MetricFreeMemory:     "Free",
				MetricTotalMemory:    "Total",
				MetricCPUUtilization: "CPU",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector, err := New(tt.metrics)
			require.NoError(t, err)

			metrics := make([]metric.Metric, 0)
			collected, err := collector.Collect()
			require.NoError(t, err)

			for metric := range collected {
				metrics = append(metrics, metric)
			}
			if collector.isset(MetricCPUUtilization) {
				cpuCount, err := cpu.Counts(true)
				require.NoError(t, err)
				assert.Len(t, metrics, len(tt.metrics)-1+cpuCount)
			} else {
				assert.Len(t, metrics, len(tt.metrics))
			}
			for _, metric := range metrics {
				assert.Equal(t, "gauge", metric.Type())
				assert.NotEmpty(t, metric.StringValue())
			}
		})
	}
}
