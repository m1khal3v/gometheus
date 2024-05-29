package main

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
)

const metricTypeGauge = "gauge"
const metricTypeCounter = "counter"

type memStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

func main() {
	storage := memStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
	router := http.NewServeMux()
	router.HandleFunc("/update/{type}/{name}/{value}", func(writer http.ResponseWriter, request *http.Request) {
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

		// Проверяем, что передан валидный тип и не пустое значение
		metricTypes := []string{metricTypeGauge, metricTypeCounter}
		if !slices.Contains(metricTypes, metricType) {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		// Проверяем, что название не пустое
		if "" == strings.TrimSpace(metricName) {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		// Проверяем, что передано непустое значение метрики
		if "" == strings.TrimSpace(metricValue) {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		// Проверяем, что передано валидное значение метрики и записываем его в память
		switch metricType {
		case metricTypeGauge:
			metricConvertedValue, err := strconv.ParseFloat(metricValue, 64)
			if nil != err {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			storage.Gauges[metricName] = metricConvertedValue
		case metricTypeCounter:
			metricConvertedValue, err := strconv.ParseInt(metricValue, 10, 64)
			if nil != err {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			storage.Counters[metricName] += metricConvertedValue
		}
	})

	err := http.ListenAndServe(`:8080`, router)
	if err != nil {
		panic(err)
	}
}
