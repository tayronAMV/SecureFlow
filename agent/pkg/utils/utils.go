package utils

import (
	"agent/pkg/kube"
	"agent/pkg/logs"
	
	"sync"
	"time"
)

var TIME_INTERVAL = float64(10.0)

const (
	IndexAnomaly = iota // 0
	IndexCPU               // 1
	IndexMemory            // 2
	IndexDiskIO            // 3
	IndexNetwork           // 4
	IndexSyscall           // 5
)

var mu_arr[6]sync.RWMutex

var Container_uid_map = make(map[string]*logs.Anomaly_log)

var CpuTrackers = make(map[string]logs.CpuTracker)
var DiskTrackers = make(map[string]logs.DiskTracker)
var MemoryTrackers = make(map[string]logs.MemoryTracker)
var SyscallTrackers = make(map[string]int)
var NetwrokTracker = make(map[string]float64)

func Update_uid_Map(uid string , container kube.ContainerMapping){
	mu_arr[IndexAnomaly].Lock()
	defer mu_arr[IndexAnomaly].Unlock()
	if _,ok := Container_uid_map[uid] ; !ok{
		Container_uid_map[uid] = &logs.Anomaly_log{
			Container: container,
		}
	}
	

}

func Update_network_Tracker(uid string , packet_len float64 ){
	mu_arr[IndexNetwork].Lock()
	sum , ok := NetwrokTracker[uid]
	if !ok{
		NetwrokTracker[uid] = packet_len
	}else{
		NetwrokTracker[uid] = sum + packet_len
	}
	mu_arr[IndexNetwork].Unlock()

}


func Update_syscall_Tracker(uid string) {
	mu_arr[IndexSyscall].Lock()
	SyscallTrackers[uid]++
	mu_arr[IndexSyscall].Unlock()
}

func Update_disk_Tracker(uid string, cur logs.DiskTracker) {
	mu_arr[IndexDiskIO].Lock()
	defer mu_arr[IndexDiskIO].Unlock()

	var DiskIOUsage float64
	if prev, ok := DiskTrackers[uid]; ok {
		deltaTime := cur.PrevTime.Sub(prev.PrevTime).Seconds()
		if deltaTime > 0 {
			deltaRead := cur.PrevReadBytes - prev.PrevReadBytes
			deltaWrite := cur.PrevWriteBytes - prev.PrevWriteBytes
			
			DiskIOUsage = float64(deltaRead+deltaWrite) / deltaTime
			
		}
	}

	DiskTrackers[uid] = logs.DiskTracker{
		PrevTime:       cur.PrevTime,
		PrevReadBytes:  cur.PrevReadBytes,
		PrevWriteBytes: cur.PrevWriteBytes,
		DiskIOUsage: DiskIOUsage,
	}
}

func Update_cpu_Tracker(uid string, cur logs.CpuTracker) {
	mu_arr[IndexCPU].Lock()
	defer mu_arr[IndexCPU].Unlock()

	
	if prev, ok := CpuTrackers[uid]; ok {
		deltaTime := cur.PrevTime.Sub(prev.PrevTime).Seconds()
		if deltaTime > 0 {
			deltaCPU := float64(cur.PrevCPUTime - prev.PrevCPUTime)
			cur.CPUUsage = deltaCPU / deltaTime
		}
	}
	CpuTrackers[uid] = cur


}

func Update_memory_Tracker(uid string, cur logs.MemoryTracker) {
	mu_arr[IndexMemory].Lock()
	defer mu_arr[IndexMemory].Unlock()

	var MemoryUsage float64
	if prev, ok := MemoryTrackers[uid]; ok {
		deltaTime := cur.PrevTime.Sub(prev.PrevTime).Seconds()
		if deltaTime > 0 {
			deltaMem := float64(cur.PrevUsedBytes - prev.PrevUsedBytes)
			MemoryUsage = deltaMem / deltaTime
			
		}
	}

	MemoryTrackers[uid] = logs.MemoryTracker{
		PrevTime:      cur.PrevTime,
		PrevUsedBytes: cur.PrevUsedBytes,
		MemoryUsage: MemoryUsage,
	}
}

