package core

import (
	"log"

	"github.com/caarlos0/env/v11"
)

type CHConfig struct {
	CHHost     string `env:"CH_HOST" envDefault:"localhost"`
	CHPort     string `env:"CH_PORT" envDefault:"9000"`
	CHUser     string `env:"CH_USER" envDefault:"postgres"`
	CHPassword string `env:"CH_PASSWORD" envDefault:"password"`
	CHDatabase string `env:"CH_DATABASE" envDefault:"postgres"`
}

func LoadCHConfigFromEnv() *CHConfig {
	config := &CHConfig{}
	if err := env.Parse(config); err != nil {
		log.Fatalf("Failed to parse environment variables: %v", err)
	}
	return config
}