package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	

	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	
	"k8s.io/client-go/tools/clientcmd"
)


func GetKubernetesClient() (*kubernetes.Clientset, error) {
	// If inside cluster
	config, err := clientcmd.BuildConfigFromFlags("", "/etc/rancher/k3s/k3s.yaml")
	if err != nil {
		fmt.Println("nooooooo clientswet")
		return nil , err
	}
	return kubernetes.NewForConfig(config)
}




func FetchPodLogs(clientset *kubernetes.Clientset, namespace, podName, containerName string,anomalyTime time.Time) (string, error) {
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container:  containerName,
		SinceTime:  &v1.Time{Time: anomalyTime.Add(-3 * time.Minute)}, // fetch logs from within 3 min when the anomaly 
		TailLines:  int64Ptr(500),
		Timestamps: true,
	})	
	podLogs, err := req.Stream(context.Background())
	if err != nil {
		return "", fmt.Errorf("error opening stream: %v", err)
	}
	defer podLogs.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, podLogs)
	if err != nil {
		return "", fmt.Errorf("error reading logs: %v", err)
	}
	
	return buf.String(),nil
}

func int64Ptr(i int64) *int64 {
	return &i
}