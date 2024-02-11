package main

import (
	"flag"
	"os"
	"strconv"
)

// неэкспортированная переменная flagRunAddr содержит адрес и порт для запуска сервера
var flagRunAddr string
var flagReportInterval int
var flagRuntime int

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func parseFlags() {
	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&flagRunAddr, "a", "localhost:80", "address and port to run server")
	flag.IntVar(&flagReportInterval, "r", 10, "address and port to run server")
	flag.IntVar(&flagRuntime, "p", 2, "address and port to run server")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()

	if envRunAddr := os.Getenv("RUN_ADDR"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		flagReportInterval, _ = strconv.Atoi(envReportInterval)
	}

	if envRuntime := os.Getenv("RUNTIME"); envRuntime != "" {
		flagRuntime, _ = strconv.Atoi(envRuntime)
	}
}
