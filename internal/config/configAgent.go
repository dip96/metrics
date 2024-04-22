package config

import (
	"flag"
	"os"
	"strconv"
	"sync"
)

type Agent struct {
	FlagRunAddr        string
	FlagReportInterval int
	FlagRuntime        int
	Key                string
	RateLimit          int
}

var (
	agentConfig *Agent
	initOnce    sync.Once
)

func LoadAgent() *Agent {
	initOnce.Do(func() {
		agentConfig = initAgentConfig()
	})
	return agentConfig
}

func initAgentConfig() *Agent {
	var cfg = Agent{}

	flag.StringVar(&cfg.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&cfg.FlagReportInterval, "r", 10, "address and port to run server")
	flag.IntVar(&cfg.FlagRuntime, "p", 2, "address and port to run server")
	flag.StringVar(&cfg.Key, "k", "", "key")
	flag.IntVar(&cfg.RateLimit, "l", 10, "Rate limit")

	//flag.StringVar(&cfg.FlagRunAddr, "a", "0.0.0.0:8080", "address and port to run server")

	flag.Parse()

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

	return &cfg
}
