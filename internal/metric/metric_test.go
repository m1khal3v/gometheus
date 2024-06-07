package metric

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetric_GetValue(t *testing.T) {
	type fields struct {
		Type       string
		Name       string
		FloatValue float64
		IntValue   int64
	}
	tests := []struct {
		name   string
		fields fields
		want   any
	}{
		{
			name: "gauge",
			fields: fields{
				Type:       "gauge",
				Name:       "test",
				FloatValue: 123.456,
				IntValue:   0,
			},
			want: float64(123.456),
		},
		{
			name: "counter",
			fields: fields{
				Type:       "counter",
				Name:       "test",
				FloatValue: 0,
				IntValue:   123,
			},
			want: int64(123),
		},
		{
			name: "invalid",
			fields: fields{
				Type:       "invalid",
				Name:       "test",
				FloatValue: 123.456,
				IntValue:   123,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric := &Metric{
				Type:       tt.fields.Type,
				Name:       tt.fields.Name,
				FloatValue: tt.fields.FloatValue,
				IntValue:   tt.fields.IntValue,
			}
			assert.Equal(t, tt.want, metric.GetValue())
		})
	}
}

func TestNewMetric(t *testing.T) {
	type args struct {
		metricType string
		name       string
		value      any
	}
	tests := []struct {
		name    string
		args    args
		want    *Metric
		wantErr error
	}{
		{
			name: "valid gauge",
			args: args{
				metricType: "gauge",
				name:       "test valid gauge",
				value:      float64(123.456),
			},
			want: &Metric{
				Type:       "gauge",
				Name:       "test valid gauge",
				FloatValue: float64(123.456),
			},
		},
		{
			name: "valid gauge string",
			args: args{
				metricType: "gauge",
				name:       "test valid gauge string",
				value:      "123.456",
			},
			want: &Metric{
				Type:       "gauge",
				Name:       "test valid gauge string",
				FloatValue: float64(123.456),
			},
		},
		{
			name: "valid counter",
			args: args{
				metricType: "counter",
				name:       "test valid counter",
				value:      int64(123),
			},
			want: &Metric{
				Type:     "counter",
				Name:     "test valid counter",
				IntValue: int64(123),
			},
		},
		{
			name: "valid counter string",
			args: args{
				metricType: "counter",
				name:       "test valid counter",
				value:      "123",
			},
			want: &Metric{
				Type:     "counter",
				Name:     "test valid counter",
				IntValue: int64(123),
			},
		},
		{
			name: "invalid metric type",
			args: args{
				metricType: "invalid",
				name:       "test invalid metric type",
				value:      "123",
			},
			wantErr: ErrUnknownType{
				Type: "invalid",
			},
		},
		{
			name: "invalid gauge value",
			args: args{
				metricType: "gauge",
				name:       "test invalid gauge value",
				value:      int64(123),
			},
			wantErr: ErrInvalidValue{
				Value: "123",
			},
		},
		{
			name: "invalid counter value",
			args: args{
				metricType: "counter",
				name:       "test invalid counter value",
				value:      float64(123.321),
			},
			wantErr: ErrInvalidValue{
				Value: "123.321",
			},
		},
		{
			name: "invalid gauge string value",
			args: args{
				metricType: "gauge",
				name:       "test invalid gauge string value",
				value:      "1b42",
			},
			wantErr: ErrInvalidValue{
				Value: "1b42",
			},
		},
		{
			name: "invalid counter string value",
			args: args{
				metricType: "counter",
				name:       "test invalid counter string value",
				value:      "123.321",
			},
			wantErr: ErrInvalidValue{
				Value: "123.321",
			},
		},
		{
			name: "empty gauge string value",
			args: args{
				metricType: "gauge",
				name:       "test empty gauge string value",
				value:      "",
			},
			wantErr: ErrInvalidValue{
				Value: "",
			},
		},
		{
			name: "empty counter string value",
			args: args{
				metricType: "counter",
				name:       "test empty counter string value",
				value:      "",
			},
			wantErr: ErrInvalidValue{
				Value: "",
			},
		},
		{
			name: "invalid value type int32",
			args: args{
				metricType: "counter",
				name:       "test invalid value type int32",
				value:      int32(123),
			},
			wantErr: ErrInvalidValueType{},
		},
		{
			name: "invalid value type float32",
			args: args{
				metricType: "counter",
				name:       "test invalid value type float32",
				value:      float32(123.321),
			},
			wantErr: ErrInvalidValueType{},
		},
		{
			name: "invalid value type bool",
			args: args{
				metricType: "counter",
				name:       "test invalid value type float32",
				value:      true,
			},
			wantErr: ErrInvalidValueType{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMetric(tt.args.metricType, tt.args.name, tt.args.value)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
