package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/nrf24l01/go-web-utils/pg_kit"
)

type DeviceInfo struct {
	pg_kit.BaseModel

	MAC       string    `gorm:"unique;size:17"`
    IP        string    `gorm:"default:''"`
    Label     string    `gorm:"default:'interface'"`
    Hostname  string    `gorm:"default:''"`
}

func (DeviceInfo) TableName() string {
	return "device_info"
}

type DeviceTraffic5s struct {
	pg_kit.BaseModel

	Bucket   time.Time `gorm:"not null;primaryKey;uniqueIndex:idx_bucket_device"`
    DeviceID uuid.UUID `gorm:"type:uuid;primaryKey;not null;uniqueIndex:idx_bucket_device"`
    UpBytes  uint64    `gorm:"default:0"`
    ReqCount uint64    `gorm:"default:0"`

    Device DeviceInfo `gorm:"foreignKey:DeviceID;references:ID;constraint:OnDelete:CASCADE"`
}

func (DeviceTraffic5s) TableName() string {
	return "devices_traffics_5s"
}

type DeviceDomain5s struct {
	pg_kit.BaseModel

	Bucket   time.Time `gorm:"not null;primaryKey;uniqueIndex:idx_bucket_device"`
	DeviceID uuid.UUID `gorm:"type:uuid;primaryKey;not null;uniqueIndex:idx_bucket_device"`
	Domain   string    `gorm:"type:jsonb;default:'{}'"`
	Requests uint64    `gorm:"default:0"`

	Device DeviceInfo `gorm:"foreignKey:DeviceID;references:ID;constraint:OnDelete:CASCADE"`
}

func (DeviceDomain5s) TableName() string {
	return "devices_domains_5s"
}

type DeviceCountry5s struct {
	pg_kit.BaseModel

	Bucket    time.Time `gorm:"not null;primaryKey;uniqueIndex:idx_bucket_device"`
	DeviceID  uuid.UUID `gorm:"type:uuid;primaryKey;not null;uniqueIndex:idx_bucket_device"`
	Companies string    `gorm:"type:jsonb;default:'{}'"`
	Countries string    `gorm:"type:jsonb;default:'{}'"`
	Requests  uint64    `gorm:"default:0"`

	Device DeviceInfo `gorm:"foreignKey:DeviceID;references:ID;constraint:OnDelete:CASCADE"`
}

func (DeviceCountry5s) TableName() string {
	return "devices_countries_5s"
}

type DeviceProto5s struct {
	pg_kit.BaseModel

	Bucket   time.Time `gorm:"not null;primaryKey;uniqueIndex:idx_bucket_device"`
	DeviceID uuid.UUID `gorm:"type:uuid;primaryKey;not null;uniqueIndex:idx_bucket_device"`
	Proto    string    `gorm:"type:jsonb;default:'{}'"`
	Requests uint64    `gorm:"default:0"`

	Device DeviceInfo `gorm:"foreignKey:DeviceID;references:ID;constraint:OnDelete:CASCADE"`
}

func (DeviceProto5s) TableName() string {
	return "devices_protos_5s"
}