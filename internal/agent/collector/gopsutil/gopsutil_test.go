package gopsutil

import (
	"testing"

	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	cpuCount, err := cpu.Counts(true)
	require.NoError(t, err)
	tests := []struct {
		name    string
		metrics MetricMap
		want    *Collector
		wantErr error
	}{
		{
			name: "valid",
			metrics: MetricMap{
				FreeMemory:     "Free",
				TotalMemory:    "Total",
				CPUUtilization: "CPU",
			},
			want: &Collector{
				metrics: MetricMap{
					FreeMemory:     "Free",
					TotalMemory:    "Total",
					CPUUtilization: "CPU",
				},
				channelSize: uint16(2 + cpuCount),
			},
			wantErr: nil,
		},
		{
			name: "invalid metric",
			metrics: MetricMap{
				"Invalid": "interest",
			},
			wantErr: newErrInvalidMetricName("Invalid"),
		},
		{
			name:    "empty metrics",
			metrics: MetricMap{},
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
		metrics MetricMap
	}{
		{
			name: "valid 1",
			metrics: MetricMap{
				FreeMemory: "Free",
			},
		},
		{
			name: "valid 2",
			metrics: MetricMap{
				TotalMemory: "Total",
			},
		},
		{
			name: "valid 3",
			metrics: MetricMap{
				CPUUtilization: "CPU",
			},
		},
		{
			name: "valid 4",
			metrics: MetricMap{
				FreeMemory:     "Free",
				TotalMemory:    "Total",
				CPUUtilization: "CPU",
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
			if collector.isset(CPUUtilization) {
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
