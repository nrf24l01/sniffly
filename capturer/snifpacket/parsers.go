package snifpacket

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/gopacket/gopacket/layers"
)

func parseDNS(dns *layers.DNS, src, dst net.IP) *SnifPacketDetailsDNS {
	if !dns.QR && dns.OpCode == layers.DNSOpCodeQuery {
		for _, q := range dns.Questions {
			return &SnifPacketDetailsDNS{
				Queries:  []string{string(q.Name)},
				IsQuery:  true,
			}
		}
	}
	return nil
}

func parseHTTP(payload []byte, src, dst net.IP, size int) *SnifPacketDetailsHTTP {
	if len(payload) == 0 {
		return nil
	}
	lineEnd := indexOf(payload, []byte("\r\n"))
	if lineEnd < 0 {
		return nil
	}

	method := ""
	path := ""
	host := ""
	body := ""
	sni := ""

	if sp1 := indexOf(payload[:lineEnd], []byte(" ")); sp1 >= 0 {
		method = string(payload[:sp1])
		if sp2rel := indexOf(payload[sp1+1:lineEnd], []byte(" ")); sp2rel >= 0 {
			sp2 := sp1 + 1 + sp2rel
			path = string(payload[sp1+1 : sp2])
		} else {
			path = string(payload[sp1+1 : lineEnd])
		}
	}

	if hIdx := indexOf(payload, []byte("\r\nHost:")); hIdx >= 0 {
		start := hIdx + len("\r\nHost:")
		if start < len(payload) && payload[start] == ' ' {
			start++
		}
		if hEndRel := indexOf(payload[start:], []byte("\r\n")); hEndRel >= 0 {
			host = string(payload[start : start+hEndRel])
		} else {
			host = string(payload[start:])
		}
	} else if hIdx := indexOf(payload, []byte("\r\nhost:")); hIdx >= 0 {
		start := hIdx + len("\r\nhost:")
		if start < len(payload) && payload[start] == ' ' {
			start++
		}
		if hEndRel := indexOf(payload[start:], []byte("\r\n")); hEndRel >= 0 {
			host = string(payload[start : start+hEndRel])
		} else {
			host = string(payload[start:])
		}
	}

	if host == "" && (len(path) > 7 && (path[:7] == "http://" || (len(path) > 8 && path[:8] == "https://"))) {
		start := 0
		if path[:7] == "http://" {
			start = 7
		} else {
			start = 8
		}
		rest := path[start:]
		if slash := indexOf([]byte(rest), []byte("/")); slash >= 0 {
			host = rest[:slash]
		} else {
			host = rest
		}
	}

	if host == "" && method == "CONNECT" {
		if colon := indexOf([]byte(path), []byte(":")); colon >= 0 {
			host = path[:colon]
		} else {
			host = path
		}
	}

	if host != "" {
		if colon := indexOf([]byte(host), []byte(":")); colon >= 0 {
			host = host[:colon]
		}
		sni = host
	}

	body = ""

	return &SnifPacketDetailsHTTP{
		Method: method,
		Host:   host,
		Path:   path,
		Body:   body,
		Sni:    sni,
	}
}

func parseTLSClientHello(payload []byte, src, dst net.IP, size int) *SnifPacketDetailsTLS {
	if len(payload) < 11 {
		return nil
	}

	if payload[0] != 0x16 {
		return nil
	}
	if payload[5] != 0x01 {
		return nil
	}

	var tlsVer string
	if len(payload) >= 11 {
		vMajor := payload[9]
		vMinor := payload[10]
		switch {
		case vMajor == 3 && vMinor == 0:
			tlsVer = "SSL 3.0"
		case vMajor == 3 && vMinor == 1:
			tlsVer = "TLS 1.0"
		case vMajor == 3 && vMinor == 2:
			tlsVer = "TLS 1.1"
		case vMajor == 3 && vMinor == 3:
			tlsVer = "TLS 1.2"
		case vMajor == 3 && vMinor == 4:
			tlsVer = "TLS 1.3"
		default:
			tlsVer = fmt.Sprintf("0x%02x%02x", vMajor, vMinor)
		}
	}

	sni := extractSNI(payload)
	return &SnifPacketDetailsTLS{
		Sni:        sni,
		TLSVersion: tlsVer,
	}
}

func extractSNI(data []byte) string {
	if len(data) < 43 {
		return ""
	}

	sessionIDLenOffset := 43
	if len(data) < sessionIDLenOffset+1 {
		return ""
	}

	sessionIDLen := int(data[sessionIDLenOffset])
	extStart := sessionIDLenOffset + 1 + sessionIDLen

	if len(data) < extStart+2 {
		return ""
	}

	csLen := int(binary.BigEndian.Uint16(data[extStart:]))
	extPos := extStart + 2 + csLen + 2

	if len(data) < extPos+2 {
		return ""
	}

	extTotalLen := int(binary.BigEndian.Uint16(data[extPos:]))
	extPos += 2

	end := extPos + extTotalLen
	if end > len(data) {
		end = len(data)
	}

	for p := extPos; p+4 <= end; {
		extType := binary.BigEndian.Uint16(data[p:])
		extLen := int(binary.BigEndian.Uint16(data[p+2:]))
		p += 4

		if extType == 0x00 {
			if p+5 <= end {
				nameLen := int(binary.BigEndian.Uint16(data[p+3:]))
				if p+5+nameLen <= end {
					return string(data[p+5 : p+5+nameLen])
				}
			}
			return ""
		}
		p += extLen
	}

	return ""
}