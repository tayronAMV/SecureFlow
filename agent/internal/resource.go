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
	"sync"
)

var mu sync.RWMutex


func ResolvePathForPID(pid int) (string, error) {
  
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cgroup", pid))
	if err != nil {
	  return "", fmt.Errorf("could not read cgroup file for PID %d: %w", pid, err)
	}
  
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
	  parts := strings.SplitN(line, ":", 3)
	  if len(parts) != 3 {
		continue
	  }
	  rel := parts[2]
	  if !strings.Contains(rel, "kubepods") {
		continue
	  }
	  rel = strings.TrimPrefix(rel, "/")
	  full := filepath.Join("/sys/fs/cgroup", rel)
	  if fi, err := os.Stat(full); err == nil && fi.IsDir() {
		return full, nil
	  }
	}
	return "", fmt.Errorf("no kubepods cgroup found for PID %d", pid)
  }

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


func StartResourceCollector(logCh chan logs.Producer_msg) {
	mappingCh := make(chan struct{}, 1) // Buffered so sender never blocks

	go func() {
		fmt.Println("📊 Starting resource collector (CPU, Memory, Disk)...")

		mappings := kube.GetCurrentMapping()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-mappingCh:
				mappings = kube.GetCurrentMapping()

			case <-ticker.C:
				for _, m := range mappings {
					if cpu, err := CollectAndUpdateCPU(m, m.PID); err != nil {
						fmt.Printf("⚠️ CPU update failed for %s: %v\n", m.ContainerID, err)
					} else {
						logCh <- logs.Producer_msg{
							Body: logs.Encode_string(cpu),
							Id:   1,
						}
					}

					if mem, err := CollectAndUpdateMemory(m, m.PID); err != nil {
						fmt.Printf("⚠️ Memory update failed for %s: %v\n", m.ContainerID, err)
					} else {
						logCh <- logs.Producer_msg{
							Body: logs.Encode_string(mem),
							Id:   1,
						}
					}

					if disk, err := CollectAndUpdateDisk(m, m.PID); err != nil {
						fmt.Printf("⚠️ Disk update failed for %s: %v\n", m.ContainerID, err)
					} else {
						logCh <- logs.Producer_msg{
							Body: logs.Encode_string(disk),
							Id:   1,
						}
					}
				}
			}
		}
	}()

	// Goroutine that waits for cond to signal
	cond := kube.GetCondition()
	go func() {
		for {
			cond.L.Lock()
			cond.Wait()
			cond.L.Unlock()

			select {
			case mappingCh <- struct{}{}:
			default: // don't block if already waiting
			}
		}
	}()
}

func CollectAndUpdateCPU(container kube.ContainerMapping , pid int) (string,error){
	cur, err := GetCPUUsage(container.ContainerID, pid)
	if err != nil {
		fmt.Printf("❌ CPU collect error for %s: %v", container.ContainerID, err)
		return "",err
	}
	


	utils.Update_uid_Map(container.UID , container)
	utils.Update_cpu_Tracker(container.UID , logs.CpuTracker{
		PrevCPUTime: cur.CPUTime,
		PrevTime: cur.Timestamp,
	} )

	
	cur.UID = container.UID
	// send to DB! 
	


	
	return cur.String() , nil
}

func CollectAndUpdateDisk(container kube.ContainerMapping, pid int) (string,error) {
	cur, err := GetDiskIOUsage(container.ContainerID, pid)
	if err != nil {
		fmt.Printf("❌ Disk I/O collect error for %s: %v\n", container.ContainerID, err)
		return "" , err
	}
	

	utils.Update_uid_Map(container.UID , container)
	utils.Update_disk_Tracker(container.UID , logs.DiskTracker{
		PrevTime: cur.Timestamp,
		PrevReadBytes: cur.DiskReadBytes,
		PrevWriteBytes: cur.DiskWriteBytes,
	})
	cur.UID = container.UID
	// send to server
	

	return cur.String() , nil
}


func CollectAndUpdateMemory(container kube.ContainerMapping, pid int) (string,error){
	cur, err := GetMemoryUsage(container.ContainerID, pid)
	if err != nil {
		fmt.Printf("❌ Memory collect error for %s: %v\n", container.ContainerID, err)
		return "",err
	}
	utils.Update_uid_Map(container.UID , container)
	utils.Update_memory_Tracker(container.UID , logs.MemoryTracker{
		PrevTime: cur.Timestamp,
		PrevUsedBytes: cur.UsedMemory,
	})
	cur.UID = container.UID

	

	return cur.String() , nil
}