package api

import (
	"net/http"
)

func (container Container) GetAllMetrics(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/html")

	metrics, err := container.manager.GetAll()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := container.templates.GetAllMetricsTemplate().Execute(writer, metrics); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}
