package clickhouse

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/nrf24l01/sniffly/analyzer/core"
)

func InitClickHouseDatabase(ctx context.Context, cfg *core.CHConfig) (ClickHouse, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.CHHost + ":" + cfg.CHPort},
		Auth: clickhouse.Auth{
			Database: cfg.CHDatabase,
			Username: cfg.CHUser,
			Password: cfg.CHPassword,
		},
	})
	if err != nil {
		return ClickHouse{}, err
	}

	// Ping
	if err := conn.Ping(ctx); err != nil {
		return ClickHouse{}, err
	}

	return ClickHouse{CH: conn}, nil
}