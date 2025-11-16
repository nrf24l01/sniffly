package core

import "github.com/nrf24l01/go-web-utils/config"

type AnalyzerConfig struct {
	RabbitMQConfig *config.RabbitMQConfig
	CHConfig       *CHConfig
}


func BuildConfigFromEnv() *AnalyzerConfig {
	cfg := &AnalyzerConfig{}
	cfg.RabbitMQConfig = config.LoadRabbitMQConfigFromEnv()
	cfg.CHConfig = LoadCHConfigFromEnv()
	return cfg
}
