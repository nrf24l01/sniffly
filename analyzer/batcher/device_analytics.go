package batcher

import (
	"encoding/json"
	"time"

	"github.com/nrf24l01/sniffly/capturer/snifpacket"
)

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
	Country          string
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

func buildDeviceTraffic(batch Batch, device_id uint64) (DeviceTraffic, error) {
	var dt DeviceTraffic
	for _, b := range batch.Packets {
		dt.ReqCount += 1
		dt.UpBytes += uint64(b.Size)
	}
	dt.DeviceID = device_id
	dt.Bucket = batch.From
	return dt, nil
}

func buildDeviceDomain(batch Batch, device_id uint64) (DeviceDomain, error) {
	domains := make(map[string]uint64)
	for _, b := range batch.Packets {
		if b.Details.Type == snifpacket.SnifPacketTypeHTTP && b.Details.HTTP != nil {
			domains[b.Details.HTTP.Host] += 1
		} else if b.Details.Type == snifpacket.SnifPacketTypeTLS && b.Details.TLS != nil {
			domains[b.Details.TLS.Sni] += 1
		}
	}
	result, err := json.Marshal(domains)
	if err != nil {
		return DeviceDomain{}, err
	}

	var dt DeviceDomain
	dt.DeviceID = device_id
	dt.Domain = string(result)
	dt.Bucket = batch.From
	dt.Requests = uint64(len(batch.Packets))
	return dt, nil
}

func buildDeviceCountry(batch Batch, device_id uint64) (DeviceCountry, error) {
	all_ips := make([]string, 0)

	for _, b := range batch.Packets {
		all_ips = append(all_ips, b.DstIP)
	}

	

	return dc, nil
}

func (b *Batcher) getDevicePackets(batches []Batch, device_id uint64) (CHBatch, error) {
	var result CHBatch

	for _, batch := range batches {
		traffic, err := buildDeviceTraffic(batch, device_id)
		if err != nil {
			return CHBatch{}, err
		}
		result.DeviceTraffics = append(result.DeviceTraffics, traffic)

		domain, err := buildDeviceDomain(batch, device_id)
		if err != nil {
			return CHBatch{}, err
		}
		result.DeviceDomains = append(result.DeviceDomains, domain)
	}
	return result, nil
}