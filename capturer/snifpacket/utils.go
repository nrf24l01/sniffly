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

func GetLocalAddrs(iface string) ([]*net.IPNet, string, error) {
	var nets []*net.IPNet

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
				nets = append(nets, v)
			}
		case *net.IPAddr:
			if v.IP != nil {
				bits := 32
				if v.IP.To4() == nil {
					bits = 128
				}
				mask := net.CIDRMask(bits, bits)
				nets = append(nets, &net.IPNet{IP: v.IP, Mask: mask})
			}
		}
	}

	mac := ""
	if i.HardwareAddr != nil {
		mac = i.HardwareAddr.String()
	}

	return nets, mac, nil
}
