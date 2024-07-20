package gopsutil

import (
	"fmt"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
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

type ErrNonUniqueOutName struct {
	Name string
}

func (err ErrNonUniqueOutName) Error() string {
	return fmt.Sprintf("out name '%s' is not unique", err.Name)
}

func newErrNonUniqueOutName(name string) error {
	return &ErrNonUniqueOutName{
		Name: name,
	}
}

func New(metrics map[string]string) (*Collector, error) {
	collector := &Collector{}
	outNames := make(map[string]struct{}, len(metrics))

	for name, outName := range metrics {
		if _, ok := outNames[outName]; ok {
			return nil, newErrNonUniqueOutName(outName)
		}
		outNames[outName] = struct{}{}

		if !slices.Contains(getAvailableMetrics(), name) {
			return nil, newErrInvalidMetricName(name)
		}

		if name == MetricCPUUtilization {
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

	var memory *mem.VirtualMemoryStat
	var utilization []float64
	var err error

	switch {
	case collector.isset(MetricTotalMemory), collector.isset(MetricFreeMemory):
		memory, err = mem.VirtualMemory()
		if err != nil {
			return nil, err
		}
	case collector.isset(MetricCPUUtilization):
		utilization, err = cpu.Percent(0, false)
		if err != nil {
			return nil, err
		}
	}

	go func() {
		defer close(channel)

		for name, outName := range collector.metrics {
			switch name {
			case MetricTotalMemory:
				channel <- gauge.New(outName, float64(memory.Total))
			case MetricFreeMemory:
				channel <- gauge.New(outName, float64(memory.Free))
			case MetricCPUUtilization:
				for i, value := range utilization {
					channel <- gauge.New(
						fmt.Sprintf("%s%d", outName, i+1),
						value,
					)
				}
			}
		}
	}()

	return channel, nil
}

func (collector *Collector) isset(name string) bool {
	_, ok := collector.metrics[name]
	return ok
}
