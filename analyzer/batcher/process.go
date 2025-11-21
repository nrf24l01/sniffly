package batcher

import (
	"context"
	"log"
	"time"

	"github.com/nrf24l01/sniffly/capturer/snifpacket"
)

func (b *Batcher) Process(ctx context.Context, batch Batch) error {
	// Grouping packets by device MAC
	per_device_mac := make(map[string][]snifpacket.SnifPacket)
	for _, packet := range batch.Packets {
		per_device_mac[packet.SrcMAC] = append(per_device_mac[packet.SrcMAC], packet)
	}

	// Retrieving or creating device IDs
	per_device_mac_device_id := make(map[string]uint64)
	for device_id, _ := range per_device_mac {
		rows, err := b.CHDB.CH.Query(ctx, "SELECT device_id FROM device_info WHERE mac = ?", device_id)
		if err != nil {
			return err
		}
		var found_device_id uint64
		if rows.Next() {
			err = rows.Scan(&found_device_id)
			if err != nil {
				return err
			}
			per_device_mac_device_id[device_id] = found_device_id
		} else {
			new_id := b.SnowflakeNode.Generate().Int64()
			err := b.CHDB.CH.Exec(ctx, "INSERT INTO device_info (device_id, mac) VALUES (?, ?)", new_id, device_id)
			if err != nil {
				return err
			}
			per_device_mac_device_id[device_id] = uint64(new_id)
		}
	}
	
	// Grouping packets by device ID
	per_device_id := make(map[uint64][]snifpacket.SnifPacket)
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

func (b *Batcher) processDevicBigBatch(ctx context.Context, device_id uint64, packets []snifpacket.SnifPacket) (CHBatch, error) {
	var first_packet_time time.Time
	var last_packet_time time.Time

	for i, packet := range packets {
		packet_time := time.Unix(int64(packet.Timestamp), 0)
		log.Printf("Packet time: %s, before processing: %d", packet_time.String(), packet.Timestamp)
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
		packet_time := time.Unix(0, int64(packet.Timestamp)*int64(time.Microsecond)).UTC()
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