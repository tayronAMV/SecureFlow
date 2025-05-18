package internal

import (
	"agent/pkg/kube"
	"agent/pkg/logs"
	// "agent/pkg/utils"
	"bytes"
	"encoding/binary"
	
	"log"

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

			var raw logs.RawFlowEvent
			if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &raw); err != nil {
				log.Printf("❌ Failed to decode event: %v", err)
				continue
			}
			event := logs.FlowEvent{
				Timestamp:     raw.Timestamp,
				SrcIP:         raw.SrcIP,
				DstIP:         raw.DstIP,
				SrcPort:       raw.SrcPort,
				DstPort:       raw.DstPort,
				Protocol:      raw.Protocol,
				Direction:     raw.Direction,
				PayloadLen:    raw.PayloadLen,
				DPIProtocol:   raw.DPIProtocol,
				Reserved1:     raw.Reserved1,
				Reserved2:     raw.Reserved2,
				HTTPMethod:    raw.HTTPMethod,
				HTTPPath:      raw.HTTPPath,
				// DNSQueryName:  raw.DNSQueryName,
				// DNSQueryType:  raw.DNSQueryType,
				// ICMPType:      raw.ICMPType,
				Pid:           raw.Pid,
		}

			log.Println("this is packet = "  , event)
			// UID := kube.PidToUid(int(event.Pid))

			// //calculate Network usage for this container , for Anomaly_log
			// if log, ok := utils.Container_uid_map[UID]; ok {
			// 	log.Network += float64(event.PayloadLen)
			// }else{
			// 	utils.Container_uid_map[UID] = &logs.Anomaly_log{}
			// }
			// event.UID = UID 
			//send to the server 
			logs.Producer(logs.Producer_msg{
				Body: event,
				Id : 4 , 
			})

			
		}
	}()
}
