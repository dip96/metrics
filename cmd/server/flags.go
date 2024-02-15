package main

import (
	"flag"
	"os"
)

type Config struct {
	flagRunAddr string
}

var conf Config

func parseFlags() {
	flag.StringVar(&conf.flagRunAddr, "a", "localhost:8080", "address and port to run server")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		conf.flagRunAddr = envRunAddr
	}
}
