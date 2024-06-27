package api

import (
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"net/http"
)

func (container Container) GetMetric(writer http.ResponseWriter, request *http.Request) {
	metricType := request.PathValue("type")
	metricName := request.PathValue("name")

	metric := container.manager.Get(metricType, metricName)
	if metric == nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	writer.Header().Set("Content-Type", "text/plain")
	_, err := writer.Write([]byte(metric.StringValue()))
	if err != nil {
		logger.Logger.Error(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}
