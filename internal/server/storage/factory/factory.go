package factory

import (
	"context"
	"fmt"
	"github.com/m1khal3v/gometheus/internal/server/storage"
	"github.com/m1khal3v/gometheus/internal/server/storage/kind/dump"
	"github.com/m1khal3v/gometheus/internal/server/storage/kind/memory"
	"github.com/m1khal3v/gometheus/internal/server/storage/kind/pgsql"
)

type ErrUnknownDriver struct {
	Driver string
}

func (err ErrUnknownDriver) Error() string {
	return fmt.Sprintf("driver '%s' is not defined", err.Driver)
}

func newErrUnknownDriver(driver string) error {
	return &ErrUnknownDriver{
		Driver: driver,
	}
}

func New(ctx context.Context, fileStoragePath, databaseDriver, databaseDSN string, storeInterval uint32, restore bool) (storage.Storage, error) {
	var storage storage.Storage = memory.New()

	if databaseDSN != "" && databaseDriver != "" {
		var err error
		storage, err = newDBStorage(databaseDriver, databaseDSN)
		if err != nil {
			return nil, err
		}
	}

	if fileStoragePath != "" {
		var err error
		storage, err = dump.New(ctx, storage, fileStoragePath, storeInterval, restore)
		if err != nil {
			return nil, err
		}
	}

	return storage, nil
}

func newDBStorage(databaseDriver, databaseDSN string) (storage.Storage, error) {
	switch databaseDriver {
	case "pgx":
		return pgsql.New(databaseDSN), nil
	default:
		return nil, newErrUnknownDriver(databaseDriver)
	}
}
