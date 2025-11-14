package snifpacket

import (
	"fmt"
	"net"

	"github.com/gopacket/gopacket/layers"
)

func parseDNS(dns *layers.DNS, src, dst net.IP) *SnifPacketDetailsDNS {
	if dns.QR == false && dns.OpCode == layers.DNSOpCodeQuery {
		for _, q := range dns.Questions {
			fmt.Printf("DNS %s → %s query=%s type=%s\n",
				src, dst, string(q.Name), q.Type)
			return &SnifPacketDetailsDNS{
				Queries:  []string{string(q.Name)},
				IsQuery:  true,
			}
		}
	}
}

func parseHTTP(http *layers.HTTP, src, dst net.IP) *SnifPacketDetailsHTTP {