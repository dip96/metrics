package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"sync"
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
	// CryptoKey - путь до файла с приватным ключом
	CryptoKey string
}

// serverConfig - глобальная переменная, содержащая конфигурацию сервера.
var serverConfig *Server
var initOnceServer sync.Once

// LoadServer загружает и инициализирует конфигурацию сервера.
// Функция обеспечивает однократную инициализацию конфигурации.
func LoadServer() *Server {
	initOnceServer.Do(func() {
		serverConfig = initServerConfig()
	})

	return serverConfig
}

// initServerConfig инициализирует конфигурацию сервера на основе переданных флагов
// командной строки и переменных окружения.
func initServerConfig() *Server {
	var cfg = Server{}
	serverFlags := flag.NewFlagSet("server", flag.ExitOnError)

	//serverFlags.StringVar(&cfg.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	//serverFlags.StringVar(&cfg.DatabaseDsn, "d", "", "")
	//serverFlags.StringVar(&cfg.CryptoKey, "crypto-key", "/tmp/keys", "private key")
	//serverFlags.StringVar(&cfg.FileStoragePath, "f", "/tmp/metrics-db.json", "File to save metrics")
	serverFlags.StringVar(&cfg.DirStorageTmpPath, "", "/tmp", "Dir storage tmp file")
	serverFlags.IntVar(&cfg.StoreInterval, "i", 5, "Interval to save metrics")
	serverFlags.BoolVar(&cfg.Restore, "r", true, "")
	serverFlags.StringVar(&cfg.MigrationPath, "m", "file:./migrations", "")
	serverFlags.StringVar(&cfg.Key, "k", "", "key")

	serverFlags.StringVar(&cfg.FlagRunAddr, "a", "0.0.0.0:8080", "address and port to run server")
	serverFlags.StringVar(&cfg.DatabaseDsn, "d", fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", "postgres", "postgres", "localhost", 5432, "metrics"), "")
	serverFlags.StringVar(&cfg.CryptoKey, "crypto-key", "/home/dip96/go_project/src/metrics/keys/private.pem", "private key")
	serverFlags.StringVar(&cfg.FileStoragePath, "f", "./metrics-db.json", "File to save metrics")

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

	if envCryptoKey := os.Getenv("CryptoKey"); envCryptoKey != "" {
		cfg.CryptoKey = envCryptoKey
	}

	return &cfg
}
