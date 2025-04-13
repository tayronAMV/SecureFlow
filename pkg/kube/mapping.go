package kube

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ContainerMapping struct {
	PodName       string
	Namespace     string
	ContainerID   string
	ContainerName string
	PID           int
}

// FetchContainerMappings connects to the Kubernetes API and maps container IDs to PIDs
func FetchContainerMappings() ([]ContainerMapping, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("cannot load in-cluster config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("cannot create clientset: %w", err)
	}

	pods, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var results []ContainerMapping
	for _, pod := range pods.Items {
		for _, status := range pod.Status.ContainerStatuses {
			cid := strings.TrimPrefix(status.ContainerID, "containerd://")
			if cid == "" {
				continue
			}

			pid, err := getPidFromCrictl(cid)
			if err != nil {
				continue
			}

			results = append(results, ContainerMapping{
				PodName:       pod.Name,
				Namespace:     pod.Namespace,
				ContainerID:  	 cid,
				ContainerName: status.Name,
				PID:           pid,
			})
		}
	}
	return results, nil
}

// getPidFromCrictl runs crictl inspect and extracts the PID
func getPidFromCrictl(containerID string) (int, error) {
	out, err := exec.Command("crictl", "inspect", containerID).Output()
	if err != nil {
		return 0, fmt.Errorf("crictl inspect failed: %w", err)
	}
	var parsed struct {
		Info struct {
			Pid int `json:"pid"`
		} `json:"info"`
	}
	if err := json.Unmarshal(out, &parsed); err != nil {
		return 0, fmt.Errorf("failed to parse crictl output: %w", err)
	}
	return parsed.Info.Pid, nil
}