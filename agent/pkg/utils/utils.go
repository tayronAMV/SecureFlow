package utils

import (
	"time"
	"agent/pkg/logs"
)

const TIME_INTERVAL = float64(10.0)

var Container_uid_map = make(map[string]*logs.Anomaly_log)

var CpuTrackers = make(map[string]logs.CpuTracker)
var DiskTrackers = make(map[string]logs.DiskTracker)
var MemoryTrackers = make(map[string]logs.MemoryTracker)

func Send_to_Server_Reset(){
	for _ , ptr := range Container_uid_map{
		count_syscalls  := ptr.Syscall
		total_network := ptr.Network


		ptr.Syscall = count_syscalls / TIME_INTERVAL // avg over interval , syscall per second for this container
		ptr.Network = total_network / TIME_INTERVAL
		ptr.Timestamp = time.Now()
		
		logs.Producer(logs.Producer_msg{
			Body: *ptr,
			Id : 0, 
		})
	}
	
	// need to make it concurent 

	Container_uid_map = make(map[string]*logs.Anomaly_log)
	CpuTrackers = make(map[string]logs.CpuTracker)
	DiskTrackers = make(map[string]logs.DiskTracker)
	MemoryTrackers = make(map[string]logs.MemoryTracker)
	

}