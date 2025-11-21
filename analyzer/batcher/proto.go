package batcher

import "time"

type DeviceTraffic struct {
	DeviceID         uint64
	Bucket           time.Time
	UpBytes          uint64
	ReqCount         uint64
}

type DeviceDomain struct {
	DeviceID         uint64
	Bucket           time.Time
	Domain           string
	Requests         uint64
}

type DeviceCountry struct {
	DeviceID         uint64
	Bucket           time.Time
	Country          []string
	Company          []string
	Requests         uint64
}

type DeviceProto struct {
	DeviceID         uint64
	Bucket           time.Time
	Proto            string
	Requests         uint64
}

type CHBatch struct {
	DeviceTraffics   []DeviceTraffic
	DeviceDomains    []DeviceDomain
	DeviceCountries  []DeviceCountry
	DeviceProtos     []DeviceProto
}