package postgres

import "gorm.io/gorm"

func InitTimescaleDB(db *gorm.DB) error {
	tx := db.Exec(`
        CREATE EXTENSION IF NOT EXISTS timescaledb;

        -- Ensure hypertables exist
        SELECT create_hypertable('devices_traffics_5s', 'bucket', if_not_exists => TRUE);
        SELECT create_hypertable('devices_domains_5s', 'bucket', if_not_exists => TRUE);
        SELECT create_hypertable('devices_countries_5s', 'bucket', if_not_exists => TRUE);
        SELECT create_hypertable('devices_protos_5s', 'bucket', if_not_exists => TRUE);

        -- Ensure unique indexes/constraints that match ON CONFLICT targets exist.
        -- ON CONFLICT (device_id, bucket) is used for traffics and countries.
        CREATE UNIQUE INDEX IF NOT EXISTS idx_devices_traffics_bucket_device ON devices_traffics_5s (device_id, bucket);
        CREATE UNIQUE INDEX IF NOT EXISTS idx_devices_countries_bucket_device ON devices_countries_5s (device_id, bucket);

        -- For domains and protos we need uniqueness including domain/proto column as used in ON CONFLICT.
        CREATE UNIQUE INDEX IF NOT EXISTS idx_devices_domains_bucket_device_domain ON devices_domains_5s (device_id, bucket);
        CREATE UNIQUE INDEX IF NOT EXISTS idx_devices_protos_bucket_device_proto ON devices_protos_5s (device_id, bucket);
    `)
	return tx.Error
}