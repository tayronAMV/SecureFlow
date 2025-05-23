package main

import (
	"agent/internal"
	"agent/pkg/kube"
	"agent/pkg/logs"
	"log"
	"os"
	"os/signal"
	"syscall"

	// "agent/pkg/utils"
	"time"

	
)

func agent_Start(){
	
	internal.InitSyscallMonitor()
	_, err := kube.FetchContainerMappings()
	if err != nil {
		log.Printf("❌ Failed to fetch container mappings: %v", err)
		return
	}
		
	// internal.StartResourceCollector(mappings)
	logs.RabbitMQ_producer_Start()
	defer logs.RabbitMQ_producer_Close()
	// internal.StartTrraficCollector()
	for {


		// need to make this councurrent

		// internal.StartSyscallReader()
	
		

		time.Sleep(120 * time.Second)
		// utils.Send_to_Server_Reset()

		





	}
}


func agent_stop(){
	// internal.StopSyscallMonitor()

	logs.RabbitMQ_producer_Close()
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