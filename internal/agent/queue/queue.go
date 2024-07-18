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

func (queue *Queue) Push(metric metric.Metric) {
	queue.items <- metric.Clone()
}

func (queue *Queue) Pop(count uint64) []metric.Metric {
	if count == 0 || len(queue.items) == 0 {
		return []metric.Metric{}
	}

	metrics := make([]metric.Metric, 0, count)

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

func (queue *Queue) Count() uint64 {
	return uint64(len(queue.items))
}
