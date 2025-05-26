package main

import (
	"agent/internal"
	"agent/pkg/kube"
	"agent/pkg/logs"
	"agent/pkg/utils"

	// "agent/pkg/logs"
	"log"
	"os"
	"os/signal"
	"syscall"
	// "agent/pkg/utils"
	// "time"
)

func agent_Start(){
	log.Println("🚀 Starting SecureFlow agent...")
	logCh := make(chan logs.Producer_msg,100)
	NetworkCh := make(chan []logs.FlowRule,20)
	SyscallCh := make(chan []logs.SyscallEventRule,100)
	MemoryCh := make(chan []logs.MemoryUsageRule,100)
	DiskcCh := make(chan []logs.DiskIOUsageRule,100)
	CPUCh := make(chan []logs.CPUUsageRule,100)
	logs.StartProducer(logCh)
	logs.RabbitMQ_Consumer_Start(NetworkCh , SyscallCh , MemoryCh, DiskcCh , CPUCh )
	go kube.MappingTracker() 
	go internal.StartSyscallReader(logCh , SyscallCh) 
	go internal.StartResourceCollector(logCh , MemoryCh ,DiskcCh , CPUCh)  
	go internal.StartTrraficCollector(logCh , NetworkCh) 
	go utils.Anomaly_log_generator(logCh)
}


func main() {
	
	agent_Start()

	// sigs := make(chan os.Signal, 1)
	// signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// <-sigs

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop // wait for termination signal
	
	log.Println("🛑 Shutting down SecureFlow agent...")
	
	
	// log.Println("🛑 Agent stopped.")
}