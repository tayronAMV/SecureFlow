package internal

import (
	"agent/pkg/kube"
	"agent/pkg/logs"
	"agent/pkg/utils"
	"bytes"
	"encoding/binary"
	"net"
	"log"
	"time"
	"os/exec"
	"strconv"
	"strings"
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
	if objs.MonitorIngress == nil {
		log.Println("❌ MonitorIngress is nil")
	} else {
		fd :=  objs.MonitorIngress.FD()
		log.Printf("✅ MonitorIngress loaded with FD = %d", fd)
		
	}

	if objs.MonitorEgress == nil {
		log.Println("❌ MonitorEgress is nil")
	} else {
		fd := objs.MonitorEgress.FD()
		log.Printf("✅ MonitorEgress loaded with FD = %d", fd)
	}

	cw = New()
}

func Attach_bpf_network(mappings []kube.ContainerMapping) {
	log.Println("🔄 Syncing container PIDs and attaching BPF trrafic code ")
	for _, m := range mappings {
		cgPath, err := ResolvePathForPID(m.PID)
		log.Println(cgPath , " plsssss ")
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
				Timestamp:     convertKtimeToTimestamp(raw.Timestamp),
				SrcIP:         uint32ToIP(raw.SrcIP),
				DstIP:        uint32ToIP(raw.DstIP),
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
				DNSQueryName:  raw.DNSQueryName,
				DNSQueryType:  raw.DNSQueryType,
				ICMPType:      raw.ICMPType,
				Pid:           raw.Pid,
		}

		
			UID := kube.PidToUid(int(event.Pid))

			//calculate Network usage for this container , for Anomaly_log
			if log, ok := utils.Container_uid_map[UID]; ok {
				log.Network += float64(event.PayloadLen)
			}else{
				utils.Container_uid_map[UID] = &logs.Anomaly_log{}
			}
			
			
			event.UID = UID 
			//send to the server 
			// if event.SrcIP== "0.0.0.0" && event.DstIP== "0.0.0.0"{
			// 	continue 
			// }	
			log.Printf(
				`{"timestamp":%v, "src_ip":%v, "dst_ip":%v, "src_port":%d, "dst_port":%d, "protocol":%d, "direction":%d, "payload_len":%d, "dpi_protocol":%d, "http_method":"%s", "http_path":"%s", "dns_query_name":"%s", "dns_query_type":%d, "icmp_type":%d, "pid":%d, "uid":"%s \n"}`,
				event.Timestamp,
				event.SrcIP,
				event.DstIP,
				event.SrcPort,
				event.DstPort,
				event.Protocol,
				event.Direction,
				event.PayloadLen,
				event.DPIProtocol,
				string(bytes.Trim(event.HTTPMethod[:], "\x00")),
				string(bytes.Trim(event.HTTPPath[:], "\x00")),
				string(bytes.Trim(event.DNSQueryName[:], "\x00")),
				event.DNSQueryType,
				event.ICMPType,
				event.Pid,
				event.UID,
			)
			logs.Producer(logs.Producer_msg{
				Body: event,
				Id : 4 , 
			})

			
		}
	}()
}



func uint32ToIP(ip uint32) string {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, ip)
	return net.IP(b).String()
}

func convertKtimeToTimestamp(ktimeNS uint64) (time.Time) {
	// Step 1: Get boot time (in seconds since epoch)
	out, err := exec.Command("cut", "-f1", "-d.", "/proc/uptime").Output()
	if err != nil {
		return time.Time{}
	}

	uptimeSec, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return time.Time{}
	}

	bootTime := time.Now().Add(-time.Duration(uptimeSec) * time.Second)

	// Step 2: Convert ktime to duration
	eventTime := bootTime.Add(time.Duration(ktimeNS) * time.Nanosecond)
	return eventTime
}