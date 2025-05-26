package logs

import (
	"agent/pkg/logs"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)


var (
	ProducerConn    *amqp.Connection
	ProducerChannel *amqp.Channel
	ProducerQueue   amqp.Queue

	ConsumerConn    *amqp.Connection
	ConsumerChannel *amqp.Channel
	ConsumerQueue   amqp.Queue
)

func StartProducer(logCh <-chan Producer_msg) {
	RabbitMQ_producer_Start()
	go Producer(logCh)
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


	

func Producer(logCh <-chan Producer_msg) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-stop:
			RabbitMQ_producer_Close()
			return

		case msg, ok := <-logCh:
			if !ok {
				log.Println("🚪 logCh closed, shutting down producer")
				RabbitMQ_producer_Close()
				return
			}
			send_to_server(msg.Body, msg.Id)
		}
	}
}

func send_to_server(msg []byte , id int ){
	err := ProducerChannel.Publish(
		"", ProducerQueue.Name, false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msg,
			Headers: amqp.Table{
			"id": id,
			},
		},
	)

	if err != nil {
		log.Printf("Failed to publish message: %v", err)
		return
	}
	if id == 2{
		log.Println(DecodeAnomalyLog(msg))
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

func RabbitMQ_Consumer_Close() {
	if ConsumerChannel != nil {
		ConsumerChannel.Close()
	}
	if ConsumerConn != nil {
		ConsumerConn.Close()
	}
	
}

func RabbitMQ_Consumer_Start(
    NetworkCh <-chan []logs.FlowRule,
    MemoryCh <-chan []logs.MemoryUsageRule,
    DiskCh   <-chan []logs.DiskIOUsageRule,
    CPUCh    <-chan []logs.CPUUsageRule,
){
	var err error
	
	ConsumerConn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("❌ Failed to connect to RabbitMQ: %v", err)
	}

	ConsumerChannel, err = ConsumerConn.Channel()
	if err != nil {
		log.Fatalf("❌ Failed to open a channel: %v", err)
	}

	q, err := ConsumerChannel.QueueDeclare(
		"ConsumerQueue",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("❌ Failed to declare queue: %v", err)
	}

	Consumer_msgs, err := ConsumerChannel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("❌ Failed to register consumer: %v", err)
	}



	go func(){
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		for {
			select{
			case <- stop:
				RabbitMQ_Consumer_Close()
			case <-Consumer_msgs:
				for msg := range Consumer_msgs{
					val, ok := d.Headers["arg"]
					if !ok {
						log.Println("⚠️ 'arg' header missing")
						continue
					}

					var arg int
					switch v := val.(type) {
					case int32:
						arg = int(v)
					case int64:
						arg = int(v)
					case int:
						arg = v
					default:
						log.Printf("❌ Unsupported 'arg' type: %T\n", v)
						continue
					}

					switch arg {
					case 1 : // network 
						
					case 2 : // syscall 
					
					case 3 : // resource 

					}
				} 
			}
		}
	}()

}


