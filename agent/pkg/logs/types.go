package logs

import(
	"time"
	"agent/pkg/kube"
	
)
type MemoryUsage struct {
	ContainerID     string    `json:"container_id" bson:"container_id"`
	Timestamp       time.Time `json:"timestamp" bson:"timestamp"`
	UsedMemory      int64     `json:"used_memory" bson:"used_memory"`
	MemoryLimit     int64     `json:"memory_limit" bson:"memory_limit"`
	RSS             int64     `json:"rss" bson:"rss"`
	CacheMemory     int64     `json:"cache_memory" bson:"cache_memory"`
	MemoryUsageRate float64   `json:"memory_usage_rate" bson:"memory_usage_rate"`
	UID string 					`json : "UID"  bson:"UID"`
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

}

type FlowEvent struct {
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
			_        [1]byte  // Padding for alignment
		IfIndex     uint32   // Network interface index (for container resolution)

}

type Anomaly_log struct {
	CPU     float64 `json:"cpu" bson:"cpu"`
	DiskIO  float64 `json:"disk_io" bson:"disk_io"`
	Memory  float64 `json:"memory" bson:"memory"`
	Network float64 `json:"network" bson:"network"`
	Syscall float64 `json:"syscall" bson:"syscall"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"` 
	Container kube.ContainerMapping `json:"container" bson:"container"`
}


type CpuTracker struct {
	PrevTime   time.Time
	PrevCPUTime int64
	CPUUsage float64
}

type MemoryTracker struct {
	PrevTime      time.Time
	PrevUsedBytes int64
	MemoryUsage float64
}


type DiskTracker struct {
	PrevTime       time.Time
	PrevReadBytes  int64
	PrevWriteBytes int64
	DiskIOUsage 	float64
}

type Producer_msg struct{
	Body []byte `json:"body"`
	Id  int  `json:"id"`
}



type RawSyscallEvent struct {
	Pid      uint32
	Type     uint32
	Comm     [16]byte
	Filename [256]byte
	Cgid    uint64
}



type FlowRule struct {
    SrcIP       uint32   `json:"src_ip"`
    DstIP       uint32   `json:"dst_ip"`
    SrcPort     uint16   `json:"src_port"`
    DstPort     uint16   `json:"dst_port"`
    Protocol    uint8    `json:"protocol"`
    Direction   uint8    `json:"direction"`
    DpiProtocol uint8    `json:"dpi_protocol"`
    Action      uint8    `json:"action"`

    Method      [8]byte  `json:"method"`
    Path        [64]byte `json:"path"`
    QueryName   [64]byte `json:"query_name"`
    QueryType   uint16   `json:"query_type"`
    IcmpType    uint8    `json:"icmp_type"`
    _           uint8    `json:"-"` // ignored in JSON, used for padding
}


type FlowRuleInput struct {
    SrcIP       uint32 `json:"src_ip"`
    DstIP       uint32 `json:"dst_ip"`
    SrcPort     uint16 `json:"src_port"`
    DstPort     uint16 `json:"dst_port"`
    Protocol    uint8  `json:"protocol"`
    Direction   uint8  `json:"direction"`
    DpiProtocol uint8  `json:"dpi_protocol"`
    Action      uint8  `json:"action"`

    Method      string `json:"method"`      // Will convert to [8]byte
    Path        string `json:"path"`        // Will convert to [64]byte
    QueryName   string `json:"query_name"`  // Will convert to [64]byte
    QueryType   uint16 `json:"query_type"`
    IcmpType    uint8  `json:"icmp_type"`
}


type SyscallEventRule struct {
	Pid       uint32    `json:"pid" bson:"pid"`
	Type      uint32    `json:"type" bson:"type"`
	Comm      [16]byte  `json:"comm" bson:"comm"`
	Filename  [256]byte `json:"filename" bson:"filename"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	Action    int       `json:"action" bson:"action"`
}


type DiskIOUsageRule struct {
	ContainerID     string    `json:"container_id" bson:"container_id"`
	Timestamp       time.Time `json:"timestamp" bson:"timestamp"`
	DiskReadBytes   int64     `json:"disk_read_bytes" bson:"disk_read_bytes"`
	DiskWriteBytes  int64     `json:"disk_write_bytes" bson:"disk_write_bytes"`
	DiskUsageRate   float64   `json:"disk_usage_rate" bson:"disk_usage_rate"`
	UID             string    `json:"UID" bson:"UID"`
	Action          int       `json:"action" bson:"action"`
}


type CPUUsageRule struct {
	ContainerID   string    `json:"container_id" bson:"container_id"`
	Timestamp     time.Time `json:"timestamp" bson:"timestamp"`
	CPUTime       int64     `json:"cpu_time" bson:"cpu_time"`
	CPUUsageRate  float64   `json:"cpu_usage_rate" bson:"cpu_usage_rate"`
	CPULimit      int64     `json:"cpu_limit" bson:"cpu_limit"`
	UID           string    `json:"UID" bson:"UID"`
	Action        int       `json:"action" bson:"action"`
}

type MemoryUsageRule struct {
	ContainerID     string    `json:"container_id" bson:"container_id"`
	Timestamp       time.Time `json:"timestamp" bson:"timestamp"`
	UsedMemory      int64     `json:"used_memory" bson:"used_memory"`
	MemoryLimit     int64     `json:"memory_limit" bson:"memory_limit"`
	RSS             int64     `json:"rss" bson:"rss"`
	CacheMemory     int64     `json:"cache_memory" bson:"cache_memory"`
	MemoryUsageRate float64   `json:"memory_usage_rate" bson:"memory_usage_rate"`
	UID             string    `json:"UID" bson:"UID"`
	Action          int       `json:"action" bson:"action"`
}