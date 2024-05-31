package route

import (
	"github.com/m1khal3v/gometheus/internal/storage"
	"net/http"
	"strings"
)

func (routeContainer Container) SaveMetric(writer http.ResponseWriter, request *http.Request) {
	// Разрешаем только POST
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Проверяем, что передан валидный Content-Type
	// if request.Header.Get("Content-Type") != "text/plain" {
	//	writer.WriteHeader(http.StatusBadRequest)
	//	return
	// }

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

	// Проверяем, что передано непустое значение метрики
	if strings.TrimSpace(metricValue) == "" {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// Создаем метрику и обрабатываем ошибки
	metric, err := storage.NewMetric(metricType, metricName, metricValue)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// Сохраняем метрику и обрабатываем ошибки
	if routeContainer.Storage.Save(metric) != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	}
}
