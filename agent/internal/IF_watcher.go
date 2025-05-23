package internal

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"os"
	"path/filepath"
	"os/exec"
	"regexp"
	"sync"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"agent/pkg/kube"
	"log"
	"net"
	
)

func GetPeerIfindexFromContainerEth0(pid int) (int, error) {
	cmd := exec.Command("nsenter", "-t", fmt.Sprint(pid), "-n", "ip", "link", "show", "eth0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("failed to run nsenter: %w\n%s", err, string(out))
	}

	// Output will be like: "2: eth0@if12: ..."
	// We extract the 12 using regex
	re := regexp.MustCompile(`eth0@if(\d+)`)
	matches := re.FindStringSubmatch(string(out))
	if len(matches) < 2 {
		return 0, fmt.Errorf("could not extract peer ifindex from: %s", string(out))
	}

	peerIfindex, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid peer ifindex: %w", err)
	}

	return peerIfindex, nil
}

func FindHostVethByIndex(targetIfindex int) (string, error) {
	files, err := os.ReadDir("/sys/class/net")
	if err != nil {
		return "", fmt.Errorf("failed to read /sys/class/net: %w", err)
	}

	for _, entry := range files {
		iface := entry.Name()
		path := filepath.Join("/sys/class/net", iface, "ifindex")

		data, err := ioutil.ReadFile(path)
		if err != nil {
			continue
		}
		ifindex, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil {
			continue
		}

		if ifindex == targetIfindex {
			return iface, nil
		}
	}

	return "", fmt.Errorf("no interface found with ifindex %d", targetIfindex)
}




type LinkTracker struct {
	mu             sync.RWMutex
	attachedLinks  map[int][]link.Link  // ifindex -> [ingress_link, egress_link]
	attachedIfaces map[int]string       // ifindex -> interface name
}

func NewLinkTracker() *LinkTracker {
	return &LinkTracker{
		attachedLinks:  make(map[int][]link.Link),
		attachedIfaces: make(map[int]string),
	}
}

func (lt *LinkTracker) IsAttached(ifindex int) bool {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	links, exists := lt.attachedLinks[ifindex]
	return exists && len(links) > 0
}

func (lt *LinkTracker) AddLinks(ifindex int, ifaceName string, ingressLink, egressLink link.Link) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	lt.attachedLinks[ifindex] = []link.Link{ingressLink, egressLink}
	lt.attachedIfaces[ifindex] = ifaceName
}

func (lt *LinkTracker) GetAttachedInterfaces() []string {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	
	interfaces := make([]string, 0, len(lt.attachedIfaces))
	for _, name := range lt.attachedIfaces {
		interfaces = append(interfaces, name)
	}
	return interfaces
}

func (lt *LinkTracker) CloseAll() {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	
	for ifindex, links := range lt.attachedLinks {
		for i, l := range links {
			if err := l.Close(); err != nil {
				log.Printf("❌ Failed to close link %d for ifindex %d: %v", i, ifindex, err)
			}
		}
		log.Printf("✅ Closed links for interface: %s", lt.attachedIfaces[ifindex])
	}
	
	// Clear the maps
	lt.attachedLinks = make(map[int][]link.Link)
	lt.attachedIfaces = make(map[int]string)
}

func attachToContainers(objs *struct {
	TcIngress *ebpf.Program `ebpf:"tc_ingress"`
	TcEgress  *ebpf.Program `ebpf:"tc_egress"`
	Events    *ebpf.Map     `ebpf:"events"`
}, tracker *LinkTracker) error {
	
	containers, err := kube.FetchContainerMappings()
	if err != nil {
		return fmt.Errorf("failed to fetch container mappings: %v", err)
	}

	if len(containers) == 0 {
		log.Println("⚠️  No containers found")
		return nil
	}

	log.Printf("🔍 Found %d containers", len(containers))

	for _, container := range containers {
		// Get the peer veth interface index from container's eth0
		ifindex, err := GetPeerIfindexFromContainerEth0(container.PID)
		if err != nil {
			log.Printf("❌ Failed to get peer ifindex for container PID %d: %v", container.PID, err)
			continue
		}

		// Skip if already attached
		if tracker.IsAttached(ifindex) {
			log.Printf("⏭️  Interface with index %d already attached, skipping", ifindex)
			continue
		}

		// Find the host veth interface name
		ifaceName, err := FindHostVethByIndex(ifindex)
		if err != nil {
			log.Printf("❌ Failed to find host veth by index %d: %v", ifindex, err)
			continue
		}

		// Get interface by name
		iface, err := net.InterfaceByName(ifaceName)
		if err != nil {
			log.Printf("❌ Failed to get interface %s: %v", ifaceName, err)
			continue
		}

		// Attach ingress TC program
		ingressLink, err := link.AttachTCX(link.TCXOptions{
			Interface: iface.Index,
			Program:   objs.TcIngress,
			Attach:    ebpf.AttachTCXIngress,
		})
		if err != nil {
			log.Printf("❌ Failed to attach TCX ingress to %s (index: %d): %v", ifaceName, ifindex, err)
			continue
		}

		// Attach egress TC program
		egressLink, err := link.AttachTCX(link.TCXOptions{
			Interface: iface.Index,
			Program:   objs.TcEgress,
			Attach:    ebpf.AttachTCXEgress,
		})
		if err != nil {
			log.Printf("❌ Failed to attach TCX egress to %s (index: %d): %v", ifaceName, ifindex, err)
			ingressLink.Close() // Clean up ingress link
			continue
		}

		// Track both links
		tracker.AddLinks(ifindex, ifaceName, ingressLink, egressLink)
		log.Printf("✅ Attached ingress+egress to interface: %s (index: %d) for container PID: %d", 
			ifaceName, ifindex, container.PID)
	}

	attachedIfaces := tracker.GetAttachedInterfaces()
	if len(attachedIfaces) > 0 {
		log.Printf("🎯 Successfully attached to %d interfaces: %v", len(attachedIfaces), attachedIfaces)
	} else {
		log.Println("⚠️  No interfaces were successfully attached")
	}

	return nil
}



