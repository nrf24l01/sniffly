package snifpacket

import (
	"sync"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/pcap"
)

func ReceivePackets(handle *pcap.Handle, packets chan *SnifPacket, wg *sync.WaitGroup) {
	defer wg.Done()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
    for packet := range packetSource.Packets() {
        packet, err := ProcessPacket(packet)
        if err != nil {
            continue
        }
        packets <- packet
    }
}