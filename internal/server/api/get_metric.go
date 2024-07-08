package api

import (
	"errors"
	"github.com/m1khal3v/gometheus/internal/server/manager"
	"net/http"
)

func (container Container) GetMetric(writer http.ResponseWriter, request *http.Request) {
	metricType := request.PathValue("type")
	metricName := request.PathValue("name")

	metric, err := container.manager.Get(metricType, metricName)
	if err != nil {
		if errors.As(err, &manager.ErrMetricNotFound{}) {
			writer.WriteHeader(http.StatusNotFound)
		} else {
			writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	writer.Header().Set("Content-Type", "text/plain")
	if _, err := writer.Write([]byte(metric.StringValue())); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
}
