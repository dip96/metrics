package config

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
	"sync"
)

// Agent представляет конфигурацию агента.
type Agent struct {
	// FlagRunAddr - адрес и порт для запуска сервера.
	FlagRunAddr string `json:"address"`
	// FlagReportInterval - интервал отчетности в секундах.
	FlagReportInterval int `json:"report_interval"`
	// FlagRuntime - время работы приложения в секундах.
	FlagRuntime int `json:"runtime"`
	// Key - ключ для аутентификации.
	Key string `json:"key"`
	// RateLimit - ограничение скорости в запросах в секунду.
	RateLimit int `json:"rate_limit"`
	// CryptoKey - путь до файла с публичным ключом
	CryptoKey string `json:"crypto_key"`
	// Config - путь до файла конфигурации
	Config string
}

// agentConfig - глобальная переменная, содержащая конфигурацию агента.
var agentConfig *Agent

// initOnce - объект для обеспечения однократной инициализации конфигурации.
var initOnce sync.Once

// LoadAgent загружает и инициализирует конфигурацию агента.
// Функция обеспечивает однократную инициализацию конфигурации.
func LoadAgent() (*Agent, error) {
	var err error
	initOnce.Do(func() {
		agentConfig, err = initAgentConfig()
	})
	if err != nil {
		return nil, err
	}
	return agentConfig, nil
}

// initAgentConfig инициализирует конфигурацию агента на основе переданных флагов
// командной строки и переменных окружения.
func initAgentConfig() (*Agent, error) {
	var cfg = Agent{}

	agentFlags := flag.NewFlagSet("agent", flag.ExitOnError)
	agentFlags.StringVar(&cfg.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	agentFlags.StringVar(&cfg.CryptoKey, "crypto-key", "", "public key")
	agentFlags.IntVar(&cfg.FlagReportInterval, "r", 10, "address and port to run server")
	agentFlags.IntVar(&cfg.FlagRuntime, "p", 2, "address and port to run server")
	agentFlags.StringVar(&cfg.Key, "k", "", "key")
	agentFlags.IntVar(&cfg.RateLimit, "l", 10, "Rate limit")
	agentFlags.StringVar(&cfg.Config, "c", "/home/dip96/go_project/src/metrics/config_agent.json", "Config path")

	if cfg.Config != "" {
		err := readConfigFileAgent(cfg.Config, cfg)
		if err != nil {
			return nil, err
		}
	}

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.FlagRunAddr = envRunAddr
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		cfg.FlagReportInterval, _ = strconv.Atoi(envReportInterval)
	}

	if envRuntime := os.Getenv("RUNTIME"); envRuntime != "" {
		cfg.FlagRuntime, _ = strconv.Atoi(envRuntime)
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		cfg.Key = envKey
	}

	if envRateLimit := os.Getenv("RATE_LIMIT"); envRateLimit != "" {
		cfg.RateLimit, _ = strconv.Atoi(envRateLimit)
	}

	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		cfg.CryptoKey = envCryptoKey
	}

	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		cfg.Config = envConfig
	}

	return &cfg, nil
}

func readConfigFileAgent(path string, cfg Agent) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return err
	}

	return nil
}
