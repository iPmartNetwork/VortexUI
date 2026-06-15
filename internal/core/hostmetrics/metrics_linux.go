//go:build linux

package hostmetrics

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var (
	mu        sync.Mutex
	prevIdle  uint64
	prevTotal uint64
)

// Read samples current host utilisation. CPU is the busy percentage since the
// previous call (the first call returns the since-boot average).
func Read() Metrics {
	return Metrics{CPU: cpuPercent(), Mem: memPercent(), Disk: diskPercent("/")}
}

func cpuPercent() float64 {
	idle, total := readCPU()
	mu.Lock()
	defer mu.Unlock()
	di := idle - prevIdle
	dt := total - prevTotal
	prevIdle, prevTotal = idle, total
	if dt == 0 {
		return 0
	}
	usage := (1 - float64(di)/float64(dt)) * 100
	if usage < 0 {
		usage = 0
	}
	if usage > 100 {
		usage = 100
	}
	return usage
}

func readCPU() (idle, total uint64) {
	b, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0
	}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "cpu ") {
			for i, f := range strings.Fields(line)[1:] {
				v, _ := strconv.ParseUint(f, 10, 64)
				total += v
				if i == 3 { // idle is the 4th field
					idle = v
				}
			}
			return idle, total
		}
	}
	return 0, 0
}

func memPercent() float64 {
	b, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	var memTotal, memAvail float64
	for _, line := range strings.Split(string(b), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		v, _ := strconv.ParseFloat(fields[1], 64)
		switch fields[0] {
		case "MemTotal:":
			memTotal = v
		case "MemAvailable:":
			memAvail = v
		}
	}
	if memTotal == 0 {
		return 0
	}
	return (1 - memAvail/memTotal) * 100
}

func diskPercent(path string) float64 {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0
	}
	total := float64(st.Blocks) * float64(st.Bsize)
	free := float64(st.Bavail) * float64(st.Bsize)
	if total == 0 {
		return 0
	}
	return (1 - free/total) * 100
}
