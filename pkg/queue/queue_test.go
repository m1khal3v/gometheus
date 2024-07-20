package queue

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQueue(t *testing.T) {
	tests := []struct {
		name  string
		items []bool
	}{
		{
			"empty",
			[]bool{},
		},
		{
			"not empty",
			[]bool{
				true,
				false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := uint64(len(tt.items))
			queue := New[bool](count)
			for _, item := range tt.items {
				queue.Push(item)
			}
			items := make([]bool, 0, len(tt.items))
			items = append(items, queue.Pop(count)...)
			require.Equal(t, tt.items, items)
		})
	}
}
