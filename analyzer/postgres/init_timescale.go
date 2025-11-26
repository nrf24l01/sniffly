package postgres

import "gorm.io/gorm"

func InitTimescaleDB(db *gorm.DB) error {
	tx := db.Exec(`
		CREATE EXTENSION IF NOT EXISTS timescaledb;
        SELECT create_hypertable('devices_traffics_5s', 'bucket', if_not_exists => TRUE);
        SELECT create_hypertable('devices_traffics_5s', 'bucket', if_not_exists => TRUE);
        SELECT create_hypertable('devices_domains_5s', 'bucket', if_not_exists => TRUE);
        SELECT create_hypertable('devices_countries_5s', 'bucket', if_not_exists => TRUE);
        SELECT create_hypertable('devices_protos_5s', 'bucket', if_not_exists => TRUE);
    `)
	return tx.Error
}