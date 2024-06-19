package config

import (
	"flag"
	"os"
	"strconv"
	"sync"
)

// Agent представляет конфигурацию агента.
type Agent struct {
	// FlagRunAddr - адрес и порт для запуска сервера.
	FlagRunAddr string
	// FlagReportInterval - интервал отчетности в секундах.
	FlagReportInterval int
	// FlagRuntime - время работы приложения в секундах.
	FlagRuntime int
	// Key - ключ для аутентификации.
	Key string
	// RateLimit - ограничение скорости в запросах в секунду.
	RateLimit int
	// CryptoKey - путь до файла с публичным ключом
	CryptoKey string
}

// agentConfig - глобальная переменная, содержащая конфигурацию агента.
var agentConfig *Agent

// initOnce - объект для обеспечения однократной инициализации конфигурации.
var initOnce sync.Once

// LoadAgent загружает и инициализирует конфигурацию агента.
// Функция обеспечивает однократную инициализацию конфигурации.
func LoadAgent() *Agent {
	initOnce.Do(func() {
		agentConfig = initAgentConfig()
	})
	return agentConfig
}

// initAgentConfig инициализирует конфигурацию агента на основе переданных флагов
// командной строки и переменных окружения.
func initAgentConfig() *Agent {
	var cfg = Agent{}

	agentFlags := flag.NewFlagSet("agent", flag.ExitOnError)
	//agentFlags.StringVar(&cfg.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	//agentFlags.StringVar(&cfg.CryptoKey, "crypto-key", "", "public key")
	agentFlags.IntVar(&cfg.FlagReportInterval, "r", 10, "address and port to run server")
	agentFlags.IntVar(&cfg.FlagRuntime, "p", 2, "address and port to run server")
	agentFlags.StringVar(&cfg.Key, "k", "", "key")
	agentFlags.IntVar(&cfg.RateLimit, "l", 10, "Rate limit")

	agentFlags.StringVar(&cfg.FlagRunAddr, "a", "0.0.0.0:8080", "address and port to run server")
	agentFlags.StringVar(&cfg.CryptoKey, "crypto-key", "/home/dip96/go_project/src/metrics/keys/public.pem", "public key")

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

	if envCryptoKey := os.Getenv("CryptoKey"); envCryptoKey != "" {
		cfg.CryptoKey = envCryptoKey
	}

	return &cfg
}
