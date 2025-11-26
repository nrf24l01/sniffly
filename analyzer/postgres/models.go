package postgres

import (
	"time"

	"github.com/lib/pq"
	"github.com/nrf24l01/go-web-utils/pg_kit"
)

type DeviceInfo struct {
	pg_kit.BaseModel

	DeviceID uint64 `gorm:"column:device_id;primaryKey;not null;index" json:"device_id"`
	MAC      string `gorm:"column:mac;size:17" json:"mac"`
	IP       string `gorm:"column:ip" json:"ip"`
	Label    string `gorm:"column:label" json:"label"`
	Hostname string `gorm:"column:hostname" json:"hostname"`
}

func (DeviceInfo) TableName() string {
	return "device_info"
}

type DeviceTraffic5s struct {
	pg_kit.BaseModel

	Bucket   time.Time `gorm:"column:bucket;not null;index:idx_bucket_device" json:"bucket"`
	DeviceID uint64    `gorm:"column:device_id;not null;index:idx_bucket_device" json:"device_id"`
	UpBytes  uint64    `gorm:"column:up_bytes" json:"up_bytes"`
	ReqCount uint64    `gorm:"column:req_count" json:"req_count"`

	Device DeviceInfo `gorm:"foreignKey:DeviceID;references:DeviceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (DeviceTraffic5s) TableName() string {
	return "devices_traffics_5s"
}

type DeviceDomain5s struct {
	pg_kit.BaseModel

	Bucket   time.Time `gorm:"column:bucket;not null;index:idx_bucket_device" json:"bucket"`
	DeviceID uint64    `gorm:"column:device_id;not null;index:idx_bucket_device" json:"device_id"`
	Domain   string    `gorm:"column:domain" json:"domain"`
	Requests uint64    `gorm:"column:requests" json:"requests"`

	Device DeviceInfo `gorm:"foreignKey:DeviceID;references:DeviceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (DeviceDomain5s) TableName() string {
	return "devices_domains_5s"
}

type DeviceCountry5s struct {
	pg_kit.BaseModel

	Bucket    time.Time       `gorm:"column:bucket;not null;index:idx_bucket_device" json:"bucket"`
	DeviceID  uint64          `gorm:"column:device_id;not null;index:idx_bucket_device" json:"device_id"`
	Companies pq.StringArray  `gorm:"column:companies;type:text[]" json:"companies"`
	Countries pq.StringArray  `gorm:"column:countries;type:text[]" json:"countries"`
	Requests  uint64          `gorm:"column:requests" json:"requests"`

	Device DeviceInfo `gorm:"foreignKey:DeviceID;references:DeviceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (DeviceCountry5s) TableName() string {
	return "devices_countries_5s"
}

type DeviceProto5s struct {
	pg_kit.BaseModel

	Bucket   time.Time `gorm:"column:bucket;not null;index:idx_bucket_device" json:"bucket"`
	DeviceID uint64    `gorm:"column:device_id;not null;index:idx_bucket_device" json:"device_id"`
	Proto    string    `gorm:"column:proto" json:"proto"`
	Requests uint64    `gorm:"column:requests" json:"requests"`

	Device DeviceInfo `gorm:"foreignKey:DeviceID;references:DeviceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (DeviceProto5s) TableName() string {
	return "devices_protos_5s"
}