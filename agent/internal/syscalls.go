package internal

import (
	"bytes"
	"encoding/binary"
	"log"

	"agent/pkg/kube"
	"agent/pkg/logs"
	"fmt"
	"os"
	"os/signal"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
)


func StartSyscallReader() {
	spec, err := ebpf.LoadCollectionSpec("bpf/syscalls.bpf.o")
	if err != nil {
		log.Fatalf("❌ Failed to load BPF spec: %v", err)
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
		log.Fatalf("❌ Failed to assign BPF programs: %v", err)
	}
	defer objs.SyscallEvents.Close()

	// Attach tracepoints
	links := []link.Link{}
	attach := func(tp string, prog *ebpf.Program) {
		lnk, err := link.Tracepoint("syscalls", tp, prog, nil)
		if err != nil {
			log.Printf("⚠️ Failed to attach %s: %v", tp, err)
		} else {
			log.Printf("🔗 Attached %s", tp)
			links = append(links, lnk)
		}
	}

	attach("sys_enter_execve", objs.LogExecve)
	attach("sys_enter_execveat", objs.LogExecveat)
	attach("sys_enter_openat", objs.LogOpen)
	attach("sys_enter_unlinkat", objs.LogUnlink)
	attach("sys_enter_chmod", objs.LogChmod)
	attach("sys_enter_mount", objs.LogMount)
	attach("sys_enter_setuid", objs.LogSetuid)
	attach("sys_enter_socket", objs.LogSocket)
	attach("sys_enter_connect", objs.LogConnect)

	rd, err := ringbuf.NewReader(objs.SyscallEvents)
	if err != nil {
		log.Fatalf("❌ Failed to open ring buffer: %v", err)
	}
	defer rd.Close()

	log.Println("🟢 Syscall monitor running...")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		for {
			record, err := rd.Read()
			if err != nil {
				log.Printf("⚠️ Ringbuf read error: %v", err)
				return
			}

			var event logs.RawSyscallEvent
			if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &event); err != nil {
				log.Printf("❌ Decode error: %v", err)
				continue
			}
			
			if info , ok := kube.Cgroup_mapping[event.Cgid] ; ok {
				fmt.Printf("im gonna cum , this is the syscall %s \n and this is from conatiner %s\n",event.String(),info.PodName)
			}else{
				continue
			}

			// uid := utils.PidToUid(int(event.Pid))
			// event.UID = uid

			// if obj, ok := utils.Container_uid_map[uid]; ok {
			// 	obj.Syscall++
			// } else {
			// 	utils.Container_uid_map[uid] = &logs.Anomaly_log{Syscall: 1}
			// }

			
		}
	}()

	<-stop
	log.Println("👋 Stopping...")

	for _, l := range links {
		_ = l.Close()
	}
}

