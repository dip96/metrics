package config

import (
	"flag"
	"os"
	"strconv"
)

type Server struct {
	FlagRunAddr     string
	StoreInterval   int
	FileStoragePath string
	Restore         bool
}

var ServerConfig *Server

func LoadServer() *Server {
	if ServerConfig == nil {
		var cfg = Server{}

		flag.StringVar(&cfg.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
		flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/metrics-db.json", "File to save metrics")
		flag.IntVar(&cfg.StoreInterval, "i", 5, "Interval to save metrics")
		flag.BoolVar(&cfg.Restore, "r", true, "")

		//flag.StringVar(&cfg.flagRunAddr, "a", "0.0.0.0:8080", "address and port to run server")
		//flag.StringVar(&cfg.fileStoragePath, "f", "./metrics-db.json", "File to save metrics")

		flag.Parse()

		if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
			cfg.FlagRunAddr = envRunAddr
		}

		if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
			cfg.StoreInterval, _ = strconv.Atoi(envStoreInterval)
		}

		if envStoragePath := os.Getenv("FILE_STORAGE_PATH"); envStoragePath != "" {
			cfg.FileStoragePath = envStoragePath
		}

		if envRestore := os.Getenv("RESTORE"); envRestore != "" {
			cfg.Restore, _ = strconv.ParseBool(envRestore)
		}

		ServerConfig = &cfg
	}

	return ServerConfig
}
