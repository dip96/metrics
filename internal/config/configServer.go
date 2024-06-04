package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// Server представляет конфигурацию сервера.
type Server struct {
	// FlagRunAddr - адрес и порт для запуска сервера.
	FlagRunAddr string
	// StoreInterval - интервал сохранения метрик в секундах.
	StoreInterval int
	// FileStoragePath - путь к файлу для хранения метрик.
	FileStoragePath string
	// DirStorageTmpPath - путь к временному каталогу для хранения файлов.
	DirStorageTmpPath string
	// Restore - флаг для восстановления данных из хранилища при запуске.
	Restore bool
	// DatabaseDsn - строка подключения к базе данных.
	DatabaseDsn string
	// MigrationPath - путь к файлам миграций базы данных.
	MigrationPath string
	// Key - ключ для аутентификации.
	Key string
}

// serverConfig - глобальная переменная, содержащая конфигурацию сервера.
var serverConfig *Server

// LoadServer загружает и инициализирует конфигурацию сервера.
// Функция обеспечивает однократную инициализацию конфигурации.
func LoadServer() *Server {
	initOnce.Do(func() {
		serverConfig = initServerConfig()
	})

	return serverConfig
}

// initServerConfig инициализирует конфигурацию сервера на основе переданных флагов
// командной строки и переменных окружения.
func initServerConfig() *Server {
	var cfg = Server{}

	//flag.StringVar(&cfg.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	//flag.StringVar(&cfg.DatabaseDsn, "d", "", "")
	//flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/metrics-db.json", "File to save metrics")
	flag.StringVar(&cfg.DirStorageTmpPath, "", "/tmp", "Dir storage tmp file")
	flag.IntVar(&cfg.StoreInterval, "i", 5, "Interval to save metrics")
	flag.BoolVar(&cfg.Restore, "r", true, "")
	flag.StringVar(&cfg.MigrationPath, "m", "file:./migrations", "")
	flag.StringVar(&cfg.Key, "k", "", "key")

	flag.StringVar(&cfg.FlagRunAddr, "a", "0.0.0.0:8080", "address and port to run server")
	flag.StringVar(&cfg.DatabaseDsn, "d", fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", "postgres", "postgres", "localhost", 5432, "metrics"), "")
	flag.StringVar(&cfg.FileStoragePath, "f", "./metrics-db.json", "File to save metrics")

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

	if envDatabaseDsn := os.Getenv("DATABASE_DSN"); envDatabaseDsn != "" {
		cfg.DatabaseDsn = envDatabaseDsn
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		cfg.Key = envKey
	}

	return &cfg
}
