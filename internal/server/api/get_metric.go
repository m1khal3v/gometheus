package api

import (
	"net/http"
)

func (container Container) GetMetric(writer http.ResponseWriter, request *http.Request) {
	metricType := request.PathValue("type")
	metricName := request.PathValue("name")

	metric, err := container.manager.Get(request.Context(), metricType, metricName)
	if err != nil {
		container.writeErrorResponse(http.StatusInternalServerError, writer, "Can`t get metric", err)
		return
	}
	if metric == nil {
		container.writeErrorResponse(http.StatusNotFound, writer, "Metric not found", nil)
		return
	}

	writer.Header().Set("Content-Type", "text/plain")
	if _, err := writer.Write([]byte(metric.StringValue())); err != nil {
		container.writeErrorResponse(http.StatusInternalServerError, writer, "Can`t write response", err)
		return
	}
}
