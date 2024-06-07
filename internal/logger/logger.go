// Package logger Заготовка под нормальное логирование
package logger

import "go.uber.org/zap"

var Logger = zap.Must(zap.NewProduction())
