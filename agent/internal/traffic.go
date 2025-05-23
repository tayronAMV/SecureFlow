package internal

import (
	
	"agent/pkg/logs"
	
	"bytes"
	"encoding/binary"
	
	"log"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/ringbuf"
	"context"
	"os"
	"os/signal"
	"syscall"
	"fmt"
)




func StartTrraficCollector() {
	// Load eBPF program
	spec, err := ebpf.LoadCollectionSpec("bpf/traffic.bpf.o")
	if err != nil {
		log.Fatalf("❌ LoadCollectionSpec: %v", err)
	}
	
	objs := struct {
		TcIngress *ebpf.Program `ebpf:"tc_ingress"`
		TcEgress  *ebpf.Program `ebpf:"tc_egress"`
		Events    *ebpf.Map     `ebpf:"events"`
	}{}
	
	if err := spec.LoadAndAssign(&objs, nil); err != nil {
		log.Fatalf("❌ LoadAndAssign: %v", err)
	}
	defer objs.TcIngress.Close()
	defer objs.TcEgress.Close()
	defer objs.Events.Close()

	// Initialize link tracker
	tracker := NewLinkTracker()
	defer tracker.CloseAll()

	// Attach to containers
	if err := attachToContainers(&objs, tracker); err != nil {
		log.Fatalf("❌ Failed to attach to containers: %v", err)
	}

	// Setup ringbuf reader
	rd, err := ringbuf.NewReader(objs.Events)
	if err != nil {
		log.Fatalf("❌ Failed to open ringbuf: %v", err)
	}
	defer rd.Close()

	log.Println("📦 Listening to ring buffer...")

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Channel for re-scanning containers
	rescanTicker := time.NewTicker(30 * time.Second)
	defer rescanTicker.Stop()

	// Start reading from ringbuf in a goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				record, err := rd.Read()
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					log.Printf("❌ ringbuf read error: %v", err)
					time.Sleep(100 * time.Millisecond)
					continue
				}

				var event logs.RawFlowEvent
				if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &event); err != nil {
					log.Printf("❌ Failed to parse event: %v", err)
					continue
				}

				// Enhanced packet information display
				srcIP := ipToString(event.SrcIP)
				dstIP := ipToString(event.DstIP)
				proto := protocolToString(event.Protocol)
				dir := directionToString(event.Direction)
				dpi := dpiProtocolToString(event.DpiProtocol)

				fmt.Printf("📦 [%s] %s %s:%d -> %s:%d (%s) Len=%d", 
					dir, proto, srcIP, event.SrcPort, dstIP, event.DstPort, dpi, event.PayloadLen)

				// Add protocol-specific information
				if event.DpiProtocol == 1 && event.Method[0] != 0 { // HTTP
					method := nullTerminatedString(event.Method[:])
					path := nullTerminatedString(event.Path[:])
					fmt.Printf(" [%s %s]", method, path)
				} else if event.DpiProtocol == 2 && event.QueryName[0] != 0 { // DNS
					queryName := nullTerminatedString(event.QueryName[:])
					fmt.Printf(" [Query: %s Type: %d]", queryName, event.QueryType)
				} else if event.DpiProtocol == 3 { // ICMP
					fmt.Printf(" [Type: %d]", event.IcmpType)
				}

				fmt.Printf(" @%d\n", event.Timestamp)
			}
		}
	}()

	// Main event loop
	for {
		select {
		case <-stop:
			log.Println("👋 Received signal, cleaning up...")
			goto cleanup

		case <-rescanTicker.C:
			log.Println("🔄 Rescanning for new containers...")
			if err := attachToContainers(&objs, tracker); err != nil {
				log.Printf("❌ Failed to rescan containers: %v", err)
			}
		}
	}

cleanup:
	// Cancel context to stop ringbuf reader
	cancel()
	
	// Close all links (defer will also handle this)
	tracker.CloseAll()
	rd.Close()
	
	log.Println("✅ Cleanup complete")
}



func ipToString(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip), byte(ip>>8), byte(ip>>16), byte(ip>>24))
}

func protocolToString(proto uint8) string {
	switch proto {
	case 1:
		return "ICMP"
	case 6:
		return "TCP"
	case 17:
		return "UDP"
	default:
		return fmt.Sprintf("Proto_%d", proto)
	}
}

func dpiProtocolToString(dpi uint8) string {
	switch dpi {
	case 0:
		return "Unknown"
	case 1:
		return "HTTP"
	case 2:
		return "DNS"
	case 3:
		return "ICMP"
	default:
		return fmt.Sprintf("DPI_%d", dpi)
	}
}

func directionToString(dir uint8) string {
	if dir == 0 {
		return "Egress"
	}
	return "Ingress"
}

func nullTerminatedString(b []byte) string {
	for i, v := range b {
		if v == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}