package snifpacket

import (
	"fmt"
	"net"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
)

func ProcessPacket(packet gopacket.Packet) (*SnifPacket, error) {
	// Ethernet
	ethLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethLayer == nil {
		return nil, fmt.Errorf("no ethernet layer found")
	}

	// IP
	ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
	ipv6Layer := packet.Layer(layers.LayerTypeIPv6)

	var srcIP, dstIP net.IP

	if ipv4Layer != nil {
		ip := ipv4Layer.(*layers.IPv4)
		srcIP = ip.SrcIP
		dstIP = ip.DstIP
	} else if ipv6Layer != nil {
		ip := ipv6Layer.(*layers.IPv6)
		srcIP = ip.SrcIP
		dstIP = ip.DstIP
	} else {
		return nil, fmt.Errorf("no IP layer found")
	}

	snif_packet := &SnifPacket{
		SrcIP:      srcIP.String(),
		DstIP:      dstIP.String(),
		SrcMAC:     ethLayer.(*layers.Ethernet).SrcMAC.String(),
		DstMAC:     ethLayer.(*layers.Ethernet).DstMAC.String(),
		Timestamp:  packet.Metadata().Timestamp.Unix(),
	}

	// UDP → DNS?
	if udp := packet.Layer(layers.LayerTypeUDP); udp != nil {
		u := udp.(*layers.UDP)
		snif_packet.SrcPort = u.SrcPort.String()
		snif_packet.DstPort = u.DstPort.String()
		snif_packet.Size = len(u.Payload)
		snif_packet.Protocol = "UDP"
		if u.DstPort == 53 {
			if dns := packet.Layer(layers.LayerTypeDNS); dns != nil {
				details := parseDNS(dns.(*layers.DNS), srcIP, dstIP)
				if details != nil {
					snif_packet.Details.DNS = details
					snif_packet.Details.Type = SnifPacketTypeDNS
					return snif_packet, nil
				}
			}
		}
		snif_packet.Details.Type = SnifPacketTypeUDP
		return snif_packet, nil
	}

	if tcp := packet.Layer(layers.LayerTypeTCP); tcp != nil {
		t := tcp.(*layers.TCP)
		payload := t.Payload
		size := len(payload)

		snif_packet.SrcPort = t.SrcPort.String()
		snif_packet.DstPort = t.DstPort.String()
		snif_packet.Size = size
		snif_packet.Protocol = "TCP"

		// HTTP (port 80)
		if t.DstPort == 80 {
			details := parseHTTP(payload, srcIP, dstIP, size)
			if details != nil {
				snif_packet.Details.HTTP = details
				snif_packet.Details.Type = SnifPacketTypeHTTP
				return snif_packet, nil
			}
		}

		// HTTPS (port 443 → TLS ClientHello)
		if t.DstPort == 443 {
			details := parseTLSClientHello(payload, srcIP, dstIP, size)
			if details != nil {
				snif_packet.Details.TLS = details
				snif_packet.Details.Type = SnifPacketTypeTLS
				return snif_packet, nil
			}
		}

		// On plain TCP
		snif_packet.Details.Type = SnifPacketTypeTCP
		return snif_packet, nil
	}
	return nil, fmt.Errorf("no TCP/UDP layer found")
}