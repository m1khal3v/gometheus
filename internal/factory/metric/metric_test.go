package metric

import (
	"github.com/m1khal3v/gometheus/internal/metric"
	"github.com/m1khal3v/gometheus/internal/metric/counter"
	"github.com/m1khal3v/gometheus/internal/metric/gauge"
	"github.com/stretchr/testify/assert"
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
			wantErr: newUnknownTypeError("invalid"),
		},
		{
			name: "test invalid gauge",
			args: args{
				metricType: "gauge",
				name:       "test",
				value:      "abc123.321",
			},
			wantErr: newInvalidValueError("abc123.321"),
		},
		{
			name: "test invalid counter",
			args: args{
				metricType: "counter",
				name:       "test",
				value:      "123.321",
			},
			wantErr: newInvalidValueError("123.321"),
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
