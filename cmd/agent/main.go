package main

import (
	"github.com/m1khal3v/gometheus/internal/agent"
	"github.com/m1khal3v/gometheus/internal/logger"
	flag "github.com/spf13/pflag"
)

var endpoint string
var pollInterval uint32
var reportInterval uint32

func init() {
	flag.StringVarP(&endpoint, "endpoint", "a", "localhost:8080", "endpoint of gometheus server")
	flag.Uint32VarP(&pollInterval, "poll-interval", "p", 2, "interval of collecting metrics")
	flag.Uint32VarP(&reportInterval, "report-interval", "r", 10, "interval of reporting metrics")
}

func main() {
	defer logger.Logger.Sync()
	flag.Parse()
	agent.Start(endpoint, pollInterval, reportInterval)
}
