package core

import "github.com/nrf24l01/go-web-utils/config"

type AnalyzerConfig struct {
	RabbitMQConfig *config.RabbitMQConfig
	CHConfig       *CHConfig
	AppConfig	   *AppConfig
	RedisConfig    *config.RedisConfig
}


func BuildConfigFromEnv() *AnalyzerConfig {
	cfg := &AnalyzerConfig{}
	cfg.RabbitMQConfig = config.LoadRabbitMQConfigFromEnv()
	cfg.CHConfig = LoadCHConfigFromEnv()
	cfg.AppConfig = LoadAppConfigFromEnv()
	return cfg
}
