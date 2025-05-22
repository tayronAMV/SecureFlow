package models
import (
	"time"
)


type AnomalyLog struct {
	CPU     float64 `json:"cpu" bson:"cpu"`
	DiskIO  float64 `json:"disk_io" bson:"disk_io"`
	Memory  float64 `json:"memory" bson:"memory"`
	Network float64 `json:"network" bson:"network"`
	Syscall float64 `json:"syscall" bson:"syscall"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"` 
	Info Container_info `json : "Info bson:"Info`
	
}

type Container_info struct {
	Namespace      string `json:"namespace" bson:"namespace"`
	PodName        string `json:"pod_name" bson:"pod_name"`
	Container_name string `json:"container_name" bson:"container_name"`
}


type Consumer_msg struct{
	Body any`json:"body"`
	Id  int  `json:"id"`
}

type Consumer_activity_log struct{
	Body string`json:"body"`
	Id  int  `json:"id"`
}

type Consumer_Anomaly_log struct{
	Body AnomalyLog`json:"body"`
	Id  int  `json:"id"`
}

type LogItem struct {
	Timestamp string // optional
	Method    string
	Path      string
	Status    string
	Operation string // e.g., read, write
	Target    string // e.g., user, order
	Type      string // e.g., api_call, static
	Raw       string // original line for reference
}


