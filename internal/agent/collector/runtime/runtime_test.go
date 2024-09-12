package runtime

import (
	"fmt"
	"testing"

	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		metrics []string
		want    *Collector
		wantErr error
	}{
		{
			name: "valid",
			metrics: []string{
				"Sys",
				"Frees",
			},
			want: &Collector{
				metrics: []string{"Sys", "Frees"},
			},
			wantErr: nil,
		},
		{
			name: "invalid metric",
			metrics: []string{
				"Sys",
				"Frees",
				"Invalid",
			},
			wantErr: newErrInvalidMetricName("Invalid"),
		},
		{
			name:    "empty metrics",
			metrics: []string{},
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
		metrics []string
	}{
		{
			name: "valid 1",
			metrics: []string{
				"Sys",
				"Frees",
			},
		},
		{
			name: "valid 2",
			metrics: []string{
				"NextGC",
				"LastGC",
				"NumGC",
			},
		},
		{
			name: "valid 3",
			metrics: []string{
				"Alloc",
				"TotalAlloc",
				"HeapAlloc",
				"Mallocs",
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
			assert.Len(t, metrics, len(tt.metrics)+1)
			for _, metric := range metrics {
				if metric.Name() == "PollCount" {
					assert.Equal(t, "counter", metric.Type())
					assert.Equal(t, fmt.Sprintf("%d", len(tt.metrics)), metric.StringValue())
				} else {
					assert.Equal(t, "gauge", metric.Type())
					assert.NotEmpty(t, metric.StringValue())
				}
			}
		})
	}
}
