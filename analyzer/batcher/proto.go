package batcher

import (
	"time"

	"github.com/google/uuid"
)

type BaseDeviceStat struct {
	DeviceID         uuid.UUID
	Bucket           time.Time
	Requests         uint64
}

type DeviceTraffic struct {
	BaseDeviceStat
	UpBytes          uint64
}

type DeviceDomain struct {
	BaseDeviceStat
	Domain           []byte
}

type DeviceCountry struct {
	BaseDeviceStat
	Country          []byte
	Company          []byte
}

type DeviceProto struct {
	BaseDeviceStat
	Proto            []byte
}

type DeviceStatLike interface {
	GetBucket() time.Time
	GetDeviceID() uuid.UUID
}

// Provide accessors on the embedded BaseDeviceStat so concrete types satisfy
// the DeviceStatLike interface via promotion.
func (b BaseDeviceStat) GetBucket() time.Time {
	return b.Bucket
}

func (b BaseDeviceStat) GetDeviceID() uuid.UUID {
	return b.DeviceID
}

// type AnyStat interface {
// 	isAnyStat()
// }

// func (DeviceCountry) isAnyStat()  {}
// func (DeviceDomain) isAnyStat()   {}
// func (DeviceProto) isAnyStat()    {}
// func (DeviceTraffic) isAnyStat()  {}

type CHBatch struct {
	DeviceTraffics   []DeviceTraffic
	DeviceDomains    []DeviceDomain
	DeviceCountries  []DeviceCountry
	DeviceProtos     []DeviceProto
}