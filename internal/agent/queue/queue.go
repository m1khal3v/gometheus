package queue

import (
	"github.com/m1khal3v/gometheus/internal/common/metric"
)

type Queue struct {
	items chan metric.Metric
}

func New(size uint64) *Queue {
	return &Queue{
		items: make(chan metric.Metric, size),
	}
}

func (queue *Queue) Push(metrics <-chan metric.Metric) {
	for metric := range metrics {
		queue.items <- metric.Clone()
	}
}

func (queue *Queue) PushSlice(metrics []metric.Metric) {
	for _, metric := range metrics {
		queue.items <- metric.Clone()
	}
}

func (queue *Queue) Pop(count uint64) <-chan metric.Metric {
	channel := make(chan metric.Metric, count)
	if count == 0 {
		close(channel)
		return channel
	}

	go func() {
		for i := uint64(0); i < count; i++ {
			channel <- <-queue.items
		}

		close(channel)
	}()

	return channel
}
