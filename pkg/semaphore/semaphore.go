// Package semaphore
// contains channel based semaphore implementation
package semaphore

import "context"

type Semaphore struct {
	channel chan struct{}
}

func New(max uint64) *Semaphore {
	if max == 0 {
		panic("max cannot be 0")
	}

	return &Semaphore{
		channel: make(chan struct{}, max),
	}
}

// Acquire trying to increment semaphore, waits if can`t. Return cause if context closed
func (semaphore *Semaphore) Acquire(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case semaphore.channel <- struct{}{}:
		return nil
	}
}

// Release semaphore
func (semaphore *Semaphore) Release() {
	<-semaphore.channel
}
