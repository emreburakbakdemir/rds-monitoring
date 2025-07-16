package collector

import (
	"strings"
	"strconv"
	"os"
)

func extractKB(line string) int {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0
	}
	val, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0
	}
	return val
}

func Check_memory() map[string]string {
	meminfo, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return map[string]string{
			"error": "cannot read /proc/meminfo",
		}
	}

	lines := strings.Split(string(meminfo), "\n")
	var total, free int

	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			total = extractKB(line)
		} else if strings.HasPrefix(line, "MemFree:") {
			free = extractKB(line)
		}
	}
	used := total - free

	return map[string]string {
		"total_kb": strconv.Itoa(total),
		"used_kb": strconv.Itoa(used),
		"free_kb": strconv.Itoa(free),
	}
}
