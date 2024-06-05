package main

import (
	"github.com/m1khal3v/gometheus/internal/logger"
	"github.com/m1khal3v/gometheus/internal/server"
	flag "github.com/spf13/pflag"
)

var endpoint string

func init() {
	flag.StringVarP(&endpoint, "endpoint", "e", "localhost:8080", "endpoint of gometheus server")
}

func main() {
	defer logger.Logger.Sync()
	flag.Parse()
	server.Start(endpoint)
}
