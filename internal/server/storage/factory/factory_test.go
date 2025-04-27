package factory

import (
	"context"
	"errors"
	"testing"

	"github.com/m1khal3v/gometheus/internal/server/storage/kind/dump"
	"github.com/m1khal3v/gometheus/internal/server/storage/kind/memory"
	"github.com/stretchr/testify/assert"
)

func TestNewMemoryStorage(t *testing.T) {
	ctx := context.Background()

	storage, err := New(ctx, "", "", "", 0, false)

	assert.NoError(t, err)
	assert.IsType(t, &memory.Storage{}, storage)
}

func TestNewDumpStorage(t *testing.T) {
	ctx := context.Background()
	fileStoragePath := "/tmp/test.db"
	storeInterval := uint32(60)
	restore := true

	storage, err := New(ctx, fileStoragePath, "", "", storeInterval, restore)

	assert.NoError(t, err)
	assert.IsType(t, &dump.Storage{}, storage)
}

func TestNewUnknownDriver(t *testing.T) {
	ctx := context.Background()
	databaseDriver := "unknown"
	databaseDSN := "unknown://test"

	storage, err := New(ctx, "", databaseDriver, databaseDSN, 0, false)

	assert.Nil(t, storage)
	assert.Error(t, err)

	var unknownDriverErr *ErrUnknownDriver
	isUnknownDriverErr := errors.As(err, &unknownDriverErr)
	assert.True(t, isUnknownDriverErr)
	assert.Equal(t, databaseDriver, unknownDriverErr.Driver)
}
