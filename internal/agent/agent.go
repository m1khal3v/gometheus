package agent

import (
	"github.com/m1khal3v/gometheus/internal/client"
	"github.com/m1khal3v/gometheus/internal/collector"
	"github.com/m1khal3v/gometheus/internal/collector/random"
	"github.com/m1khal3v/gometheus/internal/collector/runtime"
	"github.com/m1khal3v/gometheus/internal/storage"
	"sync"
	"time"
)

const pollInterval = 2 * time.Second
const reportInterval = 10 * time.Second

var mutex sync.Mutex
var allMetrics []*storage.Metric

func Start() {
	go sendMetrics()
	go collectMetrics()
	for {
		time.Sleep(1 * time.Second)
	}
}

func collectMetrics() {
	for range time.Tick(pollInterval) {
		metrics, err := collector.CollectAll(
			runtime.NewCollector(),
			random.NewCollector(0, 512),
		)
		// TODO: logger
		if err != nil {
			return
		}
		mutex.Lock()
		allMetrics = append(allMetrics, metrics...)
		mutex.Unlock()
	}
}

func sendMetrics() {
	for range time.Tick(reportInterval) {
		mutex.Lock()
		for _, metric := range allMetrics {
			err := client.SendMetric(*metric)
			// TODO: logger
			if err != nil {
				return
			}
		}
		allMetrics = allMetrics[:0]
		mutex.Unlock()
	}
}
