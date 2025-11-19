package clickhouse

import (
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ClickHouse struct {
	CH driver.Conn
}