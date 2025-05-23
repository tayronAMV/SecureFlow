package logs

import(
	"time"
	
)
type MemoryUsage struct {
	ContainerID     string    `json:"container_id" bson:"container_id"`
	Timestamp       time.Time `json:"timestamp" bson:"timestamp"`
	UsedMemory      int64     `json:"used_memory" bson:"used_memory"`
	MemoryLimit     int64     `json:"memory_limit" bson:"memory_limit"`
	RSS             int64     `json:"rss" bson:"rss"`
	CacheMemory     int64     `json:"cache_memory" bson:"cache_memory"`
	MemoryUsageRate float64   `json:"memory_usage_rate" bson:"memory_usage_rate"`
	UID string `json : "UID  bson:"UID`
}


type CPUUsage struct {
	ContainerID   string    `json:"container_id" bson:"container_id"`
	Timestamp     time.Time `json:"timestamp" bson:"timestamp"`
	CPUTime       int64     `json:"cpu_time" bson:"cpu_time"`
	CPUUsageRate  float64   `json:"cpu_usage_rate" bson:"cpu_usage_rate"`
	CPULimit      int64     `json:"cpu_limit" bson:"cpu_limit"`
	UID string `json : "UID  bson:"UID`
}

type DiskIOUsage struct {
	ContainerID     string    `json:"container_id" bson:"container_id"`
	Timestamp       time.Time `json:"timestamp" bson:"timestamp"`
	DiskReadBytes   int64     `json:"disk_read_bytes" bson:"disk_read_bytes"`
	DiskWriteBytes  int64     `json:"disk_write_bytes" bson:"disk_write_bytes"`
	DiskUsageRate   float64   `json:"disk_usage_rate" bson:"disk_usage_rate"`
	UID string `json : "UID  bson:"UID`
}

type SyscallEvent struct {
	Pid      uint32    `json:"pid" bson:"pid"`
	Type     uint32    `json:"type" bson:"type"`
	Comm     [16]byte  `json:"comm" bson:"comm"`
	Filename [256]byte `json:"filename" bson:"filename"`
	Timestamp   time.Time  `json:"timestamp" bson:"timestamp"`
	UID string `json : "UID  bson:"UID`

}

type FlowEvent struct {
	Timestamp   time.Time  `json:"timestamp" bson:"timestamp"`
	SrcIP       string   `json:"src_ip" bson:"src_ip"`
	DstIP       string   `json:"dst_ip" bson:"dst_ip"`
	SrcPort     uint16   `json:"src_port" bson:"src_port"`
	DstPort     uint16   `json:"dst_port" bson:"dst_port"`
	Protocol    uint8    `json:"protocol" bson:"protocol"`
	Direction   uint8    `json:"direction" bson:"direction"`
	PayloadLen  uint16   `json:"payload_len" bson:"payload_len"`
	DPIProtocol uint8    `json:"dpi_protocol" bson:"dpi_protocol"`
	Reserved1   uint8    `json:"reserved1" bson:"reserved1"`
	Reserved2   uint16   `json:"reserved2" bson:"reserved2"`

	HTTPMethod   [8]byte   `json:"http_method" bson:"http_method"`
	HTTPPath     [64]byte  `json:"http_path" bson:"http_path"`
	DNSQueryName [64]byte  `json:"dns_query_name" bson:"dns_query_name"`
	DNSQueryType uint16    `json:"dns_query_type" bson:"dns_query_type"`
	ICMPType     uint8     `json:"icmp_type" bson:"icmp_type"`

	_   [3]byte `json:"padding" bson:"padding"`
	Pid uint32  `json:"pid" bson:"pid"`
	UID string `json : "UID  bson:"UID`
}

type Anomaly_log struct {
	CPU     float64 `json:"cpu" bson:"cpu"`
	DiskIO  float64 `json:"disk_io" bson:"disk_io"`
	Memory  float64 `json:"memory" bson:"memory"`
	Network float64 `json:"network" bson:"network"`
	Syscall float64 `json:"syscall" bson:"syscall"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"` 
	UID string `json : "UID  bson:"UID`
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

// 0 - Anomaly log 
// 1 - CPU 
//	2 - MEMORY 
// 3 -  DISKIO
// 4- NETWORK
// 5 - SYSCALL 
type Producer_msg struct{
	Body any `json:"body"`
	Id  int  `json:"id"`
}



type RawSyscallEvent struct {
	Pid      uint32
	Type     uint32
	Comm     [16]byte
	Filename [256]byte
	Cgid    uint64
}


type RawFlowEvent struct {
	Timestamp   uint64
	SrcIP       uint32
	DstIP       uint32
	SrcPort     uint16
	DstPort     uint16
	Protocol    uint8
	Direction   uint8
	PayloadLen  uint16
	DpiProtocol uint8
	Reserved1   uint8
	Reserved2   uint16
	Method      [8]byte
	Path        [64]byte
	QueryName   [64]byte
	QueryType   uint16
	IcmpType    uint8
	Pid         uint32
}

