package config

import (
	"flag"
	"os"
	"strconv"
)

type Agent struct {
	FlagRunAddr        string
	FlagReportInterval int
	FlagRuntime        int
}

var AgentConfig *Agent

func LoadAgent() *Agent {
	if AgentConfig == nil {
		var cfg = Agent{}

		flag.StringVar(&cfg.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
		flag.IntVar(&cfg.FlagReportInterval, "r", 10, "address and port to run server")
		flag.IntVar(&cfg.FlagRuntime, "p", 2, "address and port to run server")

		//flag.StringVar(&conf.flagRunAddr, "a", "0.0.0.0:8080", "address and port to run server")

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

		AgentConfig = &cfg
	}

	return AgentConfig
}
