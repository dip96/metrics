package main

import (
	"flag"
	"os"
)

type Config struct {
	flagRunAddr string
	pathForLogs string
}

var conf Config

func NewConfig() *Config {
	conf := &Config{}
	flag.StringVar(&conf.flagRunAddr, "a", "localhost:8080", "address and port to run server")
	//TODO изменить на программное получение абсолютного пути к корневой директории проекта
	//flag.StringVar(&conf.pathForLogs, "p", "/home/dip96/go_project/metrics/requests.log", "path for logs file")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		conf.flagRunAddr = envRunAddr
	}

	//if envRunAddr := os.Getenv("PATH_FOR_FILE_LOGS"); envRunAddr != "" {
	//	conf.flagRunAddr = envRunAddr
	//}

	return conf
}

func parseFlags() {
	flag.StringVar(&conf.flagRunAddr, "a", "127.0.0.1:8080", "address and port to run server")
	//TODO изменить на программное получение абсолютного пути к корневой директории проекта
	flag.StringVar(&conf.pathForLogs, "a", "/home/dip96/go_project/metrics/requests.log", "address and port to run server")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		conf.flagRunAddr = envRunAddr
	}
}
