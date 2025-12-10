package postgres

import (
	"time"

	"github.com/nrf24l01/go-web-utils/pg_kit"
)

type DayCacheVersion struct {
	pg_kit.BaseModel

	Day     time.Time `gorm:"type:date;not null;uniqueIndex:idx_day"`
	Version int    `gorm:"not null;default:0"`
}

func (DayCacheVersion) TableName() string {
	return "day_cache_versions"
}