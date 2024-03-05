package main

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	flagRunAddr        string
	flagReportInterval int
	flagRuntime        int
}

var conf Config

func parseFlags() {
	flag.StringVar(&conf.flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&conf.flagReportInterval, "r", 10, "address and port to run server")
	flag.IntVar(&conf.flagRuntime, "p", 2, "address and port to run server")

	//flag.StringVar(&conf.flagRunAddr, "a", "0.0.0.0:8080", "address and port to run server")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		conf.flagRunAddr = envRunAddr
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		conf.flagReportInterval, _ = strconv.Atoi(envReportInterval)
	}

	if envRuntime := os.Getenv("RUNTIME"); envRuntime != "" {
		conf.flagRuntime, _ = strconv.Atoi(envRuntime)
	}
}
