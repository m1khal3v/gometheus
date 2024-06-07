package route

import (
	"errors"
	_metric "github.com/m1khal3v/gometheus/internal/metric"
	_storage "github.com/m1khal3v/gometheus/internal/storage"
	"net/http"
)

func (routeContainer Container) GetMetric(writer http.ResponseWriter, request *http.Request) {
	// Разбираем путь
	metricType := request.PathValue("type")
	metricName := request.PathValue("name")

	// Проверяем, что тип валидный
	if nil != _metric.ValidateMetricType(metricType) {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, err := routeContainer.Storage.Get(metricName)
	// Проверяем что метрика существует и передан верный тип
	if err != nil {
		if errors.As(err, &_storage.ErrMetricNotFound{}) {
			writer.WriteHeader(http.StatusNotFound)
		} else {
			writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if metric.Type != metricType {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	// Отдаем значение метрики
	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(metric.GetStringValue()))
}
