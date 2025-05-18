package internal

import (
	"agent/pkg/kube"
	"agent/pkg/logs"
	"agent/pkg/utils"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

)

// -----------------------------
// MEMORY USAGE
// -----------------------------

// GetMemoryUsage reads memory stats for a container by PID (cgroup v2 only)
func GetMemoryUsage(containerID string, pid int) (*logs.MemoryUsage, error) {
	// preserve display ID
	displayID := containerID
	// resolve the actual cgroup v2 path for this PID
	cgroupPath, err := ResolvePathForPID(pid)
	if err != nil {
		return nil, fmt.Errorf("memory: could not resolve cgroup for PID %d: %w", pid, err)
	}

	// helper to read a single integer file
	readInt := func(path string) int64 {
		data, err := os.ReadFile(filepath.Join(cgroupPath, path))
		if err != nil {
			return 0
		}
		val, _ := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
		return val
	}

	// parse memory.stat for keys
	parseStat := func(key string) int64 {
		data, err := os.ReadFile(filepath.Join(cgroupPath, "memory.stat"))
		if err != nil {
			return 0
		}
		for _, line := range strings.Split(string(data), "\n") {
			parts := strings.Fields(line)
			if len(parts) == 2 && parts[0] == key {
				val, _ := strconv.ParseInt(parts[1], 10, 64)
				return val
			}
		}
		return 0
	}

	used := readInt("memory.current")

	// memory.max may be "max" or a number
	var limit int64
	if raw, err := os.ReadFile(filepath.Join(cgroupPath, "memory.max")); err == nil {
		if s := strings.TrimSpace(string(raw)); s != "max" {
			limit, _ = strconv.ParseInt(s, 10, 64)
		}
	}

	rss := parseStat("rss")
	fileCache := parseStat("file")

	usage := &logs.MemoryUsage{
		ContainerID:     displayID,
		Timestamp:       time.Now(),
		UsedMemory:      used,
		MemoryLimit:     limit,
		RSS:             rss,
		CacheMemory:     fileCache,
		MemoryUsageRate: 0,
	}
	if limit > 0 {
		usage.MemoryUsageRate = float64(used) / float64(limit)
	}
	return usage, nil
}

// -----------------------------
// CPU USAGE
// -----------------------------

// GetCPUUsage reads CPU stats for a container by PID (cgroup v2 only)
func GetCPUUsage(containerID string, pid int) (*logs.CPUUsage, error) {
	displayID := containerID
	cgroupPath, err := ResolvePathForPID(pid)
	if err != nil {
		return nil, fmt.Errorf("cpu: could not resolve cgroup for PID %d: %w", pid, err)
	}

	// read usage_usec from cpu.stat
	readCPUStat := func() int64 {
		data, err := os.ReadFile(filepath.Join(cgroupPath, "cpu.stat"))
		if err != nil {
			return 0
		}
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "usage_usec") {
				parts := strings.Fields(line)
				if len(parts) == 2 {
					val, _ := strconv.ParseInt(parts[1], 10, 64)
					return val * 1000 // convert usec -> nsec
				}
			}
		}
		return 0
	}

	// read quota from cpu.max
	readCPULimit := func() int64 {
		data, err := os.ReadFile(filepath.Join(cgroupPath, "cpu.max"))
		if err != nil {
			return 0
		}
		parts := strings.Fields(string(data))
		if len(parts) == 2 && parts[0] != "max" {
			val, _ := strconv.ParseInt(parts[0], 10, 64)
			return val
		}
		return 0
	}

	cpuTime := readCPUStat()
	limit := readCPULimit()

	usage := &logs.CPUUsage{
		ContainerID:  displayID,
		Timestamp:    time.Now(),
		CPUTime:      cpuTime,
		CPULimit:     limit,
		CPUUsageRate: 0,
	}
	return usage, nil
}

// -----------------------------
// DISK I/O USAGE
// -----------------------------

// GetDiskIOUsage reads disk I/O stats for a container by PID (cgroup v2 only)
func GetDiskIOUsage(containerID string, pid int) (*logs.DiskIOUsage, error) {
	displayID := containerID
	cgroupPath, err := ResolvePathForPID(pid)
	if err != nil {
		return nil, fmt.Errorf("disk: could not resolve cgroup for PID %d: %w", pid, err)
	}

	// parse io.stat for rbytes and wbytes
	parseIO := func() (int64, int64) {
		data, err := os.ReadFile(filepath.Join(cgroupPath, "io.stat"))
		if err != nil {
			return 0, 0
		}
		var rTotal, wTotal int64
		for _, line := range strings.Split(string(data), "\n") {
			for _, field := range strings.Fields(line) {
				if strings.HasPrefix(field, "rbytes=") {
					val, _ := strconv.ParseInt(strings.TrimPrefix(field, "rbytes="), 10, 64)
					rTotal += val
				} else if strings.HasPrefix(field, "wbytes=") {
					val, _ := strconv.ParseInt(strings.TrimPrefix(field, "wbytes="), 10, 64)
					wTotal += val
				}
			}
		}
		return rTotal, wTotal
	}

	read, write := parseIO()
	usage := &logs.DiskIOUsage{
		ContainerID:    displayID,
		Timestamp:      time.Now(),
		DiskReadBytes:  read,
		DiskWriteBytes: write,
		DiskUsageRate:  0,
	}
	return usage, nil
}


