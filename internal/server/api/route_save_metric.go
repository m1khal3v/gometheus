package api

import (
	"github.com/m1khal3v/gometheus/internal/common/metric/factory"
	"net/http"
	"strings"
)

func (container Container) SaveMetric(writer http.ResponseWriter, request *http.Request) {
	metricType := request.PathValue("type")
	metricName := request.PathValue("name")
	metricValue := request.PathValue("value")

	if strings.TrimSpace(metricType) == "" {
		writeJsonErrorResponse(http.StatusBadRequest, writer, "Empty type received", nil)
		return
	}

	if strings.TrimSpace(metricName) == "" {
		writeJsonErrorResponse(http.StatusBadRequest, writer, "Empty name received", nil)
		return
	}

	metric, err := factory.New(metricType, metricName, metricValue)
	if err != nil {
		writeJsonErrorResponse(http.StatusBadRequest, writer, "Invalid metric data received", err)
		return
	}

	if _, err := container.manager.Save(request.Context(), metric); err != nil {
		writeJsonErrorResponse(http.StatusInternalServerError, writer, "Can`t save metric", err)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
