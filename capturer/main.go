package main

import (
	"log"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/pcap"
)

func main() {
    iface := "wlp0s20f3"
    handle, err := pcap.OpenLive(iface, 1600, true, pcap.BlockForever)
    if err != nil {
        log.Fatal(err)
    }
    defer handle.Close()

    packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
    
    for packet := range packetSource.Packets() {
        processPacket(packet)
    }
}
