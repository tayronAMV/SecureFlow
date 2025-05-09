package logs

import (
	"encoding/json"
	"fmt"
	"log"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
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


var MongoClient *mongo.Client

var (
	SyscallCollection *mongo.Collection
	CPUCollection      *mongo.Collection
	MemoryCollection   *mongo.Collection
	DiskCollection     *mongo.Collection
	TrafficCollection  *mongo.Collection
)

func InitMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
	}

	MongoClient = client
	db := client.Database("secureflow")

	SyscallCollection = db.Collection("syscalls")
	CPUCollection = db.Collection("cpu")
	MemoryCollection = db.Collection("memory")
	DiskCollection = db.Collection("disk_io")
	TrafficCollection = db.Collection("traffic")

	log.Println("✅ Connected to MongoDB and initialized collections")
}

func SaveCPUUsage(data *CPUUsage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := CPUCollection.InsertOne(ctx, data)
	return err
}

func SaveMemoryUsage(data *MemoryUsage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := MemoryCollection.InsertOne(ctx, data)
	return err
}

func SaveDiskIOUsage(data *DiskIOUsage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := DiskCollection.InsertOne(ctx, data)
	return err
}

func SaveFlowEvent(evt *FlowEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := TrafficCollection.InsertOne(ctx, evt)
	return err
}

func SaveSyscall(evt *SyscallEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := SyscallCollection.InsertOne(ctx, evt)
	return err
}
