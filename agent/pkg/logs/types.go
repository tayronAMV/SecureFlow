package logs

import(
	"time"
)

type MemoryUsage struct {
	ContainerID     string
	Timestamp       time.Time
	UsedMemory      int64
	MemoryLimit     int64
	RSS             int64
	CacheMemory     int64
	MemoryUsageRate float64
}

type CPUUsage struct {
	ContainerID   string
	Timestamp     time.Time
	CPUTime       int64
	CPUUsageRate  float64
	CPULimit      int64
}

type DiskIOUsage struct {
	ContainerID     string
	Timestamp       time.Time
	DiskReadBytes   int64
	DiskWriteBytes  int64
	DiskUsageRate   float64
}


type FlowEvent struct {
	Timestamp   uint64   // __u64
	SrcIP       uint32   // __u32
	DstIP       uint32   // __u32
	SrcPort     uint16   // __u16
	DstPort     uint16   // __u16
	Protocol    uint8    // __u8
	Direction   uint8    // __u8
	PayloadLen  uint16   // __u16
	DPIProtocol uint8    // __u8
	Reserved1   uint8    // __u8
	Reserved2   uint16   // __u16

	// Union of HTTP, DNS, ICMP – we'll just read them all, even though only one is valid per event
	HTTPMethod [8]byte   // char[8]
	HTTPPath   [64]byte  // char[64]

	DNSQueryName [64]byte // char[64]
	DNSQueryType uint16   // __u16

	ICMPType uint8   // __u8

	// Padding to ensure struct size alignment (total size = 8 + 4 + 4 + 2 + 2 + 1 + 1 + 2 + 1 + 1 + 2 + 8 + 64 + 64 + 2 + 1 = 157 bytes)
	_ [3]byte // padding to make it multiple of 8 (optional depending on kernel + Go alignment)
	pid 	uint32
}

type Anomaly_log struct {
	CPU     float64 `json:"cpu"`
	DiskIO  float64 `json:"disk_io"`
	Memory  float64 `json:"memory"`
	Network float64 `json:"network"`
	Syscall float64 `json:"syscall"`
}

