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


type TrafficInfo struct {
	SrcIP    [4]byte
	DstIP    [4]byte
	SrcPort  uint16
	DstPort  uint16
	Protocol uint8
	PktLen   uint64
	_        [3]byte // Padding
}
