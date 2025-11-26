package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/joho/godotenv"
	"github.com/nrf24l01/go-web-utils/pg_kit"
	"github.com/nrf24l01/go-web-utils/rabbitMQ"
	redisutil "github.com/nrf24l01/go-web-utils/redis"
	"github.com/nrf24l01/sniffly/analyzer/batcher"
	"github.com/nrf24l01/sniffly/analyzer/core"
	"github.com/nrf24l01/sniffly/analyzer/postgres"
)

func main() {
	if os.Getenv("PRODUCTION_ENV") != "true" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("failed to load .env: %v", err)
		}
	}

	cfg := core.BuildConfigFromEnv()
	ctx := context.Background()
	
	// Init Postgres
	pg_db, err := pg_kit.RegisterPostgres(cfg.PGConfig, &postgres.DeviceInfo{}, &postgres.DeviceCountry5s{}, &postgres.DeviceDomain5s{}, &postgres.DeviceProto5s{}, &postgres.DeviceTraffic5s{})
	if err != nil {
		log.Fatalf("failed to connect to Postgres: %v", err)
	}
	// Init timescaleDB
	err = postgres.InitTimescaleDB(pg_db)
	if err != nil {
		log.Fatalf("failed to initialize TimescaleDB: %v", err)
	}


	// Init RabbitMQ
	rmq, err := rabbitMQ.RegisterRabbitMQ(cfg.RabbitMQConfig)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}

	// Init Redis
	rdb := redisutil.NewRedisClient(cfg.RedisConfig)

	// Init snowflake
	node, err := snowflake.NewNode(1)
	if err != nil {
		log.Fatalf("failed to initialize snowflake node: %v", err)
	}

	batcher := batcher.Batcher{
		RMQ:  rmq,
		PGDB: pg_db,
		CFG:  cfg,
		SnowflakeNode: node,
		RDB: rdb,
	}

	log.Printf("Starting analyzer batcher...")
	batch, err := batcher.RecordBatch(time.Duration(2) * time.Second)
	if err != nil {
		log.Fatalf("failed to record batch: %v", err)
	}
	err = batcher.Process(ctx, batch)
	if err != nil {
		log.Fatalf("failed to process batch: %v", err)
	}
	log.Printf("Finished batch analyze")
}