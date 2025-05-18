package internal

import (
	"bytes"
	"encoding/binary"
	"log"
	"time"
	"agent/pkg/kube"
	"agent/pkg/logs"
	"agent/pkg/utils"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
)

var syscallReader *ringbuf.Reader
var syscallLinks []link.Link


var syscallEventsMap *ebpf.Map

func InitSyscallMonitor() {
	spec, err := ebpf.LoadCollectionSpec("bpf/syscalls.bpf.o")
	if err != nil {
		log.Printf("❌ Failed to load syscall BPF spec: %v", err)
		return
	}

	objs := struct {
		LogExecve     *ebpf.Program `ebpf:"log_execve"`
		LogExecveat   *ebpf.Program `ebpf:"log_execveat"`
		LogOpen       *ebpf.Program `ebpf:"log_open"`
		LogUnlink     *ebpf.Program `ebpf:"log_unlink"`
		LogChmod      *ebpf.Program `ebpf:"log_chmod"`
		LogMount      *ebpf.Program `ebpf:"log_mount"`
		LogSetuid     *ebpf.Program `ebpf:"log_setuid"`
		LogSocket     *ebpf.Program `ebpf:"log_socket"`
		LogConnect    *ebpf.Program `ebpf:"log_connect"`
		SyscallEvents *ebpf.Map     `ebpf:"syscall_events"`
	}{}

	if err := spec.LoadAndAssign(&objs, nil); err != nil {
		log.Printf("❌ Failed to assign syscall BPF programs: %v", err)
		return
	}

	syscallEventsMap = objs.SyscallEvents

	// Attach tracepoints
	traceAttach := func(tp string, prog *ebpf.Program) {
		lnk, err := link.Tracepoint("syscalls", tp, prog, nil)
		if err != nil {
			log.Printf("⚠️ Failed to attach tracepoint %s: %v", tp, err)
		} else {
			log.Printf("🔗 Attached tracepoint: %s", tp)
			syscallLinks = append(syscallLinks, lnk)
		}
	}

	traceAttach("sys_enter_execve", objs.LogExecve)
	traceAttach("sys_enter_openat", objs.LogOpen)
	traceAttach("sys_enter_unlinkat", objs.LogUnlink)
	traceAttach("sys_enter_execveat", objs.LogExecveat)
	traceAttach("sys_enter_chmod", objs.LogChmod)
	traceAttach("sys_enter_mount", objs.LogMount)
	traceAttach("sys_enter_setuid", objs.LogSetuid)
	traceAttach("sys_enter_socket", objs.LogSocket)
	traceAttach("sys_enter_connect", objs.LogConnect)
}

func StartSyscallReader() {
	var err error
	syscallReader, err = ringbuf.NewReader(syscallEventsMap)
	if err != nil {
		log.Printf("❌ Failed to open syscall ring buffer: %v", err)
		return
	}

	go func() {
		log.Println("🟢 Syscall monitoring active.")
		for {
			record, err := syscallReader.Read()
			if err != nil {
				log.Printf("⚠️ Syscall ringbuf error: %v", err)
				break
			}
			var raw logs.RawSyscallEvent
			if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &raw); err != nil {
				log.Printf("❌ Failed to decode syscall event: %v", err)
				continue
			}

			event := logs.SyscallEvent{
				Pid: raw.Pid,
				Type: raw.Type,
				Comm: raw.Comm,
				Filename: raw.Filename,
			}
			if _, ok := kube.Pid_toContainer_Map[int(event.Pid)]; !ok{
				continue
			}
			//insert syscall count into Anomaly_Log calculation
			UID := kube.PidToUid(int(event.Pid))
			if obj , ok := utils.Container_uid_map[UID] ; ok {
				obj.Syscall +=1
			}else{
				utils.Container_uid_map[UID] = &logs.Anomaly_log{}
			}
			event.UID = UID 
			event.Timestamp = time.Now()
			// push to the server 
			logs.Producer(logs.Producer_msg{
				Body: event,
				Id : 5 , 
			})

		}	
	}()
}

func StopSyscallMonitor() {
	if syscallReader != nil {
		syscallReader.Close()
	}
	for _, l := range syscallLinks {
		l.Close()
	}
}
