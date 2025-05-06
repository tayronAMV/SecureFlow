package agent

import (
	"fmt"
	"log"
	"sync"

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
		log.Printf("❌ Failed to attach %v to %s: %v", attachType, path, err)
		return
	}

	cw.attached[key] = cgLink
	log.Printf("✅ Successfully attached %v to %s", attachType, path)
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
