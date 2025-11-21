package snifpacket

import "net"

func indexOf(haystack, needle []byte) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := range needle {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func GetLocalAddrs(iface string) (map[string]struct{}, string, error) {
	ips := make(map[string]struct{})

	i, err := net.InterfaceByName(iface)
	if err != nil {
		return nil, "", err
	}

	addrs, err := i.Addrs()
	if err != nil {
		return nil, "", err
	}

	for _, a := range addrs {
		switch v := a.(type) {
		case *net.IPNet:
			if v.IP != nil {
				ips[v.IP.String()] = struct{}{}
			}
		case *net.IPAddr:
			if v.IP != nil {
				ips[v.IP.String()] = struct{}{}
			}
		}
	}

	mac := ""
	if i.HardwareAddr != nil {
		mac = i.HardwareAddr.String()
	}

	return ips, mac, nil
}