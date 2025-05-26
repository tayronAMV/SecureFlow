package internal

import (
	"agent/pkg/kube"
	"agent/pkg/logs"
	"agent/pkg/utils"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/ringbuf"
)




func StartTrraficCollector(logCh chan logs.Producer_msg , NetworkCh chan []logs.FlowRule) {
	// Load eBPF program
	spec, err := ebpf.LoadCollectionSpec("bpf/traffic.bpf.o")
	if err != nil {
		log.Fatalf("❌ LoadCollectionSpec: %v", err)
	}
	
	objs := struct {
		TcIngress *ebpf.Program `ebpf:"tc_ingress"`
		TcEgress  *ebpf.Program `ebpf:"tc_egress"`
		FlowRules *ebpf.Map `ebpf:"flow_rules"` 
		Events    *ebpf.Map     `ebpf:"events"`
	}{}
	
	if err := spec.LoadAndAssign(&objs, nil); err != nil {
		log.Fatalf("❌ LoadAndAssign: %v", err)
	}
	defer objs.TcIngress.Close()
	defer objs.TcEgress.Close()
	defer objs.Events.Close()
	defer objs.FlowRules.Close()

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


	
	// Start reading from ringbuf in a goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <- NetworkCh:
				err := LoadFlowRules( <-NetworkCh, objs.FlowRules)
				if err != nil{
					log.Println("some happend " ,err)
				}
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

				var event logs.FlowEvent
				if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &event); err != nil {
					log.Printf("❌ Failed to parse event: %v", err)
					continue
				}
				container := IfIndex_Mapper[int(event.IfIndex)]

				utils.Update_uid_Map(container.UID , container)
				utils.Update_network_Tracker(container.UID , float64(event.PayloadLen))


				logCh <- logs.Producer_msg{
					Body: logs.Encode_string(event.String()),
					Id: 1,
				}
			}
		}
	}()

	// Main event loop
	mappingCh := make(chan struct{}, 1)
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

	for {
		select {
		case <-stop:
			log.Println(" Received signal, cleaning up...")
			goto cleanup

		case <-mappingCh:
			log.Println(" Rescanning for new containers...")

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


// LoadFlowRules loads the given list of rules into the BPF map.
func LoadFlowRules(rules []logs.FlowRule, bpfMap *ebpf.Map) error {
    if len(rules) > 128 {
        return fmt.Errorf("too many rules: max is 128")
    }

    // Debug: Check map info
    info, err := bpfMap.Info()
    if err != nil {
        log.Printf("❌ Failed to get map info: %v", err)
    } else {
        log.Printf("🔍 Map info - KeySize: %d, ValueSize: %d, MaxEntries: %d", 
            info.KeySize, info.ValueSize, info.MaxEntries)
    }



    for i, rule := range rules {
        idx := uint32(i)
        
        // Debug the rule being inserted
        log.Printf("🔍 Inserting rule %d: %+v", i, rule)
        
        if err := bpfMap.Put(idx, rule); err != nil {
            log.Printf("❌ Failed to insert rule %d: %v", i, err)
            log.Printf("❌ Rule data: %+v", rule)
            return fmt.Errorf("failed to insert rule %d: %w", i, err)
        }
        
        // Verify the rule was actually inserted
        var retrieved logs.FlowRule
        if err := bpfMap.Lookup(idx, &retrieved); err != nil {
            log.Printf("❌ Failed to verify rule %d: %v", i, err)
        } else {
            log.Printf("✅ Verified rule %d: %+v", i, retrieved)
        }
    }

    log.Printf("✅ Successfully loaded %d flow rules into BPF map", len(rules))
    return nil
}



