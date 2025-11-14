package snifpacket

type SnifPacketType int

const (
	SnifPacketTypeHTTP SnifPacketType = iota
	SnifPacketTypeHTTPS
	SnifPacketTypeDNS
	SnifPacketTypeFTP
)

type SnifPacketDetailsHTTP struct {
	Method     string                  `json:"method"`
	Host       string                  `json:"host"`
	Path       string                  `json:"path"`
	Body       string                  `json:"body"`
	Sni        string                  `json:"sni"`
}

type SnifPacketDetailsHTTPS struct {
	Sni        string                  `json:"sni"`
	TLSVersion string                  `json:"tls_version"`
}

type SnifPacketDetailsDNS struct {
	Queries    []string                `json:"queries"`
	IsQuery    bool                    `json:"is_query"`
}

type SnifPacketDetailsFTP struct {
	Command    string                  `json:"command"`
	Args       string                  `json:"args"`
}

type SnifPacketDetails struct {
	HTTP       *SnifPacketDetailsHTTP  `json:"http,omitempty"`
	HTTPS      *SnifPacketDetailsHTTPS `json:"https,omitempty"`
	DNS        *SnifPacketDetailsDNS   `json:"dns,omitempty"`
	FTP        *SnifPacketDetailsFTP   `json:"ftp,omitempty"`
	Type       SnifPacketType          `json:"type"`
}

type SnifPacket struct {
	SrcIP      string                  `json:"src_ip"`
	DstIP      string                  `json:"dst_ip"`
	SrcPort    string                  `json:"src_port"`
	DstPort    string                  `json:"dst_port"`
	Size       int                     `json:"size"`
	Protocol   string                  `json:"protocol"`
	Details    SnifPacketDetails       `json:"details"`
}