package core

import "github.com/nrf24l01/go-web-utils/config"

type AppConfig struct {
	PGConfig       *config.PGConfig
	RabbitMQConfig *config.RabbitMQConfig
	CaptureConfig  *CaptureConfig
}

func BuildConfigFromEnv() *AppConfig {
	cfg := &AppConfig{}
	cfg.PGConfig = config.LoadPGConfigFromEnv()
	cfg.RabbitMQConfig = config.LoadRabbitMQConfigFromEnv()
	cfg.CaptureConfig = LoadCaptureConfigFromEnv()
	return cfg
}
