package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"agent/pkg/kube"
	"agent/pkg/logs"
	"agent/pkg/metrics"
	"agent/pkg/utils"
		

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
)

type CgroupWatcher struct {
	mu       sync.Mutex
	attached map[string]link.Link
}

func (cw *CgroupWatcher) AttachIfNeeded(path string, prog *ebpf.Program, attachType ebpf.AttachType) {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	key := path + fmt.Sprintf(":%d", attachType)
	if _, exists := cw.attached[key]; exists {
		return
	}

	cgLink, err := link.AttachCgroup(link.CgroupOptions{
		Path:    path,
		Attach:  attachType,
		Program: prog,
	})
	if err != nil {
		log.Printf("❌ Failed to attach %v to %s: %v", attachType, path, err)
		return
	}

	cw.attached[key] = cgLink
	log.Printf("✅ Successfully attached %v to %s", attachType, path)
}

func (cw *CgroupWatcher) Cleanup() {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	for key, l := range cw.attached {
		if err := l.Close(); err != nil {
			log.Printf("⚠️ Failed to close link %s: %v", key, err)
		} else {
			log.Printf("🧹 Cleaned up link: %s", key)
		}
	}
}

func main() {
	log.Println("🚀 Starting SecureFlow agent...")

	spec, err := ebpf.LoadCollectionSpec("bpf/traffic.bpf.o")
	if err != nil {
		log.Fatalf("❌ Failed to load BPF spec: %v", err)
	}

	objs := struct {
		MonitorIngress *ebpf.Program `ebpf:"monitor_ingress"`
		MonitorEgress  *ebpf.Program `ebpf:"monitor_egress"`
		Events         *ebpf.Map     `ebpf:"events"`
	}{}

	if err := spec.LoadAndAssign(&objs, nil); err != nil {
		log.Fatalf("❌ Failed to load and assign BPF programs: %v", err)
	}
	defer objs.MonitorIngress.Close()
	defer objs.MonitorEgress.Close()
	defer objs.Events.Close()

	watcher := &CgroupWatcher{attached: make(map[string]link.Link)}

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
				log.Printf("🔍 Mapping: %s/%s (container: %s)", m.Namespace, m.PodName, m.ContainerID)
				cgroupPath,err := utils.ResolvePathForPID(m.PID)
				if err == nil {
					watcher.AttachIfNeeded(cgroupPath, objs.MonitorIngress, ebpf.AttachCGroupInetIngress)
					watcher.AttachIfNeeded(cgroupPath, objs.MonitorEgress, ebpf.AttachCGroupInetEgress)
				}
			}
		}
	
		// Run immediately
		attach()
	
		// Then repeat on every tick
		for range ticker.C {
			attach()
		}
	}()
	

	// Async resource logging
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
				if mem, err := metrics.GetMemoryUsage(m.ContainerID,m.PID); err == nil {
					logs.LogMemory(mem)
				}
				if cpu, err := metrics.GetCPUUsage(m.ContainerID,m.PID); err == nil {
					logs.LogCPU(cpu)
				}
				if disk, err := metrics.GetDiskIOUsage(m.ContainerID,m.PID); err == nil {
					logs.LogDisk(disk)
				}
			}
		}
	
		// Run immediately
		collect()
	
		// Then run every tick
		for range ticker.C {
			collect()
		}
	}()
	

	reader, err := ringbuf.NewReader(objs.Events)
	if err != nil {
		log.Fatalf("❌ Failed to open ring buffer: %v", err)
	}
	defer reader.Close()

	log.Println("🟢 Agent running: monitoring all pod traffic via cgroup/skb...")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Println("🛑 Signal received, shutting down...")
		reader.Close()
		watcher.Cleanup()
		os.Exit(0)
	}()

	for {
		record, err := reader.Read()
		if err != nil {
			log.Printf("⚠️ Error reading from ring buffer: %v", err)
			break
		}

		var event logs.TrafficInfo
		err = binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &event)
		if err != nil {
			log.Printf("❌ Failed to decode event: %v", err)
			continue
		}

		logs.LogTraffic(&event)
	}
}
