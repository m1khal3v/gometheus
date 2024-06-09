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
		wantErr error
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
		{
			name: "add invalid name",
			current: &Metric{
				name:  "test",
				value: 1,
			},
			other: &Metric{
				name:  "invalid",
				value: 1,
			},
			wantErr: newErrNamesDontMatch("test", "invalid"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.current.Add(tt.other)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
