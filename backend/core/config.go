package core

import (
	utilsConfig "github.com/nrf24l01/go-web-utils/config"
)

type Config struct {
	WebAppConfig          *utilsConfig.WebAppConfig
	PGConfig              *utilsConfig.PGConfig
	RedisConfig           *utilsConfig.RedisConfig
	Argon2idConfig        *utilsConfig.Argon2idConfig
	JWTConfig             *utilsConfig.JWTConfig
}

func BuildConfigFromEnv() (*Config, error) {
	config := &Config{
		WebAppConfig:         utilsConfig.LoadWebAppConfigFromEnv(),
		PGConfig:             utilsConfig.LoadPGConfigFromEnv(),
		RedisConfig:          utilsConfig.LoadRedisConfigFromEnv(),
		Argon2idConfig:       utilsConfig.LoadArgon2idConfigFromEnv(),
		JWTConfig:            utilsConfig.LoadJWTConfigFromEnv(),
	}

	return config, nil
}