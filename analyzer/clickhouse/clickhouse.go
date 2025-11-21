package clickhouse

import (
	ch "github.com/ClickHouse/clickhouse-go/v2"
)

type ClickHouse struct {
	CH ch.Conn
}