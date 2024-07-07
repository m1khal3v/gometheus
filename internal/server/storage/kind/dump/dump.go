package dump

import (
	"bufio"
	"encoding/json"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/factory"
	"github.com/m1khal3v/gometheus/internal/server/storage"
	"os"
	"sync"
	"time"
)

type Storage struct {
	storage  storage.Storage
	filepath string
	sync     bool
	mutex    sync.Mutex
}

func New(storage storage.Storage, filepath string, storeInterval uint32, restore bool) (*Storage, error) {
	if storage == nil {
		panic("Decorated storage cannot be nil")
	}

	if filepath == "" {
		panic("Dump file path cannot be empty")
	}

	if restore {
		if err := restoreFromFile(storage, filepath); err != nil {
			return nil, err
		}
	}

	decorator := &Storage{
		storage:  storage,
		filepath: filepath,
		sync:     storeInterval == 0,
	}

	if !decorator.sync {
		go func() {
			for range time.Tick(time.Duration(storeInterval) * time.Second) {
				if err := decorator.Dump(); err != nil {
					panic(err)
				}
			}
		}()
	}

	return decorator, nil
}

func (storage *Storage) Get(name string) (metric.Metric, error) {
	return storage.storage.Get(name)
}

func (storage *Storage) GetAll() (<-chan metric.Metric, error) {
	return storage.storage.GetAll()
}

func (storage *Storage) Save(metric metric.Metric) error {
	if err := storage.storage.Save(metric); err != nil {
		return err
	}

	if storage.sync {
		if err := storage.Dump(); err != nil {
			panic(err)
		}
	}

	return nil
}

func (storage *Storage) SaveBatch(metrics []metric.Metric) error {
	if err := storage.storage.SaveBatch(metrics); err != nil {
		return err
	}

	if storage.sync {
		if err := storage.Dump(); err != nil {
			panic(err)
		}
	}

	return nil
}

func (storage *Storage) Ok() bool {
	return storage.storage.Ok()
}

func (storage *Storage) Close() error {
	if err := storage.Dump(); err != nil {
		return err
	}

	return storage.storage.Close()
}

func (storage *Storage) Reset() error {
	return storage.storage.Reset()
}

type anonymousMetric struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (storage *Storage) Dump() error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	file, err := os.OpenFile(storage.filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	allMetrics, err := storage.storage.GetAll()
	if err != nil {
		return err
	}

	for metric := range allMetrics {
		anonymousMetric := anonymousMetric{
			Type:  metric.Type(),
			Name:  metric.Name(),
			Value: metric.StringValue(),
		}
		jsonMetric, err := json.Marshal(anonymousMetric)
		if err != nil {
			return err
		}

		if _, err := file.Write(append(jsonMetric, "\n"...)); err != nil {
			return err
		}
	}

	return nil
}

func restoreFromFile(storage storage.Storage, filepath string) error {
	file, err := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := storage.Reset(); err != nil {
		return err
	}

	reader := bufio.NewScanner(file)
	for reader.Scan() {
		if reader.Text() == "" {
			continue
		}

		anonymousMetric := &anonymousMetric{}
		if err := json.Unmarshal(reader.Bytes(), anonymousMetric); err != nil {
			return err
		}

		metric, err := factory.New(anonymousMetric.Type, anonymousMetric.Name, anonymousMetric.Value)
		if err != nil {
			return err
		}

		if err := storage.Save(metric); err != nil {
			return err
		}
	}

	return nil
}
