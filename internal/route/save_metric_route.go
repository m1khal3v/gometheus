package route

import (
	"github.com/m1khal3v/gometheus/internal/logger"
	"github.com/m1khal3v/gometheus/internal/store"
	"net/http"
	"strings"
)

func (routeContainer Container) SaveMetric(writer http.ResponseWriter, request *http.Request) {
	// Разбираем путь
	metricType := request.PathValue("type")
	metricName := request.PathValue("name")
	metricValue := request.PathValue("value")

	// Проверяем, что тип не пустой
	if strings.TrimSpace(metricType) == "" {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// Проверяем, что название не пустое
	if strings.TrimSpace(metricName) == "" {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	// Создаем метрику и обрабатываем ошибки
	metric, err := store.NewMetric(metricType, metricName, metricValue)
	if err != nil {
		logger.Logger.Error(err.Error())
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// Сохраняем метрику и обрабатываем ошибки
	err = routeContainer.Storage.Save(metric)
	if err != nil {
		logger.Logger.Fatal(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
	}

	writer.WriteHeader(http.StatusOK)
}
