package pgsql

import (
	"database/sql"
	"embed"
	"errors"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/factory"
	store "github.com/m1khal3v/gometheus/internal/server/storage"
	"github.com/m1khal3v/gometheus/pkg/generator"
	"github.com/pressly/goose/v3"
	"sync"
)

type Storage struct {
	db     *sql.DB
	mutex  *sync.Mutex
	closed bool
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

func New(databaseDSN string) *Storage {
	db, err := goose.OpenDBWithDriver("pgx", databaseDSN)
	if err != nil {
		panic(err)
	}

	goose.SetBaseFS(embedMigrations)
	if err := goose.Up(db, "migrations"); err != nil {
		panic(err)
	}

	return &Storage{
		db:     db,
		mutex:  &sync.Mutex{},
		closed: false,
	}
}

func (storage *Storage) Get(name string) (metric.Metric, error) {
	var metricType, metricValue string
	row := storage.db.QueryRow("SELECT type, value FROM metric WHERE name = $1", name)
	if err := row.Scan(&metricType, &metricValue); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	metric, err := factory.New(metricType, name, metricValue)
	if err != nil {
		return nil, err
	}

	return metric, nil
}

func (storage *Storage) GetAll() (<-chan metric.Metric, error) {
	if err := storage.checkStorageClosed(); err != nil {
		return nil, err
	}

	rows, err := storage.db.Query("SELECT type, name, value FROM metric")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return generator.NewFromFunction(func() (metric.Metric, bool) {
		if !rows.Next() {
			return nil, false
		}

		var metricType, metricName, metricValue string
		if err := rows.Scan(&metricType, &metricName, &metricValue); err != nil {
			logger.Logger.Error(err.Error())

			return nil, false
		}

		metric, err := factory.New(metricType, metricName, metricValue)
		if err != nil {
			logger.Logger.Error(err.Error())

			return nil, false
		}

		return metric.Clone(), true
	}), nil
}

func (storage *Storage) Save(metric metric.Metric) error {
	if err := storage.checkStorageClosed(); err != nil {
		return err
	}

	if _, err := storage.db.Exec(
		`
		INSERT INTO metric (type, name, value) 
		VALUES ($1, $2, $3::DOUBLE PRECISION)
		ON CONFLICT (name) DO UPDATE
		SET type = $1, value = $3::DOUBLE PRECISION
		`,
		metric.Type(),
		metric.Name(),
		metric.StringValue(),
	); err != nil {
		return err
	}

	return nil
}

func (storage *Storage) SaveBatch(metrics []metric.Metric) error {
	if err := storage.checkStorageClosed(); err != nil {
		return err
	}

	transaction, err := storage.db.Begin()
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		if _, err := transaction.Exec(
			`
			INSERT INTO metric (type, name, value) 
			VALUES ($1, $2, $3::DOUBLE PRECISION)
			ON CONFLICT (name) DO UPDATE
			SET type = $1, value = $3::DOUBLE PRECISION
			`,
			metric.Type(),
			metric.Name(),
			metric.StringValue(),
		); err != nil {
			if err := transaction.Rollback(); err != nil {
				return err
			}

			return err
		}
	}

	if err := transaction.Commit(); err != nil {
		return err
	}

	return nil
}

func (storage *Storage) Ok() bool {
	return !storage.closed && storage.db.Ping() == nil
}

func (storage *Storage) Close() error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	if storage.closed {
		return store.ErrStorageClosed
	}

	if err := storage.db.Close(); err != nil {
		return err
	}

	storage.closed = true
	return nil
}

func (storage *Storage) Reset() error {
	if err := storage.checkStorageClosed(); err != nil {
		return err
	}

	if _, err := storage.db.Exec("TRUNCATE TABLE metric"); err != nil {
		return err
	}

	return nil
}

func (storage *Storage) checkStorageClosed() error {
	if storage.closed {
		return store.ErrStorageClosed
	}

	return nil
}
