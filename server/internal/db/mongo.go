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
	LogCollection      		*mongo.Collection
)


func InitMongo() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return err
	}

	mongoClient = client
	LogCollection = client.Database("secureflow").Collection("LogCollection")
	anomalyLogCollection = client.Database("secureflow").Collection("anomalyLogCollection")
	return nil
}


func InsertLog(log interface{}, isAnomalyLog bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if isAnomalyLog {
		anomalyLogCollection.InsertOne(ctx, log)
	}else{
		LogCollection.InsertOne(ctx, log)
	}
	return nil 
}



func FindLogs(anomaly models.AnomalyLog) ([]string ,  error) {
	filter := bson.M{
		"timestamp": bson.M{
			"$gte": anomaly.Timestamp.Add(-10 * time.Second),
			"$lte": anomaly.Timestamp.Add(10 * time.Second),
		},
		"UID": anomaly.UID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var logs []string
	cpuCursor, err := LogCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("cpu query error: %w", err)
	}
	defer cpuCursor.Close(ctx)
	if err := cpuCursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("cpu decode error: %w", err)
	}
	return logs, nil
}


