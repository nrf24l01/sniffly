package core

import (
	"log"

	"github.com/caarlos0/env/v11"
)

type AppConfig struct {
	CapturePacketsTopic  string `env:"CAPTURE_PACKETS_TOPIC" envDefault:"sniffed"`
	GeoIPCacheTTL        int    `env:"GEOIP_CACHE_TTL" envDefault:"86400"`
	GeoIPCacheKeyPrefix  string `env:"GEOIP_CACHE_KEY_PREFIX" envDefault:"geoip-cache:"`
}

func LoadAppConfigFromEnv() *AppConfig {
	config := &AppConfig{}
	if err := env.Parse(config); err != nil {
		log.Fatalf("Failed to parse environment variables: %v", err)
	}
	return config
}