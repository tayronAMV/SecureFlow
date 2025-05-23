package logs

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	ProducerConn    *amqp.Connection
	ProducerChannel *amqp.Channel
	ProducerQueue   amqp.Queue
)



// wtf is this ? 
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

func Producer(evt Producer_msg) {
	
	fmt.Println(evt.Body)
	

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

	

}


func RabbitMQ_producer_Close() {
	if ProducerChannel != nil {
		ProducerChannel.Close()
	}
	if ProducerConn != nil {
		ProducerConn.Close()
	}
}


func (m MemoryUsage) String() string {
	return fmt.Sprintf(
		"Memory Usage [UID: %s]\n"+
			"Timestamp:        %s\n"+
			"Used Memory:      %d bytes\n"+
			"Memory Limit:     %d bytes\n"+
			"RSS:              %d bytes\n"+
			"Cache Memory:     %d bytes\n"+
			"Memory Usage Rate: %.2f%%\n",
		m.UID,
		m.Timestamp.Format(time.RFC3339),
		m.UsedMemory,
		m.MemoryLimit,
		m.RSS,
		m.CacheMemory,
		m.MemoryUsageRate*100,
	)
}

func (d DiskIOUsage) String() string {
	return fmt.Sprintf(
		"Disk I/O Usage [UID: %s]\n"+
			"Timestamp:         %s\n"+
			"Disk Read Bytes:   %d\n"+
			"Disk Write Bytes:  %d\n"+
			"Disk Usage Rate:   %.2f%%\n",
		d.UID,
		d.Timestamp.Format(time.RFC3339),
		d.DiskReadBytes,
		d.DiskWriteBytes,
		d.DiskUsageRate*100,
	)
}

func (c CPUUsage) String() string {
	return fmt.Sprintf(
		"CPU Usage [UID: %s]\n"+
			"Timestamp:        %s\n"+
			"CPU Time:         %d ns\n"+
			"CPU Usage Rate:   %.2f%%\n"+
			"CPU Limit:        %d units\n",
		c.UID,
		c.Timestamp.Format(time.RFC3339),
		c.CPUTime,
		c.CPUUsageRate*100,
		c.CPULimit,
	)
}