package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/nrf24l01/go-web-utils/pg_kit"
	"github.com/nrf24l01/go-web-utils/rabbitMQ"
	"github.com/nrf24l01/sniffly/capture_receiver/core"
	"github.com/nrf24l01/sniffly/capture_receiver/handler"
	"github.com/nrf24l01/sniffly/capture_receiver/postgres"
)

func main() {
	if os.Getenv("PRODUCTION_ENV") != "true" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("failed to load .env: %v", err)
		}
	}

	cfg := core.BuildConfigFromEnv()

	db, err := pg_kit.RegisterPostgres(cfg.PGConfig, &postgres.Capturer{})
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	
	rmq, err := rabbitMQ.RegisterRabbitMQ(cfg.RabbitMQConfig)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	h := handler.PacketGatewayServer{
		Config: cfg,
		DB:     db,
		RMQ:    rmq,
	}

	StartGRPCServer(cfg, &h)
}