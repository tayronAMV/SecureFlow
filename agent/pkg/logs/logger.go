package logs

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	
	"bytes"

)


func Encode_string(s string)[]byte{
	body, err := json.Marshal(s)
	if err != nil {
		log.Printf("❌ JSON marshal failed: %v", err)
		return nil
	}

	return body 
	
}

func Decode_string(body []byte) string {
	var s string
	err := json.Unmarshal(body, &s)
	if err != nil {
		log.Printf("❌ JSON unmarshal failed: %v", err)
		return ""
	}
	return s

}
func (a Anomaly_log) Encode()[]byte{
	body, err := json.Marshal(a)
	if err != nil {
		log.Printf("❌ JSON marshal failed: %v", err)
		return nil
	}

	return body 
}

func DecodeAnomalyLog(data []byte) (Anomaly_log) {
	var a Anomaly_log
	err := json.Unmarshal(data, &a)
	if err != nil {
		log.Printf("❌ JSON unmarshal failed: %v", err)
		return Anomaly_log{}
	}
	return a
}
func (m MemoryUsage) String() string {
	return fmt.Sprintf(
		"Memory Usage [UID: %s]\n"+
			"Timestamp:        %s\n"+
			"Used Memory:      %d bytes\n"+
			"Memory Limit:     %d bytes\n"+
			"RSS:              %d bytes\n"+
			"Cache Memory:     %d bytes\n"+
			"Memory Usage Rate: %.2f%%\n",
		m.UID,
		m.Timestamp.Format(time.RFC3339),
		m.UsedMemory,
		m.MemoryLimit,
		m.RSS,
		m.CacheMemory,
		m.MemoryUsageRate*100,
	)
}

func (d DiskIOUsage) String() string {
	return fmt.Sprintf(
		"Disk I/O Usage [UID: %s]\n"+
			"Timestamp:         %s\n"+
			"Disk Read Bytes:   %d\n"+
			"Disk Write Bytes:  %d\n"+
			"Disk Usage Rate:   %.2f%%\n",
		d.UID,
		d.Timestamp.Format(time.RFC3339),
		d.DiskReadBytes,
		d.DiskWriteBytes,
		d.DiskUsageRate*100,
	)
}

func (c CPUUsage) String() string {
	return fmt.Sprintf(
		"CPU Usage [UID: %s]\n"+
			"Timestamp:        %s\n"+
			"CPU Time:         %d ns\n"+
			"CPU Usage Rate:   %.2f%%\n"+
			"CPU Limit:        %d units\n",
		c.UID,
		c.Timestamp.Format(time.RFC3339),
		c.CPUTime,
		c.CPUUsageRate*100,
		c.CPULimit,
	)
}

func (e RawSyscallEvent) String() string {
	eventTypes := map[uint32]string{
		1: "execve",
		2: "execveat",
		3: "open",
		4: "unlink",
		5: "chmod",
		6: "mount",
		7: "setuid",
		8: "socket",
		9: "connect",
	}

	eventTypeStr, ok := eventTypes[e.Type]
	if !ok {
		eventTypeStr = fmt.Sprintf("unknown(%d)", e.Type)
	}

	return fmt.Sprintf(
		"Syscall [%s] PID: %d COMM: %s FILE: %s CGID: %d",
		eventTypeStr,
		e.Pid,
		bytes.TrimRight(e.Comm[:], "\x00"),
		bytes.TrimRight(e.Filename[:], "\x00"),
		e.Cgid,
	)
}

func (event *FlowEvent) String() string {
	srcIP := ipToString(event.SrcIP)
	dstIP := ipToString(event.DstIP)
	proto := protocolToString(event.Protocol)
	dir := directionToString(event.Direction)
	dpi := dpiProtocolToString(event.DpiProtocol)

	// Base info
	result := fmt.Sprintf("📦 [%s] %s %s:%d -> %s:%d (%s) Len=%d",
		dir, proto, srcIP, event.SrcPort, dstIP, event.DstPort, dpi, event.PayloadLen)

	// Protocol-specific fields
	switch event.DpiProtocol {
	case 1: // HTTP
		if event.Method[0] != 0 {
			method := nullTerminatedString(event.Method[:])
			path := nullTerminatedString(event.Path[:])
			result += fmt.Sprintf(" [%s %s]", method, path)
		}
	case 2: // DNS
		if event.QueryName[0] != 0 {
			queryName := nullTerminatedString(event.QueryName[:])
			result += fmt.Sprintf(" [Query: %s Type: %d]", queryName, event.QueryType)
		}
	case 3: // ICMP
		result += fmt.Sprintf(" [Type: %d]", event.IcmpType)
	}

	result += fmt.Sprintf(" @%d", event.Timestamp)
	return result
}


func ipToString(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip), byte(ip>>8), byte(ip>>16), byte(ip>>24))
}

func protocolToString(proto uint8) string {
	switch proto {
	case 1:
		return "ICMP"
	case 6:
		return "TCP"
	case 17:
		return "UDP"
	default:
		return fmt.Sprintf("Proto_%d", proto)
	}
}

func dpiProtocolToString(dpi uint8) string {
	switch dpi {
	case 0:
		return "Unknown"
	case 1:
		return "HTTP"
	case 2:
		return "DNS"
	case 3:
		return "ICMP"
	default:
		return fmt.Sprintf("DPI_%d", dpi)
	}
}

func directionToString(dir uint8) string {
	if dir == 0 {
		return "Egress"
	}
	return "Ingress"
}

func nullTerminatedString(b []byte) string {
	for i, v := range b {
		if v == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}


func (a Anomaly_log) String() string {
	return fmt.Sprintf(
		"📊 AnomalyLog [Pod=%s] CPU=%.2f, Memory=%.2f, DiskIO=%.2f, Network=%.2f, Syscalls=%.2f @%s",
		a.Container.PodName,
		a.CPU,
		a.Memory,
		a.DiskIO,
		a.Network,
		a.Syscall,
		a.Timestamp.Format(time.RFC3339),
	)
}


func UnmarshalFlowRules(data []byte) ([]FlowRule) {
	var rules []FlowRule
	err := json.Unmarshal(data, &rules)
	if err != nil {
		log.Printf("❌ Failed to unmarshal rules: %v", err)
		return nil
	} 
	return rules 
}

func StringToFixed8(s string) [8]byte {
	var arr [8]byte
	copy(arr[:], s)
	return arr
}

func StringToFixed64(s string) [64]byte {
	var arr [64]byte
	copy(arr[:], s)
	return arr
}


func ConvertToFlowRule(in FlowRuleInput) FlowRule {
    return FlowRule{
        SrcIP:       in.SrcIP,
        DstIP:       in.DstIP,
        SrcPort:     in.SrcPort,
        DstPort:     in.DstPort,
        Protocol:    in.Protocol,
        Direction:   in.Direction,
        DpiProtocol: in.DpiProtocol,
        Action:      in.Action,
        Method:      StringToFixed8(in.Method),
        Path:        StringToFixed64(in.Path),
        QueryName:   StringToFixed64(in.QueryName),
        QueryType:   in.QueryType,
        IcmpType:    in.IcmpType,
    }
}