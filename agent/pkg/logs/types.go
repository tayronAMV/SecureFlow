package logs

import(
	"time"
)

type MemoryUsage struct {
	ContainerID     string    `bson:"container_id"`
	Timestamp       time.Time `bson:"timestamp"`
	UsedMemory      int64     `bson:"used_memory"`
	MemoryLimit     int64     `bson:"memory_limit"`
	RSS             int64     `bson:"rss"`
	CacheMemory     int64     `bson:"cache_memory"`
	MemoryUsageRate float64   `bson:"memory_usage_rate"`
}


type CPUUsage struct {
	ContainerID   string    `bson:"container_id"`
	Timestamp     time.Time `bson:"timestamp"`
	CPUTime       int64     `bson:"cpu_time"`
	CPUUsageRate  float64   `bson:"cpu_usage_rate"`
	CPULimit      int64     `bson:"cpu_limit"`
}


type DiskIOUsage struct {
	ContainerID     string    `bson:"container_id"`
	Timestamp       time.Time `bson:"timestamp"`
	DiskReadBytes   int64     `bson:"disk_read_bytes"`
	DiskWriteBytes  int64     `bson:"disk_write_bytes"`
	DiskUsageRate   float64   `bson:"disk_usage_rate"`
}

type SyscallEvent struct {
	Pid      uint32    `bson:"pid"`
	Type     uint32    `bson:"type"`
	Comm     [16]byte  `bson:"comm"`
	Filename [256]byte `bson:"filename"`
}



type FlowEvent struct {
	Timestamp   uint64   `bson:"timestamp"`     // __u64
	SrcIP       uint32   `bson:"src_ip"`        // __u32
	DstIP       uint32   `bson:"dst_ip"`        // __u32
	SrcPort     uint16   `bson:"src_port"`      // __u16
	DstPort     uint16   `bson:"dst_port"`      // __u16
	Protocol    uint8    `bson:"protocol"`      // __u8
	Direction   uint8    `bson:"direction"`     // __u8
	PayloadLen  uint16   `bson:"payload_len"`   // __u16
	DPIProtocol uint8    `bson:"dpi_protocol"`  // __u8
	Reserved1   uint8    `bson:"reserved1"`     // __u8
	Reserved2   uint16   `bson:"reserved2"`     // __u16

	// Union of HTTP, DNS, ICMP – we'll just read them all, even though only one is valid per event
	HTTPMethod   [8]byte   `bson:"http_method"`    // char[8]
	HTTPPath     [64]byte  `bson:"http_path"`      // char[64]
	DNSQueryName [64]byte  `bson:"dns_query_name"` // char[64]
	DNSQueryType uint16    `bson:"dns_query_type"` // __u16
	ICMPType     uint8     `bson:"icmp_type"`      // __u8

	// Padding to ensure struct size alignment (total size = ...)
	_   [3]byte `bson:"padding"` // optional depending on kernel + Go alignment
	Pid uint32  `bson:"pid"`
}

type Anomaly_log struct {
	CPU     float64 `json:"cpu"`
	DiskIO  float64 `json:"disk_io"`
	Memory  float64 `json:"memory"`
	Network float64 `json:"network"`
	Syscall float64 `json:"syscall"`
}

type CpuTracker struct {
	PrevTime   time.Time
	PrevCPUTime int64
}

type MemoryTracker struct {
	PrevTime      time.Time
	PrevUsedBytes int64
}


type DiskTracker struct {
	PrevTime       time.Time
	PrevReadBytes  int64
	PrevWriteBytes int64
}