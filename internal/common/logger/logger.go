package logger

import (
	"go.uber.org/zap"
)

var Logger = zap.NewNop()
var initialized = false

type ErrAlreadyInitialized struct {
}

func (err ErrAlreadyInitialized) Error() string {
	return "Logger already initialized"
}

func Init(name, level string) error {
	defer func() { initialized = true }()
	if initialized {
		return ErrAlreadyInitialized{}
	}

	atomicLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	config := zap.NewProductionConfig()
	config.Level = atomicLevel
	logger, err := config.Build()
	if err != nil {
		return err
	}

	Logger = logger.Named(name)

	return nil
}
