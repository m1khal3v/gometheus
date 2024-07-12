package dump

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/factory"
	"github.com/m1khal3v/gometheus/internal/server/storage"
	"github.com/m1khal3v/gometheus/pkg/retry"
	"os"
	"sync"
	"syscall"
	"time"
)

type Storage struct {
	storage  storage.Storage
	filepath string
	sync     bool
	mutex    sync.Mutex
}

func New(ctx context.Context, storage storage.Storage, filepath string, storeInterval uint32, restore bool) (*Storage, error) {
	if storage == nil {
		panic("Decorated storage cannot be nil")
	}

	if filepath == "" {
		panic("Dump file path cannot be empty")
	}

	decorator := &Storage{
		storage:  storage,
		filepath: filepath,
		sync:     storeInterval == 0,
	}

	if restore {
		if err := decorator.restoreFromFile(ctx); err != nil {
			return nil, err
		}
	}

	if !decorator.sync {
		go func() {
			ticker := time.NewTicker(time.Duration(storeInterval) * time.Second)

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					decorator.mustDump(ctx)
				}
			}
		}()
	}

	return decorator, nil
}

func (storage *Storage) Get(ctx context.Context, name string) (metric.Metric, error) {
	return storage.storage.Get(ctx, name)
}

func (storage *Storage) GetAll(ctx context.Context) (<-chan metric.Metric, error) {
	return storage.storage.GetAll(ctx)
}

func (storage *Storage) Save(ctx context.Context, metric metric.Metric) error {
	if err := storage.storage.Save(ctx, metric); err != nil {
		return err
	}

	if storage.sync {
		storage.mustDump(ctx)
	}

	return nil
}

func (storage *Storage) SaveBatch(ctx context.Context, metrics []metric.Metric) error {
	if err := storage.storage.SaveBatch(ctx, metrics); err != nil {
		return err
	}

	if storage.sync {
		storage.mustDump(ctx)
	}

	return nil
}

func (storage *Storage) Ping(ctx context.Context) error {
	return storage.storage.Ping(ctx)
}

func (storage *Storage) Close(ctx context.Context) error {
	if err := storage.dump(ctx); err != nil {
		return err
	}

	return storage.storage.Close(ctx)
}

func (storage *Storage) Reset(ctx context.Context) error {
	return storage.storage.Reset(ctx)
}

type anonymousMetric struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (storage *Storage) mustDump(ctx context.Context) {
	if err := storage.dump(ctx); err != nil {
		panic(err)
	}
}

func (storage *Storage) dump(ctx context.Context) error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	var file *os.File
	err := retry.Retry(time.Second, 5*time.Second, 4, 2, func() error {
		var err error
		file, err = os.OpenFile(storage.filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		return err
	}, storage.isRetryableError)
	if err != nil {
		return err
	}

	defer file.Close()

	allMetrics, err := storage.storage.GetAll(ctx)
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

func (storage *Storage) restoreFromFile(ctx context.Context) error {
	file, err := os.OpenFile(storage.filepath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := storage.storage.Reset(ctx); err != nil {
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

		if err := storage.storage.Save(ctx, metric); err != nil {
			return err
		}
	}

	return nil
}

func (storage *Storage) isRetryableError(err error) bool {
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		return false
	}

	return errors.Is(pathErr.Err, syscall.EAGAIN) ||
		errors.Is(pathErr.Err, syscall.EBUSY) ||
		errors.Is(pathErr.Err, syscall.ENFILE) ||
		errors.Is(pathErr.Err, syscall.EMFILE)
}
