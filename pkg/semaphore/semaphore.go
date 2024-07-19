package semaphore

import "context"

type Semaphore struct {
	channel chan struct{}
}

func New(max uint64) *Semaphore {
	if max == 0 {
		panic("max cannot be 0")
	}

	channel := make(chan struct{}, max)
	for i := 0; i < int(max); i++ {
		channel <- struct{}{}
	}

	return &Semaphore{
		channel: channel,
	}
}

func (semaphore *Semaphore) Acquire(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-semaphore.channel:
		return nil
	}
}

func (semaphore *Semaphore) Release() {
	semaphore.channel <- struct{}{}
}
