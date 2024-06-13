package route

import (
	"errors"
	_storage "github.com/m1khal3v/gometheus/internal/storage"
	"net/http"
)

func (routeContainer Container) GetMetric(writer http.ResponseWriter, request *http.Request) {
	metricType := request.PathValue("type")
	metricName := request.PathValue("name")

	metric, err := routeContainer.Storage.Get(metricName)
	if err != nil {
		if errors.As(err, &_storage.ErrMetricNotFound{}) {
			writer.WriteHeader(http.StatusNotFound)
		} else {
			writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if metric.GetType() != metricType {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(metric.GetValue()))
}
