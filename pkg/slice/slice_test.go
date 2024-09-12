package slice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunk(t *testing.T) {
	tests := []struct {
		name  string
		slice []uint8
		n     uint64
		want  [][]uint8
	}{
		{
			name:  "1/3",
			slice: []uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			n:     3,
			want: [][]uint8{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
				{10, 11, 12},
				{13, 14, 15},
				{16, 17, 18},
				{19, 20},
			},
		},
		{
			name:  "2/3",
			slice: []uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			n:     5,
			want: [][]uint8{
				{1, 2, 3, 4, 5},
				{6, 7, 8, 9, 10},
				{11, 12, 13, 14, 15},
				{16, 17, 18, 19, 20},
			},
		},
		{
			name:  "3/3",
			slice: []uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			n:     9,
			want: [][]uint8{
				{1, 2, 3, 4, 5, 6, 7, 8, 9},
				{10, 11, 12, 13, 14, 15, 16, 17, 18},
				{19, 20},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var chunks [][]uint8
			for item := range Chunk(tt.slice, tt.n) {
				chunks = append(chunks, item)
			}
			assert.Equal(t, tt.want, chunks)
		})
	}
}

func TestFromChannel(t *testing.T) {
	tests := []struct {
		name  string
		items []any
	}{
		{
			name:  "one item",
			items: []any{123},
		},
		{
			name:  "multiple items",
			items: []any{123.321, 321.123, 555.646},
		},
		{
			name:  "empty items",
			items: []any{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channel := make(chan any, len(tt.items))
			for _, item := range tt.items {
				channel <- item
			}
			close(channel)

			assert.Equal(t, tt.items, FromChannel(channel))
		})
	}
}
