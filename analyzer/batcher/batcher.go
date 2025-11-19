package batcher

import (
	"github.com/bwmarrin/snowflake"
	"github.com/nrf24l01/go-web-utils/rabbitMQ"
	"github.com/nrf24l01/sniffly/analyzer/clickhouse"
	"github.com/nrf24l01/sniffly/analyzer/core"
)

type Batcher struct {
	RMQ  *rabbitMQ.RabbitMQ
	CHDB *clickhouse.ClickHouse
	CFG  *core.AnalyzerConfig
	SnowflakeNode *snowflake.Node
}