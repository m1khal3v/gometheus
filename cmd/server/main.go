package main

import (
	"github.com/m1khal3v/gometheus/internal/logger"
	"github.com/m1khal3v/gometheus/internal/server"
)

func main() {
	defer logger.Logger.Sync()
	server.Start()
}
