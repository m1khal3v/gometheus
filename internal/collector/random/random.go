package random

import (
	"fmt"
	"github.com/m1khal3v/gometheus/internal/logger"
	_metric "github.com/m1khal3v/gometheus/internal/metric"
	"math/rand/v2"
)

type Collector struct {
	Min float64
	Max float64
}

type ErrMinGreaterThanMax struct {
	Min float64
	Max float64
}

func (err ErrMinGreaterThanMax) Error() string {
	return fmt.Sprintf("Min=%v can`t be greater than Max=%v", err.Min, err.Max)
}

func newMinGreaterThanMaxError(min, max float64) ErrMinGreaterThanMax {
	return ErrMinGreaterThanMax{
		Min: min,
		Max: max,
	}
}

func NewCollector(min, max float64) (*Collector, error) {
	if max < min {
		return nil, newMinGreaterThanMaxError(min, max)
	}

	return &Collector{
		Min: min,
		Max: max,
	}, nil
}

func (collector *Collector) Collect() ([]*_metric.Metric, error) {
	metric, err := _metric.NewMetric(
		_metric.TypeGauge,
		"RandomValue",
		// так как rand.Float64 возвращает значение от 0 до 1 и не поддерживает Min/Max
		// исправляем это домножая значение на разницу Max и Min и добавляя к результату Min
		// итоговое значение будет в диапазоне от Min до Max включая оба значения
		rand.Float64()*(collector.Max-collector.Min)+collector.Min,
	)
	if err != nil {
		logger.Logger.Panic(err.Error())
	}

	return []*_metric.Metric{metric}, nil
}
