package core

import (
	"log"

	"github.com/caarlos0/env"
)

type Config struct {
	ServerAddress                 string `env:"SERVER_ADDRESS" envDefault:"localhost:50051"`
	ApiToken                      string `env:"API_TOKEN" envDefault:""`
	Interface                     string `env:"INTERFACE" envDefault:"eth0"`
	IptablesChain                 string `env:"IPTABLES_CHAIN" envDefault:"SNIFFLY_BLOCK"`
	BlockRulesSyncIntervalSeconds int    `env:"BLOCK_RULES_SYNC_INTERVAL_SECONDS" envDefault:"30"`
}

func LoadConfigFromEnv() *Config {
	config := &Config{}
	if err := env.Parse(config); err != nil {
		log.Fatalf("Failed to load config from env: %v", err)
	}
	return config
}
