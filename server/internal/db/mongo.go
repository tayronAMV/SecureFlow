package db

import (
	"context"
	
	"server/internal/db/models"
	"time"

	"server/internal/logic"
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

func InsertLog(log string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	LogCollection.InsertOne(ctx, log)
}


var Anomaly_arr = make([]*models.AnomalyLog,100)

func InsertAnomaly_Log(log *models.AnomalyLog){
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	if len(Anomaly_arr) >= 100 {
		logic.Anomaly_detection(Anomaly_arr)
		Anomaly_arr = Anomaly_arr[:0]
	}
	Anomaly_arr = append(Anomaly_arr, log)
	anomalyLogCollection.InsertOne(ctx,log)	
	//send to ui 
	

}







