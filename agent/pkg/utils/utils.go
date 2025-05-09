package utils

import (

	"agent/pkg/logs"
)


var Container_uid_map = make(map[string]*logs.Anomaly_log)

var CpuTrackers = make(map[string]logs.CpuTracker)
var DiskTrackers = make(map[string]logs.DiskTracker)
var MemoryTrackers = make(map[string]logs.MemoryTracker)

func Send_to_Server_Reset(){
	for _ , ptr := range Container_uid_map{
		logs.Producer(*ptr)
	}
	
	// need to make it concurent 

	Container_uid_map = make(map[string]*logs.Anomaly_log)
	CpuTrackers = make(map[string]logs.CpuTracker)
	DiskTrackers = make(map[string]logs.DiskTracker)
	MemoryTrackers = make(map[string]logs.MemoryTracker)

}