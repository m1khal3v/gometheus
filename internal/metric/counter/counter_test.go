package counter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetric_Add(t *testing.T) {
	tests := []struct {
		name    string
		current *Metric
		other   *Metric
		want    *Metric
	}{
		{
			name: "add positive value",
			current: &Metric{
				name:  "test",
				value: 1,
			},
			other: &Metric{
				name:  "test",
				value: 2,
			},
			want: &Metric{
				name:  "test",
				value: 3,
			},
		},
		{
			name: "add negative value",
			current: &Metric{
				name:  "test",
				value: 1,
			},
			other: &Metric{
				name:  "test",
				value: -2,
			},
			want: &Metric{
				name:  "test",
				value: -1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.current.Add(tt.other))
		})
	}
}
