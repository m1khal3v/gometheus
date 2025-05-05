// Package gopsutil
// collector for shirou/gopsutil package metrics
package gopsutil

import (
	"errors"
	"fmt"
	"slices"

	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type (
	interest  string
	MetricMap map[interest]string
)

const (
	TotalMemory    interest = "memTotal"
	FreeMemory     interest = "memFree"
	CPUUtilization interest = "cpuUtilization"
)

func getAvailableMetrics() []interest {
	return []interest{
		TotalMemory,
		FreeMemory,
		CPUUtilization,
	}
}

type Collector struct {
	metrics     MetricMap
	channelSize uint16
}

type InvalidMetricNameError struct {
	Name string
}

func (err InvalidMetricNameError) Error() string {
	return fmt.Sprintf("invalid metric name: %s", err.Name)
}

func newErrInvalidMetricName(name interest) error {
	return &InvalidMetricNameError{
		Name: string(name),
	}
}

type NonUniqueOutNameError struct {
	Name string
}

func (err NonUniqueOutNameError) Error() string {
	return fmt.Sprintf("out name '%s' is not unique", err.Name)
}

func newErrNonUniqueOutName(name string) error {
	return &NonUniqueOutNameError{
		Name: name,
	}
}

var ErrEmptyMetrics = errors.New("metrics are empty")

func New(metrics map[interest]string) (*Collector, error) {
	if len(metrics) == 0 {
		return nil, ErrEmptyMetrics
	}

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

		if name == CPUUtilization {
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

	if collector.isset(TotalMemory) || collector.isset(FreeMemory) {
		memory, err = mem.VirtualMemory()
		if err != nil {
			return nil, err
		}
	}
	if collector.isset(CPUUtilization) {
		utilization, err = cpu.Percent(0, true)
		if err != nil {
			return nil, err
		}
	}

	go func() {
		defer close(channel)

		for name, outName := range collector.metrics {
			switch name {
			case TotalMemory:
				channel <- gauge.New(outName, float64(memory.Total))
			case FreeMemory:
				channel <- gauge.New(outName, float64(memory.Free))
			case CPUUtilization:
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

func (collector *Collector) isset(name interest) bool {
	_, ok := collector.metrics[name]
	return ok
}
