package collector

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func humanReadableMem(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := unit, 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func CheckMemory() map[string]string {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return map[string]string{
			"error": fmt.Sprintf("failed to read /proc/meminfo: %v", err),
		}
	}

	memInfo := make(map[string]uint64)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		memInfo[key] = value * 1024 // kB to bytes
	}

	total := memInfo["MemTotal"]
	free := memInfo["MemFree"] + memInfo["Buffers"] + memInfo["Cached"]
	used := total - free
	usedPercent := float64(used) / float64(total) * 100

	return map[string]string{
		"total":        humanReadableMem(total),
		"used":         humanReadableMem(used),
		"free":         humanReadableMem(free),
		"used_percent": fmt.Sprintf("%.2f", usedPercent),
	}
}
