package main

import (
	"agent/internal"
	"log"
	"os"
	"os/signal"
	"syscall"

	"agent/pkg/kube"
	"agent/pkg/utils"
	"time"
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

		// need to make this councurrent

		internal.StartSyscallReader()
		internal.StartTrraficCollector()
		internal.StartResourceCollector(mappings)

		time.Sleep(10 * time.Second)
		utils.Send_to_Server_Reset()

		





	}
}


func agent_stop(){
	internal.StopSyscallMonitor()
	internal.Traffic_close()
}

func main() {
	log.Println("🚀 Starting SecureFlow agent...")
	agent_Start()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	agent_stop()
	
	log.Println("🛑 Agent stopped.")
}