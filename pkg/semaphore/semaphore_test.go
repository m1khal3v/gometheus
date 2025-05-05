package semaphore

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		max       uint64
		wantPanic bool
	}{
		{
			name: "valid",
			max:  10,
		},
		{
			name:      "invalid",
			max:       0,
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() {
					New(tt.max)
				})
			} else {
				semaphore := New(tt.max)
				assert.Len(t, semaphore.channel, 0)
			}
		})
	}
}

func TestSemaphore(t *testing.T) {
	semaphore := New(1)
	ctx := context.Background()
	require.NoError(t, semaphore.Acquire(ctx))
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	require.Error(t, semaphore.Acquire(cancelCtx))
	semaphore.Release()
	require.NoError(t, semaphore.Acquire(ctx))
}

func TestSemaphore_OverAcquire(t *testing.T) {
	semaphore := New(2)
	ctx := context.Background()

	require.NoError(t, semaphore.Acquire(ctx))
	require.NoError(t, semaphore.Acquire(ctx))

	done := make(chan struct{})
	go func() {
		require.NoError(t, semaphore.Acquire(ctx))
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("Semaphore acquired more times than allowed")
	case <-time.After(100 * time.Millisecond):
		// Expected timeout as semaphore limit is reached
	}
}

func TestSemaphore_ConcurrentAcquireRelease(t *testing.T) {
	semaphore := New(2)
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	count := 100

	acquired := make(chan struct{}, count)

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			require.NoError(t, semaphore.Acquire(ctx))
			acquired <- struct{}{}
			semaphore.Release()
		}()
	}

	wg.Wait()
	close(acquired)

	assert.Len(t, acquired, count, "Each acquire should be matched with a release")
}

func TestSemaphore_CanceledContext(t *testing.T) {
	semaphore := New(1)
	ctx, cancel := context.WithCancel(context.Background())

	require.NoError(t, semaphore.Acquire(ctx))
	cancel()

	err := semaphore.Acquire(ctx)
	require.Error(t, err, "Expected error when acquiring with canceled context")
	assert.ErrorIs(t, err, context.Canceled)
}
