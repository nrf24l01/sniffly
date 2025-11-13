package core

import (
	"log"

	"github.com/caarlos0/env/v11"
)

type CaptureConfig struct {
	ReflectionEnabled bool   `env:"CAPTURE_REFLECTION_ENABLED" envDefault:"false"`
	AppHost 		  string `env:"CAPTURE_APP_HOST" envDefault:":50051"`
	PacketsTopic	  string `env:"CAPTURE_PACKETS_TOPIC" envDefault:"sniffed_packets"`
}


func LoadCaptureConfigFromEnv() *CaptureConfig {
	config := &CaptureConfig{}
	if err := env.Parse(config); err != nil {
		log.Fatalf("Failed to parse environment variables: %v", err)
	}
	return config
}