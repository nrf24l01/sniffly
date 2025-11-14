package snifpacket

import "github.com/gopacket/gopacket"

func LoadSnifPacket(pck gopacket.Packet) *SnifPacket {
	return &SnifPacket{}
}

