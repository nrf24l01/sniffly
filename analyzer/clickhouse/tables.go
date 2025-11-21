package clickhouse

import "context"

func (c *ClickHouse) CreateTables(ctx context.Context) error {
	createTablesCommands := []string{`
		CREATE TABLE IF NOT EXISTS device_info (
			device_id UInt64,
			mac FixedString(17),
			ip String,
			label String,
			hostname String
		)
		ENGINE = MergeTree
		ORDER BY device_id;
	`,
	`
		CREATE TABLE IF NOT EXISTS device_traffic_5s (
			device_id UInt64,
			bucket DateTime,
			up_bytes UInt64,
			req_count UInt64
		)
		ENGINE = MergeTree
		PARTITION BY toYYYYMMDD(bucket)
		ORDER BY (device_id, bucket)
	`,
	`
		CREATE TABLE IF NOT EXISTS device_domain_5s (
			device_id UInt32,
			bucket DateTime,
			domain String,
			requests UInt64
		)
		ENGINE = MergeTree
		PARTITION BY toYYYYMMDD(bucket)
		ORDER BY (device_id, bucket);
	`,
	`
		CREATE TABLE IF NOT EXISTS device_country_5s (
			device_id UInt32,
			bucket DateTime,
			companies Array(String),
			countries Array(String),
			requests UInt64
		)
		ENGINE = MergeTree
		PARTITION BY toYYYYMMDD(bucket)
		ORDER BY (device_id, bucket);
	`,
	`
		CREATE TABLE IF NOT EXISTS device_proto_5s (
			device_id UInt32,
			bucket DateTime,
			proto LowCardinality(String),
			requests UInt64
		)
		ENGINE = MergeTree
		PARTITION BY toYYYYMMDD(bucket)
		ORDER BY (device_id, bucket);
	`}
	
	for _, command := range createTablesCommands {
		if err := c.CH.Exec(ctx, command); err != nil {
			return err
		}
	}
	return nil
}