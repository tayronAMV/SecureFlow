package logs

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	ProducerConn    *amqp.Connection
	ProducerChannel *amqp.Channel
	ProducerQueue   amqp.Queue
)

// ANSI colors

const (
	colorReset  = "\033[0m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorPurple = "\033[35m"
)

// Traffic Event (Blue)
func LogTraffic(event *TrafficInfo) {
	log.Printf(colorBlue+"📡 %s:%d → %s:%d | proto: %d | size: %d bytes "+colorReset,
		ipToString(event.SrcIP), event.SrcPort,
		ipToString(event.DstIP), event.DstPort,
		event.Protocol, event.PktLen,
	)
}

// Memory Usage (Green)
func LogMemory(mem *MemoryUsage) {
	log.Printf(colorGreen+"🧠 [%s] Memory used: %d / %d (%.2f%%) | RSS: %d | Cache: %d"+colorReset,
		mem.ContainerID, mem.UsedMemory, mem.MemoryLimit,
		mem.MemoryUsageRate*100, mem.RSS, mem.CacheMemory,
	)
}

// CPU Usage (Yellow)
func LogCPU(cpu *CPUUsage) {
	log.Printf(colorYellow+"⚙️ [%s] CPU time: %d ms | Usage rate: %.2f%% | Limit: %d"+colorReset,
		cpu.ContainerID, cpu.CPUTime, cpu.CPUUsageRate*100, cpu.CPULimit,
	)
}

// Disk I/O Usage (Purple)
func LogDisk(disk *DiskIOUsage) {
	log.Printf(colorPurple+"💾 [%s] Disk R: %d B, W: %d B | Usage rate: %.2f B/s"+colorReset,
		disk.ContainerID, disk.DiskReadBytes, disk.DiskWriteBytes, disk.DiskUsageRate,
	)
}

// Utility
func ipToString(ip [4]byte) string {
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}



func LogSyscall(pid uint32, eventType uint32, comm string, filename string) {
	log.Printf("🔍 [PID %d] %s %s %s", pid, decodeSyscallType(eventType), comm, filename)
}

func decodeSyscallType(t uint32) string {
	switch t {
	case 1:
		return "execve"
	case 2:
		return "execveat"
	case 3:
		return "open"
	case 4:
		return "unlink"
	case 5:
		return "chmod"
	case 6:
		return "mount"
	case 7:
		return "setuid"
	case 8:
		return "socket"
	case 9:
		return "connect"
	default:
		return fmt.Sprintf("unknown(%d)", t)
	}
}


func RabbitMQ_producer_Start() {
	var err error

	ProducerConn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("❌ Failed to connect to RabbitMQ: %v", err)
	}

	ProducerChannel, err = ProducerConn.Channel()
	if err != nil {
		log.Fatalf("❌ Failed to open channel: %v", err)
	}

	ProducerQueue, err = ProducerChannel.QueueDeclare(
		"agent_logs",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("❌ Failed to declare queue: %v", err)
	}

	log.Println("✅ RabbitMQ Producer ready")
}



func Producer(evt Anomaly_log) {
	body, err := json.Marshal(evt)
	if err != nil {
		log.Printf("❌ JSON marshal failed: %v", err)
		return
	}

	err = ProducerChannel.Publish(
		"", ProducerQueue.Name, false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		log.Printf("❌ Failed to publish message: %v", err)
		return
	}

	log.Println("📤 Sent anomaly log to queue")
}


func RabbitMQ_producer_Close() {
	if ProducerChannel != nil {
		ProducerChannel.Close()
	}
	if ProducerConn != nil {
		ProducerConn.Close()
	}
}
