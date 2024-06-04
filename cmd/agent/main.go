package main

import (
	"github.com/m1khal3v/gometheus/internal/agent"
	"github.com/m1khal3v/gometheus/internal/logger"
)

func main() {
	defer logger.Logger.Sync()
	agent.Start()
}
