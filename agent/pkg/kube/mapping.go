package kube

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"os/exec"
	
	"strings"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	// "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ContainerMapping struct {
	PodName       string
	Namespace     string
	ContainerID   string
	ContainerName string
	PID           int
	UID 		  string	
}

var Pid_toContainer_Map = make(map[int]ContainerMapping) // TODO , need to pass it to the main server  , container id and the namespace for every log 

// FetchContainerMappings connects to the Kubernetes API and maps container IDs to PIDs
func FetchContainerMappings() ([]ContainerMapping, error) {
	
	config, err := clientcmd.BuildConfigFromFlags("", "/etc/rancher/k3s/k3s.yaml")
	if err != nil {
		return nil, fmt.Errorf("❌ Cannot load kubeconfig: %w", err)
	}

	log.Println("🔌 Creating Kubernetes clientset...")
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("❌ Cannot create clientset: %w", err)
	}

	log.Println("📦 Fetching all pods from cluster...")
	pods, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("❌ Failed to list pods: %w", err)
	}

	var results []ContainerMapping
	log.Printf("📦 Found %d pods, iterating...",len(pods.Items))
	for _, pod := range pods.Items {
		// if pod.Namespace == "kube-system" || pod.Namespace == "kube-public" || pod.Namespace == "kube-node-lease" {
		// 	continue
		// }
		for _, status := range pod.Status.ContainerStatuses {
			cid := strings.TrimPrefix(status.ContainerID, "containerd://")
			if cid == "" {
				log.Printf("⚠️ Empty ContainerID for container %s in pod %s/%s", status.Name, pod.Namespace, pod.Name)
				continue
			}	

			log.Printf("🔍 Resolving PID for container %s (%s/%s)...", status.Name, pod.Namespace, pod.Name)
			pid, err := getPidFromCrictl(cid)
			if err != nil {
				log.Printf("⚠️ Failed to get PID for container %s (%s): %v", cid, status.Name, err)
				continue
			}

			results = append(results, ContainerMapping{
				PodName:       pod.Name,
				Namespace:     pod.Namespace,
				ContainerID:   cid,
				ContainerName: status.Name,
				PID:           pid,
				UID:           string(pod.UID) ,
			})
			log.Printf("✅ Added mapping: %s/%s → PID %d", pod.Namespace, pod.Name, pid)
		}
	}

	log.Printf("✅ Total container mappings collected: %d", len(results))
	return results, nil
}

// getPidFromDocker uses `docker inspect` to extract the PID of a container
// getPidFromCrictl uses `crictl inspect` to extract the PID of a container
func getPidFromCrictl(containerID string) (int, error) {
	// Remove the "cri-o://" or "containerd://" prefix if present
	containerID = strings.TrimPrefix(containerID, "cri-o://")
	containerID = strings.TrimPrefix(containerID, "containerd://")

	out, err := exec.Command("crictl", "inspect", containerID).Output()
	if err != nil {
		return 0, fmt.Errorf("crictl inspect failed: %w", err)
	}

	// Parse JSON and extract .info.pid
	var result struct {
		Info struct {
			Pid int `json:"pid"`
		} `json:"info"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return 0, fmt.Errorf("failed to parse crictl inspect output: %w", err)
	}

	return result.Info.Pid, nil
}
func PidToUid(pid int) string {
	return Pid_toContainer_Map[pid].UID 
}