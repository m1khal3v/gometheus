package random

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCollector_Collect(t *testing.T) {
	tests := []struct {
		name string
		min  float64
		max  float64
	}{
		{
			name: "test",
			min:  10,
			max:  20,
		},
		{
			name: "test 2",
			min:  0,
			max:  128.256,
		},
		{
			name: "test 3",
			min:  124234.324234,
			max:  1286644.12323,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector, err := New(tt.min, tt.max)
			assert.Nil(t, err)
			metrics, err := collector.Collect()
			assert.Nil(t, err)
			assert.Len(t, metrics, 1)
			assert.Equal(t, "RandomValue", metrics[0].GetName())
			assert.Equal(t, "gauge", metrics[0].GetType())
			value := metrics[0].GetValue()
			assert.GreaterOrEqual(t, value, tt.min)
			assert.LessOrEqual(t, value, tt.max)
		})
	}
}

func TestNewCollector(t *testing.T) {
	tests := []struct {
		name    string
		min     float64
		max     float64
		want    *Collector
		wantErr error
	}{
		{
			name: "valid collector",
			min:  1,
			max:  2,
			want: &Collector{
				min: 1,
				max: 2,
			},
		},
		{
			name: "invalid collector",
			min:  2,
			max:  1,
			wantErr: ErrMinGreaterThanMax{
				Min: 2,
				Max: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector, err := New(tt.min, tt.max)
			if tt.wantErr == nil {
				assert.Equal(t, tt.want, collector)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}
