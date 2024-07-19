package queue

type Queue[T any] struct {
	items chan T
}

func New[T any](size uint64) *Queue[T] {
	return &Queue[T]{
		items: make(chan T, size),
	}
}

func (queue *Queue[T]) Push(metric T) {
	queue.items <- metric
}

func (queue *Queue[T]) Pop(count uint64) []T {
	if count == 0 || len(queue.items) == 0 {
		return []T{}
	}

	metrics := make([]T, 0, count)

	for i := uint64(0); i < count; i++ {
		select {
		case metric := <-queue.items:
			metrics = append(metrics, metric)
		default:
			break
		}
	}

	return metrics
}

func (queue *Queue[T]) Count() uint64 {
	return uint64(len(queue.items))
}
