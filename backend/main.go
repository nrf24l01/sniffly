package main

import (
	random "github.com/nrf24l01/go-web-utils/misc/random"
	pgKit "github.com/nrf24l01/go-web-utils/pg_kit"
	"github.com/nrf24l01/go-web-utils/redis"
	"github.com/nrf24l01/sniffly/backend/core"
	"github.com/nrf24l01/sniffly/backend/launch"
	"github.com/nrf24l01/sniffly/backend/postgres"

	"log"
	"os"

	"github.com/joho/godotenv"
)
func main() {
	// Try to load .env file in non-production environment
	if os.Getenv("PRODUCTION_ENV") != "true" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("failed to load .env: %v", err)
		}
	}

	// Configuration initialization
	config, err := core.BuildConfigFromEnv()
	if err != nil {
		log.Fatalf("failed to build config: %v", err)
	}

	randomGenerator, err := random.NewRandomGenerator(random.RandomGeneratorConfig{})
	if err != nil {
		log.Fatalf("failed to create random generator: %v", err)
	}

	// Data sources initialization
	db, err := pgKit.RegisterPostgres(config.PGConfig, false, &postgres.User{})
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	rdb := redis.NewRedisClient(config.RedisConfig)

	launch.Dispatch(config, db, rdb, &randomGenerator, os.Args[1:])
}