func Anomaly_log_generator(logCh chan logs.Producer_msg) {
	rescanTicker := time.NewTicker(time.Duration(TIME_INTERVAL) * time.Second)
	defer rescanTicker.Stop()

	for {
		select {
		case <-rescanTicker.C:
			mu_arr[IndexAnomaly].RLock()
			for uid, a := range Container_uid_map {

				// === Network ===
				mu_arr[IndexNetwork].RLock()
				netSum := NetwrokTracker[uid]
				mu_arr[IndexNetwork].RUnlock()
				a.Network = netSum / float64(TIME_INTERVAL)

				// === Syscalls ===
				mu_arr[IndexSyscall].RLock()
				syscalls := SyscallTrackers[uid]
				mu_arr[IndexSyscall].RUnlock()
				a.Syscall = float64(syscalls)/ TIME_INTERVAL

				// === CPU ===
				mu_arr[IndexCPU].RLock()
				if s, ok := CpuTrackers[uid]; ok {
					a.CPU = s.CPUUsage // assuming you've stored the computed value here
					
				}
				mu_arr[IndexCPU].RUnlock()

				// === Memory ===
				mu_arr[IndexMemory].RLock()
				if s, ok := MemoryTrackers[uid]; ok {
					a.Memory = s.MemoryUsage // same assumption
				}
				mu_arr[IndexMemory].RUnlock()

				// === DiskIO ===
				mu_arr[IndexDiskIO].RLock()
				if s, ok := DiskTrackers[uid]; ok {
					a.DiskIO = s.DiskIOUsage // again, assuming you've precomputed and stored it
				}
				mu_arr[IndexDiskIO].RUnlock()

				anomaly_log := *a
				anomaly_log.Timestamp = time.Now()


				// === Send Anomaly Log ===
				logCh <- logs.Producer_msg{
					Body: anomaly_log.Encode(),
					Id:   2,
				}
			}
			mu_arr[IndexAnomaly].RUnlock()

			// === Reset All Trackers and Container Map ===
			for i:=0 ; i < 6 ; i++{
				mu_arr[i].Lock()
			}
			Container_uid_map = make(map[string]*logs.Anomaly_log)
			CpuTrackers = make(map[string]logs.CpuTracker)
			DiskTrackers = make(map[string]logs.DiskTracker)
			MemoryTrackers = make(map[string]logs.MemoryTracker)
			SyscallTrackers = make(map[string]int)
			NetwrokTracker = make(map[string]float64)

			for i:=0 ; i < 6 ; i++{
				mu_arr[i].Unlock()
			}

		}
	}
}





func GetCPUTracker(uid string) (logs.CpuTracker, bool) {
	mu_arr[IndexCPU].RLock()
	defer mu_arr[IndexCPU].RUnlock()
	val, ok := CpuTrackers[uid]
	return val, ok
}

func GetMemoryTracker(uid string) (logs.MemoryTracker, bool) {
	mu_arr[IndexMemory].RLock()
	defer mu_arr[IndexMemory].RUnlock()
	val, ok := MemoryTrackers[uid]
	return val, ok
}

func GetDiskTracker(uid string) (logs.DiskTracker, bool) {
	mu_arr[IndexDiskIO].RLock()
	defer mu_arr[IndexDiskIO].RUnlock()
	val, ok := DiskTrackers[uid]
	return val, ok
}

func GetNetworkUsage(uid string) (float64, bool) {
	mu_arr[IndexNetwork].RLock()
	defer mu_arr[IndexNetwork].RUnlock()
	val, ok := NetwrokTracker[uid]
	return val, ok
}

func GetSyscallCount(uid string) (int, bool) {
	mu_arr[IndexSyscall].RLock()
	defer mu_arr[IndexSyscall].RUnlock()
	val, ok := SyscallTrackers[uid]
	return val, ok
}


func SetContainerCPU(uid string, value float64) {
	mu_arr[IndexAnomaly].Lock()
	defer mu_arr[IndexAnomaly].Unlock()
	if log, ok := Container_uid_map[uid]; ok {
		log.CPU = value
	}
}

func SetContainerMemory(uid string, value float64) {
	mu_arr[IndexAnomaly].Lock()
	defer mu_arr[IndexAnomaly].Unlock()
	if log, ok := Container_uid_map[uid]; ok {
		log.Memory = value
	}
}

func SetContainerDiskIO(uid string, value float64) {
	mu_arr[IndexAnomaly].Lock()
	defer mu_arr[IndexAnomaly].Unlock()
	if log, ok := Container_uid_map[uid]; ok {
		log.DiskIO = value
	}
}
