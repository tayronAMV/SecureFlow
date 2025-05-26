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
	Container ContainerMapping `json:"container" bson:"container"`
}

type ContainerMapping struct {
	PodName       string `json:"pod_name" bson:"pod_name"`
	Namespace     string `json:"namespace" bson:"namespace"`
	ContainerID   string `json:"container_id" bson:"container_id"`
	ContainerName string `json:"container_name" bson:"container_name"`
	PID           int    `json:"pid" bson:"pid"`
	UID           string `json:"uid" bson:"uid"`
	Cgroup        string `json:"cgroup" bson:"cgroup"`
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


