package db

import (
	"context"
	"time"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

)

var mongoClient *mongo.Client
var (
	anomalyLogCollection    *mongo.Collection
	cpuUsageCollection      *mongo.Collection
	memoryUsageCollection   *mongo.Collection
	diskIOUsageCollection   *mongo.Collection
	flowEventCollection     *mongo.Collection
	syscallEventCollection  *mongo.Collection
)


func InitMongo() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return err
	}

	mongoClient = client
	memoryUsageCollection = client.Database("secureflow").Collection("MemoryUsage")
	cpuUsageCollection = client.Database("secureflow").Collection("CPUUsage")
	diskIOUsageCollection = client.Database("secureflow").Collection("DiskIOUsage")
	syscallEventCollection = client.Database("secureflow").Collection("SyscallEvents")
	flowEventCollection = client.Database("secureflow").Collection("FlowEvents")
	anomalyLogCollection = client.Database("secureflow").Collection("AnomalyLogs")
	return nil
}


func InsertLog(log interface{}, id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch id {
	case 0: // AnomalyLog
		_, err := anomalyLogCollection.InsertOne(ctx, log)
		return err

	case 1: // CPUUsage
		_, err := cpuUsageCollection.InsertOne(ctx, log)
		return err

	case 2: // MemoryUsage
		_, err := memoryUsageCollection.InsertOne(ctx, log)
		return err

	case 3: // DiskIOUsage
		_, err := diskIOUsageCollection.InsertOne(ctx, log)
		return err

	case 4: // FlowEvent (assuming "NETWORK")
		_, err := flowEventCollection.InsertOne(ctx, log)
		return err

	case 5: // SyscallEvent
		_, err := syscallEventCollection.InsertOne(ctx, log)
		return err

	default:
		return fmt.Errorf("invalid log ID: %d", id)
	}
}
