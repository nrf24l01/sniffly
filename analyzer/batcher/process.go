package batcher

import (
	"context"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/nrf24l01/sniffly/analyzer/postgres"
	"github.com/nrf24l01/sniffly/capturer/snifpacket"
	"gorm.io/gorm"
)

func isIPv4Address(ip string) bool {
	parsed := net.ParseIP(ip)
	return parsed != nil && parsed.To4() != nil
}

func (b *Batcher) Process(ctx context.Context, batch Batch) error {
	// Grouping packets by device MAC
	per_device_mac := make(map[string][]snifpacket.SnifPacket)
	for _, packet := range batch.Packets {
		per_device_mac[packet.SrcMAC] = append(per_device_mac[packet.SrcMAC], packet)
	}

	// Retrieving or creating device IDs
	per_device_mac_device_id := make(map[string]uuid.UUID)
	for deviceMAC, packets := range per_device_mac {
		ipv4 := ""
		for _, packet := range packets {
			if isIPv4Address(packet.SrcIP) {
				ipv4 = packet.SrcIP
				break
			}
		}
		if ipv4 == "" {
			continue
		}

		var device postgres.DeviceInfo
		err := b.PGDB.Where("mac = ?", deviceMAC).First(&device).Error
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}

			device = postgres.DeviceInfo{
				MAC: deviceMAC,
				IP:  ipv4,
			}
			if err := b.PGDB.Create(&device).Error; err != nil {
				return err
			}
		} else if !isIPv4Address(device.IP) {
			if err := b.PGDB.Model(&device).Update("ip", ipv4).Error; err != nil {
				return err
			}
		}

		per_device_mac_device_id[deviceMAC] = device.ID
	}

	// Grouping packets by device ID
	per_device_id := make(map[uuid.UUID][]snifpacket.SnifPacket)
	for mac, packets := range per_device_mac {
		device_id := per_device_mac_device_id[mac]
		per_device_id[device_id] = packets
	}

	var bigBatch CHBatch

	// Processing per device ID
	for device_id, packets := range per_device_id {
		chBatch, err := b.processDevicBigBatch(ctx, device_id, packets)
		if err != nil {
			return err
		}
		bigBatch.DeviceTraffics = append(bigBatch.DeviceTraffics, chBatch.DeviceTraffics...)
		bigBatch.DeviceDomains = append(bigBatch.DeviceDomains, chBatch.DeviceDomains...)
		bigBatch.DeviceCountries = append(bigBatch.DeviceCountries, chBatch.DeviceCountries...)
		bigBatch.DeviceProtos = append(bigBatch.DeviceProtos, chBatch.DeviceProtos...)
	}

	return bigBatch.Insert(ctx, b)
}

func (b *Batcher) processDevicBigBatch(ctx context.Context, device_id uuid.UUID, packets []snifpacket.SnifPacket) (CHBatch, error) {
	var first_packet_time time.Time
	var last_packet_time time.Time

	for i, packet := range packets {
		packet_time := time.Unix(int64(packet.Timestamp), 0)
		if i == 0 || packet_time.Before(first_packet_time) {
			first_packet_time = packet_time.UTC()
		}
		if i == 0 || packet_time.After(last_packet_time) {
			last_packet_time = packet_time.UTC()
		}
	}

	sec := first_packet_time.Unix()
	rem := sec % 5
	first_packet_time = time.Unix(sec-rem, 0).UTC()

	sec = last_packet_time.Unix()
	rem = sec % 5
	if rem == 0 && last_packet_time.Nanosecond() == 0 {
		last_packet_time = last_packet_time.UTC()
	} else {
		last_packet_time = time.Unix(sec+(5-rem), 0).UTC()
	}

	batches := make([]Batch, 0)

	interval := 5 * time.Second
	for cur := first_packet_time; cur.Before(last_packet_time); cur = cur.Add(interval) {
		batches = append(batches, Batch{
			From: cur,
			To:   cur.Add(interval),
		})
	}

	for _, packet := range packets {
		packet_time := time.Unix(int64(packet.Timestamp), 0).UTC()
		for i := range batches {
			if (packet_time.Equal(batches[i].From) || packet_time.After(batches[i].From)) && packet_time.Before(batches[i].To) {
				batches[i].Packets = append(batches[i].Packets, packet)
				break
			}
		}
	}

	chBatch, err := b.getDevicePackets(batches, device_id)
	if err != nil {
		return CHBatch{}, err
	}

	return chBatch, nil
}