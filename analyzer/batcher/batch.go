package batcher

import (
	"encoding/json"
	"time"

	"github.com/nrf24l01/sniffly/capture_receiver/rabbit"
	"github.com/nrf24l01/sniffly/capturer/snifpacket"
)

type Batch struct {
	Packets []snifpacket.SnifPacket
	From    time.Time
	To      time.Time
}

func (b *Batch) AddMessage(msg []byte) error {
	var packet rabbit.Message
	err := json.Unmarshal(msg, &packet)
	if err != nil {
		return err
	}

	var snifPacket snifpacket.SnifPacket
	err = json.Unmarshal(packet.Payload, &snifPacket)
	if err != nil {
		return err
	}

	b.Packets = append(b.Packets, snifPacket)
	return nil
}