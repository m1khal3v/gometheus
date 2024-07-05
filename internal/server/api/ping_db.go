package api

import (
	"net/http"
)

func (container Container) PingDB(writer http.ResponseWriter, request *http.Request) {
	if container.db == nil {
		writer.WriteHeader(http.StatusMisdirectedRequest)
		return
	}

	if container.db.Ping() != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
