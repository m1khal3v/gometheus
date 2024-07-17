package api

import (
	"net/http"
)

func (container Container) GetAllMetrics(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/html")

	metrics, err := container.manager.GetAll(request.Context())
	if err != nil {
		WriteJSONErrorResponse(http.StatusInternalServerError, writer, "Can`t get metrics", err)
		return
	}

	if err := container.templates.ExecuteAllMetricsTemplate(writer, metrics); err != nil {
		WriteJSONErrorResponse(http.StatusInternalServerError, writer, "Can`t use page template", err)
		return
	}
}
