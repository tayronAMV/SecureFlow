package rabbitmq

import (
	"encoding/json"
	"log"
	"server/internal/db"
	"server/internal/db/models"

	// "server/internal/db"
	"github.com/streadway/amqp"
)

// Global handles
var (
	agentConn    *amqp.Connection
	agentChannel *amqp.Channel
)

// Connect_to_agent establishes RabbitMQ connection and starts consuming logs
func Connect_to_agent() error {
	var err error

	agentConn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf(" Failed to connect to RabbitMQ: %v", err)
	}

	agentChannel, err = agentConn.Channel()
	if err != nil {
		log.Fatalf(" Failed to open a channel: %v", err)
	}

	q, err := agentChannel.QueueDeclare(
		"agent_logs",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf(" Failed to declare queue: %v", err)
	}

	msgs, err := agentChannel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf(" Failed to register consumer: %v", err)
	}

	log.Println(" Connected. Waiting for anomaly logs...")

	//TODO , change for recieving any log from agent , types 0 - 5 
	go func() {
		for msg := range msgs {
			var generic_struct models.Consumer_msg
			err := json.Unmarshal(msg.Body,&generic_struct)
			if err != nil {
				log.Printf(" Invalid JSON: %v", err)
				continue
			}

			switch generic_struct.Id{
			case 1 :
				var s models.Consumer_activity_log
				err := json.Unmarshal(msg.Body , &s)
				if err != nil {					
					log.Printf(" Invalid JSON: %v", err)
					continue
				}
				db.InsertLog(s.Body)
			
			case 2 :
				var s models.Consumer_Anomaly_log
				err := json.Unmarshal(msg.Body , &0s)
				if err != nil {					
					log.Printf(" Invalid JSON: %v", err)
					continue
				}
				db.InsertAnomaly_Log(&s.Body)

		}}
	}()

	return nil
}

// CloseAgentConnection cleanly closes RabbitMQ connection and channel
func CloseAgentConnection() {
	if agentChannel != nil {
		agentChannel.Close()
	}
	if agentConn != nil {
		agentConn.Close()
	}
	log.Println(" RabbitMQ agent connection closed.")
}
