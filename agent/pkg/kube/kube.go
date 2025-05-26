package kube

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"syscall"
	"os/exec"
	"path/filepath"
	"strings"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	// "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os/signal"
	"sync"
	"time"
	
)

type ContainerMapping struct {
	PodName       string `json:"pod_name" bson:"pod_name"`
	Namespace     string `json:"namespace" bson:"namespace"`
	ContainerID   string `json:"container_id" bson:"container_id"`
	ContainerName string `json:"container_name" bson:"container_name"`
	PID           int    `json:"pid" bson:"pid"`
	UID           string `json:"uid" bson:"uid"`
	Cgroup        string `json:"cgroup" bson:"cgroup"`
}

var Cgroup_mapping = make(map[uint64]ContainerMapping)
var Pid_toContainer_Map = make(map[int]ContainerMapping) // TODO , need to pass it to the main server  , container id and the namespace for every log 

var (
	Cur_Map ,_= FetchContainerMappings() 
	mu      sync.RWMutex
	cgroup_mu sync.RWMutex
	cond = sync.NewCond(&mu)
)


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
		if pod.Namespace == "kube-system" || pod.Namespace == "kube-public" || pod.Namespace == "kube-node-lease" {
			continue
		}
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

			container := ContainerMapping{
				PodName:       pod.Name,
				Namespace:     pod.Namespace,
				ContainerID:   cid,
				ContainerName: status.Name,
				PID:           pid,
				UID:           string(pod.UID) ,
			}
			results = append(results,container)
			log.Printf("✅ Added mapping: %s/%s → PID %d", pod.Namespace, pod.Name, pid)


			cgroupID , err:= GetContainerCgroupID(pid)

			if err != nil {
				fmt.Println("coudnt get cgroup nooooo  ",err)
				continue
			}
			cgroup_mu.Lock()
			if _ , ok := Cgroup_mapping[cgroupID] ; !ok {
				Cgroup_mapping[cgroupID] = container 
				fmt.Printf("pod is %s with pid %d and the cgroup id is %d \n ",container.PodName,container.PID ,cgroupID)
			}
			cgroup_mu.Unlock()
		}
	}

	log.Printf("✅ Total container mappings collected: %d", len(results))
	return results, nil
}

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



func GetContainerCgroupID(pid int) (uint64, error) {
	cgroupPath, err := readCgroupPath(pid)
	if err != nil {
		return 0, err
	}

	fullPath := filepath.Join("/sys/fs/cgroup", cgroupPath)
	stat, err := os.Stat(fullPath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat cgroup path: %w", err)
	}

	return stat.Sys().(*syscall.Stat_t).Ino, nil
}

func readCgroupPath(pid int) (string, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cgroup", pid))
	if err != nil {
		return "", fmt.Errorf("could not read cgroup file for PID %d: %w", pid, err)
	}

	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		parts := strings.SplitN(line, ":", 3)
		if len(parts) == 3 && strings.HasPrefix(parts[2], "/kubepods") {
			return parts[2], nil // already relative
		}
	}
	return "", fmt.Errorf("no kubepods cgroup found for PID %d", pid)
}



func KillPod(clientset *kubernetes.Clientset, namespace, podName string) error {
    return clientset.CoreV1().Pods(namespace).Delete(context.TODO(), podName, metav1.DeleteOptions{})
}




func MappingTracker() {
    rescanTicker := time.NewTicker(30 * time.Second)
    defer rescanTicker.Stop()

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

    for {
        select {
        case <-stop:
            return

        case <-rescanTicker.C:
            updated_map, err := FetchContainerMappings()
            if err != nil {
                fmt.Println("❌ Couldn't get containers:", err)
                continue
            }

            setTracker(updated_map)

            mu.Lock()
            cond.Broadcast()
            mu.Unlock()
        }
    }
}

func Get_Cgroup_mapping(cgroup uint64)(ContainerMapping , bool){
	cgroup_mu.RLock()
	defer cgroup_mu.RUnlock()
	cont , ok :=  Cgroup_mapping[cgroup]

	return cont , ok 
	
}

func GetCurrentMapping()[]ContainerMapping {
	mu.RLock()
	defer mu.RUnlock()
	cp := make([]ContainerMapping, len(Cur_Map))
	copy(cp, Cur_Map)
	return cp
}

func setTracker(cur []ContainerMapping) {
	mu.Lock()
	defer mu.Unlock()
	Cur_Map = cur
}

func GetCondition() *sync.Cond {
    return cond
}