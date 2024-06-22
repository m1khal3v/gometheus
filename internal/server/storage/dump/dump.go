package dump

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/factory"
	"github.com/m1khal3v/gometheus/internal/server/storage"
	"os"
	"sync"
	"time"
)

type Storage struct {
	storage       storage.Storage
	filepath      string
	storeInterval uint32
	mutex         sync.Mutex
}

func New(storage storage.Storage, filepath string, storeInterval uint32, restore bool) *Storage {
	if storage == nil {
		logger.Logger.Fatal("Decorated storage cannot be nil")
	}

	if filepath == "" {
		logger.Logger.Fatal("Dump file path cannot be empty")
	}

	if restore {
		restoreFromFile(storage, filepath)
	}

	decorator := &Storage{
		storage:       storage,
		filepath:      filepath,
		storeInterval: storeInterval,
	}

	if storeInterval > 0 {
		go func(storage *Storage) {
			for range time.Tick(time.Duration(storage.storeInterval) * time.Second) {
				storage.Dump()
			}
		}(decorator)
	}

	return decorator
}

func (storage *Storage) Get(name string) metric.Metric {
	return storage.storage.Get(name)
}

func (storage *Storage) GetAll() map[string]metric.Metric {
	return storage.storage.GetAll()
}

func (storage *Storage) Save(metric metric.Metric) {
	storage.storage.Save(metric)

	if storage.storeInterval == 0 {
		storage.Dump()
	}
}

type anonymousMetric struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (storage *Storage) Dump() {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	file, err := os.OpenFile(storage.filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		logger.Logger.Warn(fmt.Sprintf("Failed to open dump file %s for writing: %s", storage.filepath, err.Error()))
		return
	}
	defer file.Close()

	for _, metric := range storage.storage.GetAll() {
		anonymousMetric := anonymousMetric{
			Type:  metric.GetType(),
			Name:  metric.GetName(),
			Value: metric.GetStringValue(),
		}
		jsonMetric, err := json.Marshal(anonymousMetric)
		if err != nil {
			logger.Logger.Warn(fmt.Sprintf("Failed to marshal json '%v': %s. Skip", anonymousMetric, err.Error()))
			continue
		}

		if _, err := file.Write(jsonMetric); err != nil {
			logger.Logger.Warn(fmt.Sprintf("Failed to write json '%s': %s. Skip", jsonMetric, err.Error()))
			continue
		}

		if _, err := file.WriteString("\n"); err != nil {
			logger.Logger.Fatal(fmt.Sprintf("Failed to write new line: %s", err.Error()))
			return
		}
	}
}

func restoreFromFile(storage storage.Storage, filepath string) {
	file, err := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		logger.Logger.Warn(fmt.Sprintf("Failed to open dump file %s for reading: %s", filepath, err.Error()))
		return
	}
	defer file.Close()

	reader := bufio.NewScanner(file)
	for reader.Scan() {
		anonymousMetric := &anonymousMetric{}
		if err := json.Unmarshal(reader.Bytes(), anonymousMetric); err != nil {
			logger.Logger.Warn(fmt.Sprintf("Failed to unmarshal json '%s': %s. Skip", reader.Text(), err.Error()))
			continue
		}

		metric, err := factory.New(anonymousMetric.Type, anonymousMetric.Name, anonymousMetric.Value)
		if err != nil {
			logger.Logger.Warn(fmt.Sprintf("Failed to create metric '%s': %s. Skip", anonymousMetric.Name, err.Error()))
			continue
		}

		storage.Save(metric)
	}
}
