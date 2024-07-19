package gopsutil

import (
	"fmt"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"golang.org/x/sync/errgroup"
	"slices"
)

const MetricTotalMemory = "memTotal"
const MetricFreeMemory = "memFree"
const MetricCPUUtilization = "cpuUtilization"

func getAvailableMetrics() []string {
	return []string{
		MetricTotalMemory,
		MetricFreeMemory,
		MetricCPUUtilization,
	}
}

type Collector struct {
	channelSize uint16
	metrics     map[string]string
}

type ErrInvalidMetricName struct {
	Name string
}

func (err ErrInvalidMetricName) Error() string {
	return fmt.Sprintf("invalid metric name: %s", err.Name)
}

func newErrInvalidMetricName(name string) error {
	return &ErrInvalidMetricName{
		Name: name,
	}
}

func New(metrics map[string]string) (*Collector, error) {
	collector := &Collector{}

	for _, metric := range metrics {
		if !slices.Contains(getAvailableMetrics(), metric) {
			return nil, newErrInvalidMetricName(metric)
		}

		if metric == MetricCPUUtilization {
			cpuCount, err := cpu.Counts(true)
			if err != nil {
				return nil, err
			}

			collector.channelSize += uint16(cpuCount)
		} else {
			collector.channelSize++
		}
	}

	collector.metrics = metrics

	return collector, nil
}

func (collector *Collector) Collect() (<-chan metric.Metric, error) {
	channel := make(chan metric.Metric, collector.channelSize)
	defer close(channel)

	var errGroup errgroup.Group
	for metricName, name := range collector.metrics {
		errGroup.Go(func() error {
			switch name {
			case MetricTotalMemory:
				memory, err := mem.VirtualMemory()
				if err != nil {
					return err
				}

				channel <- gauge.New(metricName, float64(memory.Total))
			case MetricFreeMemory:
				memory, err := mem.VirtualMemory()
				if err != nil {
					return err
				}

				channel <- gauge.New(metricName, float64(memory.Free))
			case MetricCPUUtilization:
				utilization, err := cpu.Percent(0, true)
				if err != nil {
					return err
				}

				for i, value := range utilization {
					channel <- gauge.New(
						fmt.Sprintf("%s%d", metricName, i+1),
						value,
					)
				}
			}

			return nil
		})
	}

	if err := errGroup.Wait(); err != nil {
		return nil, err
	}

	return channel, nil
}
