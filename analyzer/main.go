package main

import (
	"context"
	"log"
	"os"
	"strings"
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

func isAMQPClosed(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "channel/connection is not open") || strings.Contains(s, "connection is closed") || strings.Contains(s, "channel is closed")
}

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
	pg_db, err := pg_kit.RegisterPostgres(cfg.PGConfig,
		&postgres.DeviceInfo{},
		&postgres.DeviceCountry5s{}, &postgres.DeviceDomain5s{}, &postgres.DeviceProto5s{}, &postgres.DeviceTraffic5s{},
		&postgres.DayCacheVersion{},
	)
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

	// Limit prefetch to avoid large numbers of Unacked messages per consumer
	if err := rmq.Channel.Qos(1, 0, false); err != nil {
		log.Printf("failed to set QoS on RabbitMQ channel: %v", err)
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

	for {
		log.Printf("Starting to load batch of records")
		batch, err := batcher.LoadAllRecords()
		log.Printf("Loaded batch with %d records", len(batch.Packets))
		if err != nil {
			if isAMQPClosed(err) {
				log.Printf("RabbitMQ channel closed: %v — reconnecting...", err)
				for {
					rmq, err = rabbitMQ.RegisterRabbitMQ(cfg.RabbitMQConfig)
					if err == nil {
						break
					}
					log.Printf("reconnect failed: %v; retrying in 5s", err)
					time.Sleep(5 * time.Second)
				}
				batcher.RMQ = rmq
				if err := rmq.Channel.Qos(1, 0, false); err != nil {
					log.Printf("failed to set QoS on RabbitMQ channel: %v", err)
				}
				continue
			}
			log.Printf("failed to record batch: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		err = batcher.Process(ctx, batch)
		if err != nil {
			if isAMQPClosed(err) {
				log.Printf("RabbitMQ channel closed during processing: %v — reconnecting...", err)
				for {
					rmq, err = rabbitMQ.RegisterRabbitMQ(cfg.RabbitMQConfig)
					if err == nil {
						break
					}
					log.Printf("reconnect failed: %v; retrying in 5s", err)
					time.Sleep(5 * time.Second)
				}
				batcher.RMQ = rmq
				if err := rmq.Channel.Qos(1, 0, false); err != nil {
					log.Printf("failed to set QoS on RabbitMQ channel: %v", err)
				}
				continue
			}
			log.Printf("failed to process batch: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		time.Sleep(time.Second * 10)
	}
}