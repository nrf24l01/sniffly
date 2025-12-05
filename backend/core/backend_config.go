package core

import (
	"log"

	"github.com/caarlos0/env/v11"
)

type BackendConfig struct {
	CacheEnabled       bool   `env:"CACHE_ENABLED" default:"true"`
	CacheTTL           uint   `env:"CACHE_TTL" default:"86400"`
	CacheDayAggPrefix  string `env:"CACHE_DAY_AGG_PREFIX" default:"sniffly_day_agg_"`
}

func LoadBackendConfigFromEnv() *BackendConfig {
	config := &BackendConfig{}
	if err := env.Parse(config); err != nil {
		log.Fatalf("Failed to parse environment variables: %v", err)
	}
	return config
}