package queue

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
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
			count := uint64(len(tt.items)) * 3
			queue := New[bool](count)
			channel := make(chan bool, count/3)
			for _, item := range tt.items {
				queue.Push(item)
				channel <- item
			}
			close(channel)
			queue.PushBatch(tt.items)
			queue.PushChannel(channel)

			require.Equal(t, count, queue.Count())
			items := make([]bool, 0, count)
			items = append(items, queue.Pop(count)...)
			expectedItems := make([]bool, 0, count)
			expectedItems = append(expectedItems, tt.items...)
			expectedItems = append(expectedItems, tt.items...)
			expectedItems = append(expectedItems, tt.items...)
			require.Equal(t, expectedItems, items)
			require.Equal(t, uint64(0), queue.Count())
			queue.PushBatch(tt.items)
			err := queue.RemoveBatch(2, func(items []bool) error {
				for _, item := range items {
					if !item {
						return errors.New("do not remove")
					}
				}

				return nil
			})
			if len(tt.items) > 0 {
				require.Error(t, err)
				require.Equal(t, uint64(2), queue.Count())
			} else {
				require.NoError(t, err)
				require.Equal(t, uint64(0), queue.Count())
			}
			require.NoError(t, queue.RemoveBatch(2, func(items []bool) error {
				return nil
			}))
			require.Equal(t, uint64(0), queue.Count())
		})
	}
}

func TestQueueEnhanced(t *testing.T) {
	t.Run("Pop with zero count", func(t *testing.T) {
		queue := New[int](10)
		queue.Push(1)
		queue.Push(2)

		items := queue.Pop(0)
		require.Empty(t, items, "Expected empty result when count is zero")
		require.Equal(t, uint64(2), queue.Count(), "Queue count should remain the same")
	})

	t.Run("Concurrency test", func(t *testing.T) {
		queue := New[int](100)
		wg := sync.WaitGroup{}
		wg.Add(2)

		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				queue.Push(i)
			}
		}()

		go func() {
			defer wg.Done()
			for i := 50; i < 100; i++ {
				queue.Push(i)
			}
		}()

		wg.Wait()

		require.Equal(t, uint64(100), queue.Count(), "Incorrect item count in the queue after concurrent pushes")
	})

	t.Run("RemoveBatch with zero count", func(t *testing.T) {
		queue := New[int](10)
		queue.Push(1)
		queue.Push(2)

		err := queue.RemoveBatch(0, func(items []int) error {
			return nil
		})

		require.NoError(t, err, "Expected no error when removing zero items")
		require.Equal(t, uint64(2), queue.Count(), "Queue count should remain unchanged")
	})

	t.Run("RemoveBatch filter rejects all items", func(t *testing.T) {
		queue := New[int](10)
		queue.PushBatch([]int{1, 2, 3})

		err := queue.RemoveBatch(3, func(items []int) error {
			return errors.New("reject all items")
		})

		require.Error(t, err, "Expected error when filter rejects all items")
		require.Equal(t, uint64(3), queue.Count(), "Queue count should remain unchanged after rejected removal")
	})

	t.Run("Edge case: Pop more items than present", func(t *testing.T) {
		queue := New[int](5)
		queue.Push(1)
		queue.Push(2)

		items := queue.Pop(10)
		require.Equal(t, []int{1, 2}, items, "Should return all available items when count exceeds present items")
		require.Equal(t, uint64(0), queue.Count(), "Queue count should be zero after popping all items")
	})
}
