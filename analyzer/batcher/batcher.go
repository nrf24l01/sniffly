package batcher

import (
	"github.com/bwmarrin/snowflake"
	"github.com/nrf24l01/go-web-utils/rabbitMQ"
	redisutil "github.com/nrf24l01/go-web-utils/redis"
	"github.com/nrf24l01/sniffly/analyzer/core"
	"gorm.io/gorm"
)

type Batcher struct {
	RMQ  *rabbitMQ.RabbitMQ
	PGDB *gorm.DB
	RDB  *redisutil.RedisClient
	CFG  *core.AnalyzerConfig
	SnowflakeNode *snowflake.Node
}