
package utils

import (
  "fmt"
  "log"
  "os"
  "path/filepath"
  "strings"
)

// ResolvePathForPID returns the full cgroup v2 path for a given PID.
func ResolvePathForPID(pid int) (string, error) {
  log.Printf("🔍 Resolving cgroup for PID %d", pid)
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