func StartResourceCollector(mappings []kube.ContainerMapping) {
	go func() {
		fmt.Println("📊 Starting resource collector (CPU, Memory, Disk)...")
		for {
			for _, m := range mappings {
				// Each function updates the global anomaly map
				if err := CollectAndUpdateCPU(m.ContainerID, m.PID); err != nil {
					fmt.Printf("⚠️ CPU update failed for %s: %v", m.ContainerID, err)
				}
				if err := CollectAndUpdateMemory(m.ContainerID, m.PID); err != nil {
					fmt.Printf("⚠️ Memory update failed for %s: %v", m.ContainerID, err)
				}
				if err := CollectAndUpdateDisk(m.ContainerID, m.PID); err != nil {
					fmt.Printf("⚠️ Disk update failed for %s: %v", m.ContainerID, err)
				}
			}



			time.Sleep(1 * time.Second) // wait before the next interval
		}
	}()
}

func CollectAndUpdateCPU(containerID string, pid int) error{
	cur, err := GetCPUUsage(containerID, pid)
	if err != nil {
		fmt.Printf("❌ CPU collect error for %s: %v", containerID, err)
		return err
	}
	UID := kube.PidToUid(pid)

	if _,ok := utils.Container_uid_map[UID]; !ok{
		utils.Container_uid_map[UID] = &logs.Anomaly_log{}
	}

	if prev, ok := utils.CpuTrackers[UID]; ok {
		deltaTime := cur.Timestamp.Sub(prev.PrevTime).Seconds()
		if deltaTime > 0 {
			deltaCPU := float64(cur.CPUTime - prev.PrevCPUTime)
			utils.Container_uid_map[UID].CPU = deltaCPU / deltaTime
		}
	}

	utils.CpuTrackers[UID] = logs.CpuTracker{
		PrevTime:    cur.Timestamp,
		PrevCPUTime: cur.CPUTime,
	}
	cur.UID = UID 
	// send to DB! 
	logs.Producer(logs.Producer_msg{
		Body: *cur,
		Id : 1, 
	})


	
	return nil
}

func CollectAndUpdateDisk(containerID string, pid int) error {
	cur, err := GetDiskIOUsage(containerID, pid)
	if err != nil {
		fmt.Printf("❌ Disk I/O collect error for %s: %v\n", containerID, err)
		return err
	}
	UID := kube.PidToUid(pid)
	if _,ok := utils.Container_uid_map[UID]; !ok{
		utils.Container_uid_map[UID] = &logs.Anomaly_log{}
	}

	if prev, ok := utils.DiskTrackers[UID]; ok {
		deltaTime := cur.Timestamp.Sub(prev.PrevTime).Seconds()
		if deltaTime > 0 {
			deltaRead := cur.DiskReadBytes - prev.PrevReadBytes
			deltaWrite := cur.DiskWriteBytes - prev.PrevWriteBytes
			utils.Container_uid_map[UID].DiskIO = float64(deltaRead+deltaWrite) / deltaTime
		}
	}

	utils.DiskTrackers[UID] = logs.DiskTracker{
		PrevTime:       cur.Timestamp,
		PrevReadBytes:  cur.DiskReadBytes,
		PrevWriteBytes: cur.DiskWriteBytes,
	}
	cur.UID = UID 
	// send to server
	logs.Producer(logs.Producer_msg{
		Body: *cur,
		Id : 3, 
	})

	return nil
}


func CollectAndUpdateMemory(containerID string, pid int) error {
	cur, err := GetMemoryUsage(containerID, pid)
	if err != nil {
		fmt.Printf("❌ Memory collect error for %s: %v\n", containerID, err)
		return err
	}
	UID := kube.PidToUid(pid)
	if _,ok := utils.Container_uid_map[UID]; !ok{
		utils.Container_uid_map[UID] = &logs.Anomaly_log{}
	}
	if prev, ok := utils.MemoryTrackers[UID]; ok {
		deltaTime := cur.Timestamp.Sub(prev.PrevTime).Seconds()
		if deltaTime > 0 {
			deltaMem := float64(cur.UsedMemory - prev.PrevUsedBytes)
			utils.Container_uid_map[UID].Memory = deltaMem / deltaTime
		}
	}

	utils.MemoryTrackers[UID] = logs.MemoryTracker{
		PrevTime:      cur.Timestamp,
		PrevUsedBytes: cur.UsedMemory,
	}
	cur.UID = UID 

	logs.Producer(logs.Producer_msg{
		Body: *cur,
		Id : 2, 
	})

	return nil
}