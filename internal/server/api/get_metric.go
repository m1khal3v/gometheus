package api

import (
	"net/http"
)

func (container Container) GetMetric(writer http.ResponseWriter, request *http.Request) {
	metricType := request.PathValue("type")
	metricName := request.PathValue("name")

	metric, _ := container.manager.Get(metricType, metricName)
	if metric == nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	writer.Header().Set("Content-Type", "text/plain")
	if _, err := writer.Write([]byte(metric.StringValue())); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
}
