package snifpacket

type SnifPacketType int

const (
	SnifPacketTypeHTTP SnifPacketType = iota
	SnifPacketTypeTLS
	SnifPacketTypeDNS
	SnifPacketTypeFTP
	SnifPacketTypeTCP
	SnifPacketTypeUDP
)

type SnifPacketDetailsHTTP struct {
	Method     string                  `json:"method"`
	Host       string                  `json:"host"`
	Path       string                  `json:"path"`
	Body       string                  `json:"body"`
	Sni        string                  `json:"sni"`
}

type SnifPacketDetailsTLS struct {
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

type SnifPacketDetailsTCP struct {
	Data 	   []byte                  `json:"data"`
}

type SnifPacketDetailsUDP struct {
	Data 	   []byte                  `json:"data"`
}

type SnifPacketDetails struct {
	HTTP       *SnifPacketDetailsHTTP  `json:"http,omitempty"`
	TLS        *SnifPacketDetailsTLS   `json:"tls,omitempty"`
	DNS        *SnifPacketDetailsDNS   `json:"dns,omitempty"`
	FTP        *SnifPacketDetailsFTP   `json:"ftp,omitempty"`
	TCP 	     *SnifPacketDetailsTCP   `json:"tcp,omitempty"`
	UDP        *SnifPacketDetailsUDP   `json:"udp,omitempty"`
	Type       SnifPacketType          `json:"type"`
}

type SnifPacket struct {
	SrcIP      string                  `json:"src_ip"`
	DstIP      string                  `json:"dst_ip"`
	SrcMAC     string				           `json:"src_mac"`
	DstMAC     string                  `json:"dst_mac"`
	SrcPort    string                  `json:"src_port"`
	DstPort    string                  `json:"dst_port"`
	Size       int                     `json:"size"`
	Protocol   string                  `json:"protocol"`
	Details    SnifPacketDetails       `json:"details"`
	Timestamp  int64                   `json:"timestamp"`
}