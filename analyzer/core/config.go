package core

import "github.com/nrf24l01/go-web-utils/config"

type AnalyzerConfig struct {
	RabbitMQConfig *config.RabbitMQConfig
	PGConfig       *config.PGConfig
	AppConfig	   *AppConfig
	RedisConfig    *config.RedisConfig
}


func BuildConfigFromEnv() *AnalyzerConfig {
	cfg := &AnalyzerConfig{}
	cfg.RabbitMQConfig = config.LoadRabbitMQConfigFromEnv()
	cfg.PGConfig = config.LoadPGConfigFromEnv()
	cfg.AppConfig = LoadAppConfigFromEnv()
	cfg.RedisConfig = config.LoadRedisConfigFromEnv()
	return cfg
}
