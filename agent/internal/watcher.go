package internal

import (
	"fmt"
	"log"
	"sync"
	"os"
	"strings"
	"path/filepath"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

type CgroupWatcher struct {
	mu       sync.Mutex
	attached map[string]link.Link
}

func New() *CgroupWatcher {
	return &CgroupWatcher{attached: make(map[string]link.Link)}
}

func (cw *CgroupWatcher) AttachIfNeeded(path string, prog *ebpf.Program, attachType ebpf.AttachType) {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	key := path + fmt.Sprintf(":%d", attachType)
	if _, exists := cw.attached[key]; exists {
		return
	}

	cgLink, err := link.AttachCgroup(link.CgroupOptions{
		Path:    path,
		Attach:  attachType,
		Program: prog,
	})
	if err != nil {
		log.Printf(" Failed to attach %v to %s: %v", attachType, path, err)
		return
	}

	cw.attached[key] = cgLink
	log.Printf(" Successfully attached %v to %s", attachType, path)
}

func (cw *CgroupWatcher) Cleanup() {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	for key, l := range cw.attached {
		if err := l.Close(); err != nil {
			log.Printf("⚠️ Failed to close link %s: %v", key, err)
		} else {
			log.Printf("🧹 Cleaned up link: %s", key)
		}
	}
}


// ResolvePathForPID returns the full cgroup v2 path for a given PID.
func ResolvePathForPID(pid int) (string, error) {
  
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cgroup", pid))
	if err != nil {
	  return "", fmt.Errorf("could not read cgroup file for PID %d: %w", pid, err)
	}
  
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
	  parts := strings.SplitN(line, ":", 3)
	  if len(parts) != 3 {
		continue
	  }
	  rel := parts[2]
	  if !strings.Contains(rel, "kubepods") {
		continue
	  }
	  rel = strings.TrimPrefix(rel, "/")
	  full := filepath.Join("/sys/fs/cgroup", rel)
	  if fi, err := os.Stat(full); err == nil && fi.IsDir() {
		return full, nil
	  }
	}
	return "", fmt.Errorf("no kubepods cgroup found for PID %d", pid)
  }

