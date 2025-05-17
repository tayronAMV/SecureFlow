package internal

import (
	"github.com/cilium/ebpf"
	"log"
	"bytes"
	"github.com/cilium/ebpf/ringbuf"
	"agent/pkg/kube"
	"encoding/binary"
	"agent/pkg/logs"
	"agent/pkg/utils"
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

func Traffic_INIT(){
	spec, err := ebpf.LoadCollectionSpec("bpf/traffic.bpf.o")
	if err != nil {
		log.Fatalf("❌ Failed to load BPF spec: %v", err)
	}

	if err := spec.LoadAndAssign(&objs, nil); err != nil {
		log.Fatalf("❌ Failed to assign BPF programs: %v", err)
	}

	cw = New()
}

func Attach_bpf_network(mappings []kube.ContainerMapping) {
	log.Println("🔄 Syncing container PIDs and attaching BPF trrafic code ")

	for _, m := range mappings {
		cgPath, err := ResolvePathForPID(m.PID)
		if err == nil {
			cw.AttachIfNeeded(cgPath, objs.MonitorIngress, ebpf.AttachCGroupInetIngress)
			cw.AttachIfNeeded(cgPath, objs.MonitorEgress, ebpf.AttachCGroupInetEgress)
		}
	}
}

func Traffic_close(){
	if reader != nil {
		reader.Close()
	}
	if cw != nil {
		cw.Cleanup()
	}
}


func StartTrraficCollector() {
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

			var event logs.FlowEvent
			if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &event); err != nil {
				log.Printf("❌ Failed to decode event: %v", err)
				continue
			}
			UID := kube.PidToUid(int(event.Pid))

			//calculate Network usage for this container , for Anomaly_log
			if log, ok := utils.Container_uid_map[UID]; ok {
				log.Network += float64(event.PayloadLen)
			}else{
				utils.Container_uid_map[UID] = &logs.Anomaly_log{}
			}

			//send to the server 
			logs.Producer(logs.Producer_msg{
				Body: event,
				Id : 4 , 
			})

			
		}
	}()
}
