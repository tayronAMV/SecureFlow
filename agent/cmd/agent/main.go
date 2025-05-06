package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"agent/internal/agent"
)

func main() {
	log.Println("🚀 Starting SecureFlow agent...")
	
	agent.Start()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	agent.Stop()
	log.Println("🛑 Agent stopped.")
}