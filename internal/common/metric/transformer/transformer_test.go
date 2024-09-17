package transformer

import (
	"testing"

	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"github.com/m1khal3v/gometheus/pkg/request"
	"github.com/m1khal3v/gometheus/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func TestTransformToGetResponse(t *testing.T) {
	tests := []struct {
		name    string
		metric  metric.Metric
		want    *response.GetMetricResponse
		wantErr error
	}{
		{
			name:   "counter",
			metric: counter.New("test", 123),
			want: &response.GetMetricResponse{
				MetricType: counter.MetricType,
				MetricName: "test",
				Delta:      ptr.To(int64(123)),
			},
		},
		{
			name:   "gauge",
			metric: gauge.New("test", 123.321),
			want: &response.GetMetricResponse{
				MetricType: gauge.MetricType,
				MetricName: "test",
				Value:      ptr.To(123.321),
			},
		},
		{
			name:    "invalid",
			metric:  &invalidMetric{},
			wantErr: newErrUnknownType("invalid"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TransformToGetResponse(tt.metric)
			if tt.wantErr != nil {
				assert.Nil(t, got)
				assert.Equal(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTransformToSaveRequest(t *testing.T) {
	tests := []struct {
		name    string
		metric  metric.Metric
		want    *request.SaveMetricRequest
		wantErr error
	}{
		{
			name:   "counter",
			metric: counter.New("test", 123),
			want: &request.SaveMetricRequest{
				MetricType: counter.MetricType,
				MetricName: "test",
				Delta:      ptr.To(int64(123)),
			},
		},
		{
			name:   "gauge",
			metric: gauge.New("test", 123.321),
			want: &request.SaveMetricRequest{
				MetricType: gauge.MetricType,
				MetricName: "test",
				Value:      ptr.To(123.321),
			},
		},
		{
			name:    "invalid",
			metric:  &invalidMetric{},
			wantErr: newErrUnknownType("invalid"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TransformToSaveRequest(tt.metric)
			if tt.wantErr != nil {
				assert.Nil(t, got)
				assert.Equal(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTransformToSaveResponse(t *testing.T) {
	tests := []struct {
		name    string
		metric  metric.Metric
		want    *response.SaveMetricResponse
		wantErr error
	}{
		{
			name:   "counter",
			metric: counter.New("test", 123),
			want: &response.SaveMetricResponse{
				MetricType: counter.MetricType,
				MetricName: "test",
				Delta:      ptr.To(int64(123)),
			},
		},
		{
			name:   "gauge",
			metric: gauge.New("test", 123.321),
			want: &response.SaveMetricResponse{
				MetricType: gauge.MetricType,
				MetricName: "test",
				Value:      ptr.To(123.321),
			},
		},
		{
			name:    "invalid",
			metric:  &invalidMetric{},
			wantErr: newErrUnknownType("invalid"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TransformToSaveResponse(tt.metric)
			if tt.wantErr != nil {
				assert.Nil(t, got)
				assert.Equal(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

type invalidMetric struct {
}

func (metric *invalidMetric) Name() string {
	return "invalid"
}

func (metric *invalidMetric) Type() string {
	return "invalid"
}

func (metric *invalidMetric) StringValue() string {
	return "invalid"
}

func (metric *invalidMetric) Clone() metric.Metric {
	return &invalidMetric{}
}
