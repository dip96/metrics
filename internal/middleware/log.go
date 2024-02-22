package middleware

import (
	log "github.com/sirupsen/logrus"
	"os"
)

var logFile *os.File

func InitLogger() {
	//TODO найти способ получить из конфигов значение - pathForLogs
	file, err := os.OpenFile("/home/dip96/go_project/metrics/requests.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(file)
	//log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
}

func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}
