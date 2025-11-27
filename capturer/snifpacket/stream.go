package snifpacket

import (
	"log"
	"sync"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/pcap"
)

func ReceivePackets(handle *pcap.Handle, iface string, packets chan *SnifPacket, wg *sync.WaitGroup) {
    defer wg.Done()
    // Close packets channel when this sender exits so receivers can finish cleanly.
    // ReceivePackets is the only sender on the channel, so it's safe to close it here.
    defer close(packets)

    packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

    localIPs, localMAC, err := GetLocalAddrs(iface)
    filterEnabled := true
    if err != nil {
        log.Printf("failed to get local addresses for interface %s: %v; outgoing filtering disabled", iface, err)
        filterEnabled = false
    }

    for packet := range packetSource.Packets() {
        sp, err := ProcessPacket(packet)
        if err != nil {
            log.Printf("Error processing packet: %v", err)
            continue
        }

        if filterEnabled {
            if _, ok := localIPs[sp.SrcIP]; !ok && sp.SrcMAC != localMAC {
                continue
            }
        }

        packets <- sp
    }
    log.Printf("Packet receiving goroutine for interface %s exiting", iface)
}