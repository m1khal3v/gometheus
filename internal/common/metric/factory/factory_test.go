package factory

import (
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"github.com/m1khal3v/gometheus/pkg/request"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		metricType string
		name       string
		value      string
	}
	tests := []struct {
		name    string
		args    args
		want    metric.Metric
		wantErr error
	}{
		{
			name: "test gauge",
			args: args{
				metricType: "gauge",
				name:       "test",
				value:      "123.321",
			},
			want: gauge.New("test", 123.321),
		},
		{
			name: "test counter",
			args: args{
				metricType: "counter",
				name:       "test",
				value:      "123",
			},
			want: counter.New("test", 123),
		},
		{
			name: "test invalid type",
			args: args{
				metricType: "invalid",
				name:       "test",
				value:      "123.321",
			},
			wantErr: newErrUnknownType("invalid"),
		},
		{
			name: "test invalid gauge",
			args: args{
				metricType: "gauge",
				name:       "test",
				value:      "abc123.321",
			},
			wantErr: newErrInvalidValue("abc123.321"),
		},
		{
			name: "test invalid counter",
			args: args{
				metricType: "counter",
				name:       "test",
				value:      "123.321",
			},
			wantErr: newErrInvalidValue("123.321"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.metricType, tt.args.name, tt.args.value)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestNewFromRequest(t *testing.T) {
	tests := []struct {
		name    string
		request request.SaveMetricRequest
		want    metric.Metric
		wantErr error
	}{
		{
			name: "test gauge",
			request: request.SaveMetricRequest{
				MetricName: "test",
				MetricType: "gauge",
				Value:      ptr.To(123.321),
			},
			want: gauge.New("test", 123.321),
		},
		{
			name: "test counter",
			request: request.SaveMetricRequest{
				MetricName: "test",
				MetricType: "counter",
				Delta:      ptr.To(int64(123)),
			},
			want: counter.New("test", 123),
		},
		{
			name: "test invalid type",
			request: request.SaveMetricRequest{
				MetricName: "test",
				MetricType: "invalid",
				Value:      ptr.To(123.321),
			},
			wantErr: newErrUnknownType("invalid"),
		},
		{
			name: "test nil gauge",
			request: request.SaveMetricRequest{
				MetricName: "test",
				MetricType: "gauge",
			},
			wantErr: newErrInvalidValue("nil"),
		},
		{
			name: "test nil counter",
			request: request.SaveMetricRequest{
				MetricName: "test",
				MetricType: "counter",
			},
			wantErr: newErrInvalidValue("nil"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewFromRequest(tt.request)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
