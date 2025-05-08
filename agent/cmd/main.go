package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"agent/internal"
	"agent/pkg/logs"
	"agent/pkg/kube"

)

func agent_Start(){
	internal.Traffic_INIT()
	internal.InitSyscallMonitor()
	
	for {
		mappings, err := kube.FetchContainerMappings()
		if err != nil {
			log.Printf("❌ Failed to fetch container mappings: %v", err)
			return
		}

		internal.Attach_bpf_network(mappings)


		internal.StartSyscallReader()
		internal.StartTrraficCollector()
		internal.StartResourceCollector(mappings)
	}
}


func agent_stop(){
	internal.StopSyscallMonitor()
	internal.Traffic_close()
}

func main() {
	log.Println("🚀 Starting SecureFlow agent...")


	

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	agent_stop()
	
	log.Println("🛑 Agent stopped.")
}