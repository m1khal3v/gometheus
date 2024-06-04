package route

import (
	"errors"
	storages "github.com/m1khal3v/gometheus/internal/storage"
	"github.com/m1khal3v/gometheus/internal/store"
	"net/http"
	"strings"
)

func (routeContainer Container) GetMetric(writer http.ResponseWriter, request *http.Request) {
	// Разбираем путь
	metricType := request.PathValue("type")
	metricName := request.PathValue("name")

	// Проверяем, что тип не пустой
	if strings.TrimSpace(metricType) == "" || nil != store.ValidateMetricType(metricType) {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, err := routeContainer.Storage.Get(metricName)
	// Проверяем что метрика существует и передан верный тип
	if err != nil {
		if errors.As(err, &storages.MetricNotFoundError{}) {
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
