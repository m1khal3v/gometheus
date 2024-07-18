package semaphore

type Semaphore struct {
	channel chan struct{}
}

func New(max uint64) *Semaphore {
	channel := make(chan struct{}, max)
	for i := 0; i < int(max); i++ {
		channel <- struct{}{}
	}

	return &Semaphore{
		channel: channel,
	}
}

// Acquire implementation of a semaphore with the ability to use inside a select block
// requires reading one message from the channel to Acquire once
func (semaphore *Semaphore) Acquire() <-chan struct{} {
	return semaphore.channel
}

func (semaphore *Semaphore) Release() {
	semaphore.channel <- struct{}{}
}
