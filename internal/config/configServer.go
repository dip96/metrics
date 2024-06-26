package config

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
	"sync"
)

// Server представляет конфигурацию сервера.
type Server struct {
	// FlagRunAddr - адрес и порт для запуска сервера.
	FlagRunAddr string `json:"address"`
	// StoreInterval - интервал сохранения метрик в секундах.
	StoreInterval int `json:"store_interval"`
	// FileStoragePath - путь к файлу для хранения метрик.
	FileStoragePath string `json:"store_file"`
	// DirStorageTmpPath - путь к временному каталогу для хранения файлов.
	DirStorageTmpPath string `json:"dir_storage_tmp_path"`
	// Restore - флаг для восстановления данных из хранилища при запуске.
	Restore bool `json:"restore"`
	// DatabaseDsn - строка подключения к базе данных.
	DatabaseDsn string `json:"database_dsn"`
	// MigrationPath - путь к файлам миграций базы данных.
	MigrationPath string `json:"migration_path"`
	// Key - ключ для аутентификации.
	Key string `json:"key"`
	// CryptoKey - путь до файла с приватным ключом
	CryptoKey string `json:"crypto_key"`
	// Config - путь до файла конфигурации
	Config string
}

// serverConfig - глобальная переменная, содержащая конфигурацию сервера.
var serverConfig *Server
var initOnceServer sync.Once

// LoadServer загружает и инициализирует конфигурацию сервера.
// Функция обеспечивает однократную инициализацию конфигурации.
func LoadServer() (*Server, error) {
	var err error
	initOnceServer.Do(func() {
		serverConfig, err = initServerConfig()
	})

	if err != nil {
		return nil, err
	}

	return serverConfig, nil
}

// initServerConfig инициализирует конфигурацию сервера на основе переданных флагов
// командной строки и переменных окружения.
func initServerConfig() (*Server, error) {
	var cfg = Server{}
	serverFlags := flag.NewFlagSet("server", flag.ExitOnError)

	serverFlags.StringVar(&cfg.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	serverFlags.StringVar(&cfg.DatabaseDsn, "d", "", "")
	serverFlags.StringVar(&cfg.CryptoKey, "crypto-key", "/tmp/keys", "private key")
	serverFlags.StringVar(&cfg.FileStoragePath, "f", "/tmp/metrics-db.json", "File to save metrics")
	serverFlags.StringVar(&cfg.DirStorageTmpPath, "", "/tmp", "Dir storage tmp file")
	serverFlags.IntVar(&cfg.StoreInterval, "i", 5, "Interval to save metrics")
	serverFlags.BoolVar(&cfg.Restore, "r", true, "")
	serverFlags.StringVar(&cfg.MigrationPath, "m", "file:./migrations", "")
	serverFlags.StringVar(&cfg.Key, "k", "", "key")
	serverFlags.StringVar(&cfg.Config, "c", "/home/dip96/go_project/src/metrics/config_server.json", "Config path")

	if cfg.Config != "" {
		err := readConfigFileServer(cfg.Config, &cfg)

		if err != nil {
			return nil, err
		}
	}

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

	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		cfg.CryptoKey = envCryptoKey
	}

	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		cfg.Config = envConfig
	}

	return &cfg, nil
}

func readConfigFileServer(path string, cfg *Server) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(cfg)
	if err != nil {
		return err
	}

	return nil
}
