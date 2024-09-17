package api

import (
	"net/http"
	"strings"

	"github.com/m1khal3v/gometheus/internal/common/metric/factory"
)

func (container Container) SaveMetric(writer http.ResponseWriter, request *http.Request) {
	metricType := request.PathValue("type")
	metricName := request.PathValue("name")
	metricValue := request.PathValue("value")

	if strings.TrimSpace(metricType) == "" {
		WriteJSONErrorResponse(http.StatusBadRequest, writer, "Empty type received", nil)
		return
	}

	if strings.TrimSpace(metricName) == "" {
		WriteJSONErrorResponse(http.StatusBadRequest, writer, "Empty name received", nil)
		return
	}

	metric, err := factory.New(metricType, metricName, metricValue)
	if err != nil {
		WriteJSONErrorResponse(http.StatusBadRequest, writer, "Invalid metric data received", err)
		return
	}

	if _, err := container.manager.Save(request.Context(), metric); err != nil {
		WriteJSONErrorResponse(http.StatusInternalServerError, writer, "Can`t save metric", err)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
