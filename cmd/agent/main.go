package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"

	"agent/pkg/kube"
)

type FlowEvent struct {
	SrcIP    uint32
	DstIP    uint32
	SrcPort  uint16
	DstPort  uint16
	Protocol uint8
	PktLen   uint64
	_        [3]byte
}

func ipToString(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip),
		byte(ip>>8),
		byte(ip>>16),
		byte(ip>>24))
}

// Thread-safe structure to track attached cgroups
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

func resolveCgroupFromPID(pid int) string {
	cgroupFile := fmt.Sprintf("/proc/%d/cgroup", pid)
	content, err := os.ReadFile(cgroupFile)
	if err != nil {
		log.Printf("⚠️ Could not read cgroup file for PID %d: %v", pid, err)
		return ""
	}

	for _, line := range strings.Split(string(content), "\n") {
		fields := strings.SplitN(line, ":", 3)
		if len(fields) == 3 && strings.Contains(fields[2], "kubepods") {
			path := filepath.Join("/sys/fs/cgroup", fields[2])
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}
	log.Printf("⚠️ PID %d has no valid kubepods cgroup", pid)
	return ""
}

func main() {
	log.Println("🚀 Starting traffic monitor agent...")

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

	watcher := &CgroupWatcher{
		attached: make(map[string]link.Link),
	}

	// Background PID → cgroup → attach mapping loop
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			log.Println("🔄 Syncing container PIDs and attaching BPF...")
			mappings, err := kube.FetchContainerMappings()
			if err != nil {
				log.Printf("❌ Failed to fetch container mappings: %v", err)
				continue
			}

			for _, m := range mappings {
				log.Printf("🧩 Resolving cgroup for container %s (pod %s/%s)", m.ContainerID, m.Namespace, m.PodName)
				cgroupPath := resolveCgroupFromPID(m.PID)
				if cgroupPath != "" {
					watcher.AttachIfNeeded(cgroupPath, objs.MonitorIngress, ebpf.AttachCGroupInetIngress)
					watcher.AttachIfNeeded(cgroupPath, objs.MonitorEgress, ebpf.AttachCGroupInetEgress)
				}
			}
		}
	}()

	// Ring buffer reader
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

	// Main loop to consume events
	for {
		record, err := reader.Read()
		if err != nil {
			log.Printf("⚠️ Error reading from ring buffer: %v", err)
			break
		}

		var event FlowEvent
		err = binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &event)
		if err != nil {
			log.Printf("❌ Failed to decode event: %v", err)
			continue
		}

		log.Printf("📡 %s:%d → %s:%d | proto: %d | size: %d bytes",
			ipToString(event.SrcIP), event.SrcPort,
			ipToString(event.DstIP), event.DstPort,
			event.Protocol, event.PktLen)
	}
}
