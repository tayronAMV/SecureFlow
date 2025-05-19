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
	// if evt.Id == 4 { 
	// 	fmt.Println(evt.Body)
	// }

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


