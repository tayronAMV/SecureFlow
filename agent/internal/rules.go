package internal

import(
	"net"
	"fmt"
	"encoding/binary"
	"agent/pkg/logs"
	"github.com/cilium/ebpf"
	"log"
)







func ipStringToUint32(ipStr string) (uint32) {
	ip := net.ParseIP(ipStr).To4()
	if ip == nil {
		return 0
	}
	return binary.BigEndian.Uint32(ip)
}

func protocolStringToUint8(proto string) (uint8, error) {
	switch proto {
	case "ICMP":
		return 1, nil
	case "TCP":
		return 6, nil
	case "UDP":
		return 17, nil
	default:
		return 0, fmt.Errorf("unknown protocol: %s", proto)
	}
}

func dpiProtocolStringToUint8(s string) (uint8, error) {
	switch s {
	case "Unknown":
		return 0, nil
	case "HTTP":
		return 1, nil
	case "DNS":
		return 2, nil
	case "ICMP":
		return 3, nil
	default:
		return 0, fmt.Errorf("unknown DPI protocol: %s", s)
	}
}

func directionStringToUint8(dir string) (uint8) {
	switch dir {
	case "Egress":
		return 0
	case "Ingress":
		return 1 
	default:
		return 0
	}
}
