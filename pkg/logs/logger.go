package logs

import (
	"fmt"
	"log"
)

// ANSI colors

const (
	colorReset  = "\033[0m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorPurple = "\033[35m"
)

// Traffic Event (Blue)
func LogTraffic(event *TrafficInfo) {
	log.Printf(colorBlue+"📡 %s:%d → %s:%d | proto: %d | size: %d bytes "+colorReset,
		ipToString(event.SrcIP), event.SrcPort,
		ipToString(event.DstIP), event.DstPort,
		event.Protocol, event.PktLen,
	)
}

// Memory Usage (Green)
func LogMemory(mem *MemoryUsage) {
	log.Printf(colorGreen+"🧠 [%s] Memory used: %d / %d (%.2f%%) | RSS: %d | Cache: %d"+colorReset,
		mem.ContainerID, mem.UsedMemory, mem.MemoryLimit,
		mem.MemoryUsageRate*100, mem.RSS, mem.CacheMemory,
	)
}

// CPU Usage (Yellow)
func LogCPU(cpu *CPUUsage) {
	log.Printf(colorYellow+"⚙️ [%s] CPU time: %d ms | Usage rate: %.2f%% | Limit: %d"+colorReset,
		cpu.ContainerID, cpu.CPUTime, cpu.CPUUsageRate*100, cpu.CPULimit,
	)
}

// Disk I/O Usage (Purple)
func LogDisk(disk *DiskIOUsage) {
	log.Printf(colorPurple+"💾 [%s] Disk R: %d B, W: %d B | Usage rate: %.2f B/s"+colorReset,
		disk.ContainerID, disk.DiskReadBytes, disk.DiskWriteBytes, disk.DiskUsageRate,
	)
}

// Utility
func ipToString(ip [4]byte) string {
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}
