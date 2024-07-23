package pgsql

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/m1khal3v/gometheus/internal/common/logger"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/factory"
	store "github.com/m1khal3v/gometheus/internal/server/storage"
	"github.com/m1khal3v/gometheus/pkg/generator"
	"github.com/m1khal3v/gometheus/pkg/retry"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"sync"
	"time"
)

const (
	getStatement  = "get"
	saveStatement = "save"
)

type Storage struct {
	db         *sql.DB
	mutex      *sync.Mutex
	closed     bool
	statements map[string]*sql.Stmt
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

	storage := &Storage{
		db:     db,
		mutex:  &sync.Mutex{},
		closed: false,
	}
	storage.prepareStatements()

	return storage
}

func (storage *Storage) Get(ctx context.Context, name string) (metric.Metric, error) {
	var metricType, metricValue string

	err := retry.Retry(time.Second, 5*time.Second, 4, 2, func() error {
		return storage.statements[getStatement].QueryRowContext(ctx, name).Scan(&metricType, &metricValue)
	}, storage.isRetryableError)

	if err != nil {
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

func (storage *Storage) GetAll(ctx context.Context) (<-chan metric.Metric, error) {
	if err := storage.checkStorageClosed(); err != nil {
		return nil, err
	}

	var rows *sql.Rows
	err := retry.Retry(time.Second, 5*time.Second, 4, 2, func() error {
		var err error
		rows, err = storage.db.QueryContext(ctx, "SELECT type, name, value::VARCHAR FROM metric")
		if err != nil {
			return err
		}
		return rows.Err()
	}, storage.isRetryableError)
	if err != nil {
		return nil, err
	}

	return generator.NewFromFunctionWithContext(ctx, func() (metric.Metric, bool) {
		if !rows.Next() {
			return nil, false
		}

		var metricType, metricName, metricValue string
		if err := rows.Scan(&metricType, &metricName, &metricValue); err != nil {
			logger.Logger.Error("Failed to scan row", zap.Error(err))
			if err := rows.Close(); err != nil {
				logger.Logger.Error("Failed to close rows", zap.Error(err))
			}

			return nil, false
		}

		metric, err := factory.New(metricType, metricName, metricValue)
		if err != nil {
			logger.Logger.Error("Failed to create metric", zap.Error(err))
			if err := rows.Close(); err != nil {
				logger.Logger.Error("Failed to close rows", zap.Error(err))
			}

			return nil, false
		}

		return metric.Clone(), true
	}), nil
}

func (storage *Storage) Save(ctx context.Context, metric metric.Metric) error {
	if err := storage.checkStorageClosed(); err != nil {
		return err
	}

	err := retry.Retry(time.Second, 5*time.Second, 4, 2, func() error {
		_, err := storage.statements[saveStatement].ExecContext(ctx, metric.Type(), metric.Name(), metric.StringValue())

		return err
	}, storage.isRetryableError)

	if err != nil {
		return err
	}

	return nil
}

func (storage *Storage) SaveBatch(ctx context.Context, metrics []metric.Metric) error {
	if err := storage.checkStorageClosed(); err != nil {
		return err
	}

	err := retry.Retry(time.Second, 5*time.Second, 4, 2, func() error {
		transaction, err := storage.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		statement := transaction.StmtContext(ctx, storage.statements[saveStatement])

		for _, metric := range metrics {
			if _, err := statement.ExecContext(ctx, metric.Type(), metric.Name(), metric.StringValue()); err != nil {
				if rollbackErr := transaction.Rollback(); rollbackErr != nil {
					return errors.Join(err, rollbackErr)
				}

				return err
			}
		}

		return transaction.Commit()
	}, storage.isRetryableError)

	return err
}

func (storage *Storage) Ping(ctx context.Context) error {
	if err := storage.checkStorageClosed(); err != nil {
		return err
	}

	if err := storage.db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

func (storage *Storage) Close(ctx context.Context) error {
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

func (storage *Storage) Reset(ctx context.Context) error {
	if err := storage.checkStorageClosed(); err != nil {
		return err
	}

	return retry.Retry(time.Second, 5*time.Second, 4, 2, func() error {
		_, err := storage.db.ExecContext(ctx, "TRUNCATE TABLE metric")
		return err
	}, storage.isRetryableError)
}

func (storage *Storage) checkStorageClosed() error {
	if storage.closed {
		return store.ErrStorageClosed
	}

	return nil
}

func (storage *Storage) isRetryableError(err error) bool {
	var pgsqlErr *pgconn.PgError
	if !errors.As(err, &pgsqlErr) {
		return false
	}

	return pgerrcode.IsConnectionException(pgsqlErr.Code) ||
		pgerrcode.IsInsufficientResources(pgsqlErr.Code) ||
		pgerrcode.IsSystemError(pgsqlErr.Code) ||
		pgerrcode.IsInternalError(pgsqlErr.Code) ||
		pgerrcode.IsTransactionRollback(pgsqlErr.Code)
}

func (storage *Storage) prepareStatements() {
	items := []struct {
		name string
		sql  string
	}{
		{
			name: getStatement,
			sql:  "SELECT type, value::VARCHAR FROM metric WHERE name = $1",
		},
		{
			name: saveStatement,
			sql: `
			INSERT INTO metric (type, name, value) 
			VALUES ($1, $2, $3::DOUBLE PRECISION)
			ON CONFLICT (name) DO UPDATE
			SET type = $1, value = $3::DOUBLE PRECISION`,
		},
	}
	storage.statements = make(map[string]*sql.Stmt, len(items))

	for _, item := range items {
		var err error
		storage.statements[item.name], err = storage.db.Prepare(item.sql)
		if err != nil {
			panic(err)
		}
	}
}
