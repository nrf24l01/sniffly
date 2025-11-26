package batcher

import (
	"encoding/json"
	"log"

	"github.com/nrf24l01/sniffly/analyzer/geoip"
	"github.com/nrf24l01/sniffly/capturer/snifpacket"
)

func buildDeviceTraffic(batch Batch, device_id uint64) (DeviceTraffic, error) {
	var dt DeviceTraffic
	for _, b := range batch.Packets {
		dt.Requests += 1
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
	dt.Domain = result
	dt.Bucket = batch.From
	dt.Requests = uint64(len(batch.Packets))
	return dt, nil
}

func (b *Batcher) buildDeviceCountryAndCompany(batch Batch, device_id uint64) (DeviceCountry, error) {
	var dc DeviceCountry
	all_ips := make([]string, 0)

	for _, b := range batch.Packets {
		all_ips = append(all_ips, b.DstIP)
	}

	dc.Requests = uint64(len(batch.Packets))
	dc.DeviceID = device_id
	dc.Bucket = batch.From
	
	for _, ip := range all_ips {
		county, company, err := geoip.CityCompanyFromIP(ip, b.RDB, b.CFG.AppConfig)
		if err != nil {
			log.Printf("Error looking up geoip info for IP %s: %v", ip, err)
			continue
		}
		found := false
		for _, c := range dc.Country {
			if c == county {
				found = true
				break
			}
		}
		if !found {
			dc.Country = append(dc.Country, county)
		}

		found = false
		for _, c := range dc.Company {
			if c == company {
				found = true
				break
			}
		}
		if !found {
			dc.Company = append(dc.Company, company)
		}
	}

	return dc, nil
}

func buildDeviceProto(batch Batch, device_id uint64) (DeviceProto, error) {
	protos := make(map[string]uint64)
	for _, b := range batch.Packets {
		protos[b.Protocol] += 1
	}
	result, err := json.Marshal(protos)
	if err != nil {
		return DeviceProto{}, err
	}

	var dt DeviceProto
	dt.DeviceID = device_id
	dt.Proto = result
	dt.Bucket = batch.From
	dt.Requests = uint64(len(batch.Packets))
	return dt, nil
}

func (b *Batcher) getDevicePackets(batches []Batch, device_id uint64) (CHBatch, error) {
	var result CHBatch

	for _, batch := range batches {
		// Build device traffic
		traffic, err := buildDeviceTraffic(batch, device_id)
		if err != nil {
			return CHBatch{}, err
		}
		result.DeviceTraffics = append(result.DeviceTraffics, traffic)

		// Build device domain
		domain, err := buildDeviceDomain(batch, device_id)
		if err != nil {
			return CHBatch{}, err
		}
		result.DeviceDomains = append(result.DeviceDomains, domain)

		// Build device country and company
		country, err := b.buildDeviceCountryAndCompany(batch, device_id)
		if err != nil {
			return CHBatch{}, err
		}
		result.DeviceCountries = append(result.DeviceCountries, country)

		// Build device proto
		proto, err := buildDeviceProto(batch, device_id)
		if err != nil {
			return CHBatch{}, err
		}
		result.DeviceProtos = append(result.DeviceProtos, proto)
	}
	return result, nil
}