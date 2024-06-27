package api

import (
	"net/http"
)

func (container Container) GetAllMetrics(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/html")

	if err := container.templates.GetAllMetricsTemplate().Execute(writer, container.manager.GetAll()); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}
