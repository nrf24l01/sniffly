package batcher

import "time"

type BaseDeviceStat struct {
	DeviceID         uint64
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
	Country          []string
	Company          []string
}

type DeviceProto struct {
	BaseDeviceStat
	Proto            []byte
}

type DeviceStatLike interface {
    DeviceTraffic | DeviceDomain | DeviceCountry | DeviceProto
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