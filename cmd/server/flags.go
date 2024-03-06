package main

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	flagRunAddr     string
	storeInterval   int
	fileStoragePath string
	restore         bool
}

var conf Config

func parseFlags() {
	//conf := &Config{}

	flag.StringVar(&conf.flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&conf.fileStoragePath, "f", "/tmp/metrics-db.json", "File to save metrics")
	flag.IntVar(&conf.storeInterval, "i", 30, "Interval to save metrics")
	flag.BoolVar(&conf.restore, "r", true, "")

	//flag.StringVar(&conf.flagRunAddr, "a", "0.0.0.0:8080", "address and port to run server")
	//flag.StringVar(&conf.fileStoragePath, "f", "./metrics-db.json", "File to save metrics")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		conf.flagRunAddr = envRunAddr
	}

	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		conf.storeInterval, _ = strconv.Atoi(envStoreInterval)
	}

	if envStoragePath := os.Getenv("FILE_STORAGE_PATH"); envStoragePath != "" {
		conf.fileStoragePath = envStoragePath
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		conf.restore, _ = strconv.ParseBool(envRestore)
	}

	//return conf
}
