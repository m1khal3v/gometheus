package api

import (
	"net/http"
)

func (container Container) PingStorage(writer http.ResponseWriter, request *http.Request) {
	if container.manager.IsStorageOk() {
		writer.WriteHeader(http.StatusOK)
	} else {
		writer.WriteHeader(http.StatusInternalServerError)
	}
}
