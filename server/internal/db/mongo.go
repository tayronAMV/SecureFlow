package db

import (
	"context"
	"fmt"
	"server/internal/db/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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



func FindLogs(anomaly models.AnomalyLog) (models.AnomalyContext, error) {
	filter := bson.M{
		"timestamp": bson.M{
			"$gte": anomaly.Timestamp.Add(-10 * time.Second),
			"$lte": anomaly.Timestamp.Add(10 * time.Second),
		},
		"UID": anomaly.UID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. CPU Usage
	var cpuLogs []models.CPUUsage
	cpuCursor, err := cpuUsageCollection.Find(ctx, filter)
	if err != nil {
		return models.AnomalyContext{}, fmt.Errorf("cpu query error: %w", err)
	}
	defer cpuCursor.Close(ctx)
	if err := cpuCursor.All(ctx, &cpuLogs); err != nil {
		return models.AnomalyContext{}, fmt.Errorf("cpu decode error: %w", err)
	}

	// 2. Memory Usage
	var memoryLogs []models.MemoryUsage
	memoryCursor, err := memoryUsageCollection.Find(ctx, filter)
	if err != nil {
		return models.AnomalyContext{}, fmt.Errorf("memory query error: %w", err)
	}
	defer memoryCursor.Close(ctx)
	if err := memoryCursor.All(ctx, &memoryLogs); err != nil {
		return models.AnomalyContext{}, fmt.Errorf("memory decode error: %w", err)
	}

	// 3. Disk I/O Usage
	var diskLogs []models.DiskIOUsage
	diskCursor, err := diskIOUsageCollection.Find(ctx, filter)
	if err != nil {
		return models.AnomalyContext{}, fmt.Errorf("disk query error: %w", err)
	}
	defer diskCursor.Close(ctx)
	if err := diskCursor.All(ctx, &diskLogs); err != nil {
		return models.AnomalyContext{}, fmt.Errorf("disk decode error: %w", err)
	}

	// 4. Flow Events (Network)
	var flowLogs []models.FlowEvent
	flowCursor, err := flowEventCollection.Find(ctx, filter)
	if err != nil {
		return models.AnomalyContext{}, fmt.Errorf("flow query error: %w", err)
	}
	defer flowCursor.Close(ctx)
	if err := flowCursor.All(ctx, &flowLogs); err != nil {
		return models.AnomalyContext{}, fmt.Errorf("flow decode error: %w", err)
	}

	// 5. Syscall Events
	var syscallLogs []models.SyscallEvent
	syscallCursor, err := syscallEventCollection.Find(ctx, filter)
	if err != nil {
		return models.AnomalyContext{}, fmt.Errorf("syscall query error: %w", err)
	}
	defer syscallCursor.Close(ctx)
	if err := syscallCursor.All(ctx, &syscallLogs); err != nil {
		return models.AnomalyContext{}, fmt.Errorf("syscall decode error: %w", err)
	}

	// Combine into AnomalyContext
	ret := models.AnomalyContext{
		Timestamp:     anomaly.Timestamp,
		UID:           anomaly.UID,
		CPU:           cpuLogs,
		Memory:        memoryLogs,
		DiskIO:        diskLogs,
		Network:       flowLogs,
		Syscalls:      syscallLogs,
		AnomalyVector: anomaly,
	}

	return ret, nil
}
