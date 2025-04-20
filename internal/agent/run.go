package agent

import (
	"bytes"
	"encoding/binary"
	"log"
	"time"

	
	"agent/pkg/kube"
	"agent/pkg/logs"
	"agent/pkg/metrics"
	"agent/pkg/utils"

	"github.com/cilium/ebpf"
	
	"github.com/cilium/ebpf/ringbuf"
)

var (
	cw     *CgroupWatcher
	reader *ringbuf.Reader
	objs   struct {
		MonitorIngress *ebpf.Program `ebpf:"monitor_ingress"`
		MonitorEgress  *ebpf.Program `ebpf:"monitor_egress"`
		Events         *ebpf.Map     `ebpf:"events"`
	}
)

func Start() {
	spec, err := ebpf.LoadCollectionSpec("bpf/traffic.bpf.o")
	if err != nil {
		log.Fatalf("❌ Failed to load BPF spec: %v", err)
	}

	if err := spec.LoadAndAssign(&objs, nil); err != nil {
		log.Fatalf("❌ Failed to assign BPF programs: %v", err)
	}

	cw = New()
	startAttacher()
	startResourceCollector()
	startEventReader()
}

func Stop() {
	if reader != nil {
		reader.Close()
	}
	if cw != nil {
		cw.Cleanup()
	}
}

func startAttacher() {
	go func() {
		ticker := time.NewTicker(120 * time.Second)
		defer ticker.Stop()

		attach := func() {
			log.Println("🔄 Syncing container PIDs and attaching BPF...")
			mappings, err := kube.FetchContainerMappings()
			if err != nil {
				log.Printf("❌ Failed to fetch container mappings: %v", err)
				return
			}
			for _, m := range mappings {
				cgPath, err := utils.ResolvePathForPID(m.PID)
				if err == nil {
					cw.AttachIfNeeded(cgPath, objs.MonitorIngress, ebpf.AttachCGroupInetIngress)
					cw.AttachIfNeeded(cgPath, objs.MonitorEgress, ebpf.AttachCGroupInetEgress)
				}
			}
		}

		attach()
		for range ticker.C {
			attach()
		}
	}()
}

func startResourceCollector() {
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		collect := func() {
			mappings, err := kube.FetchContainerMappings()
			if err != nil {
				log.Printf("⚠️ Failed to fetch container mappings: %v", err)
				return
			}

			for _, m := range mappings {
				if mem, err := metrics.GetMemoryUsage(m.ContainerID, m.PID); err == nil {
					logs.LogMemory(mem)
				}
				if cpu, err := metrics.GetCPUUsage(m.ContainerID, m.PID); err == nil {
					logs.LogCPU(cpu)
				}
				if disk, err := metrics.GetDiskIOUsage(m.ContainerID, m.PID); err == nil {
					logs.LogDisk(disk)
				}
			}
		}

		collect()
		for range ticker.C {
			collect()
		}
	}()
}

func startEventReader() {
	var err error
	reader, err = ringbuf.NewReader(objs.Events)
	if err != nil {
		log.Fatalf("❌ Failed to open ring buffer: %v", err)
	}

	go func() {
		log.Println("🟢 Agent running: monitoring all pod traffic via cgroup/skb...")
		for {
			record, err := reader.Read()
			if err != nil {
				log.Printf("⚠️ Error reading from ring buffer: %v", err)
				break
			}

			var event logs.TrafficInfo
			if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &event); err != nil {
				log.Printf("❌ Failed to decode event: %v", err)
				continue
			}
			logs.LogTraffic(&event)
		}
	}()
}
