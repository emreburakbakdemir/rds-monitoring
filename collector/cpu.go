package collector

import (
    "fmt"
    "os"
    "strconv"
    "strings"
    "time"
)

// Get load averages from /proc/loadavg
func getCPULoad() map[string]string {
    data, err := os.ReadFile("/proc/loadavg")
    if err != nil {
        return map[string]string{
            "error": fmt.Sprintf("failed to read /proc/loadavg: %v", err),
        }
    }
    fields := strings.Fields(string(data))
    if len(fields) < 3 {
        return map[string]string{
            "error": "unexpected format in /proc/loadavg",
        }
    }
    return map[string]string{
        "1min":  fields[0],
        "5min":  fields[1],
        "15min": fields[2],
    }
}

// Get CPU usage percentage by reading /proc/stat twice
func getCPUUsage() string {
    getStats := func() (idle, total uint64) {
        data, err := os.ReadFile("/proc/stat")
        if err != nil {
            return 0, 0
        }
        fields := strings.Fields(strings.Split(string(data), "\n")[0])
        if len(fields) < 5 {
            return 0, 0
        }
        var vals [10]uint64
        for i := 1; i < len(fields) && i < 11; i++ {
            vals[i-1], _ = strconv.ParseUint(fields[i], 10, 64)
        }
        idle = vals[3] + vals[4] // idle + iowait
        total = 0
        for _, v := range vals {
            total += v
        }
        return
    }

    idle1, total1 := getStats()
    time.Sleep(500 * time.Millisecond)
    idle2, total2 := getStats()

    if total2 == total1 {
        return "error"
    }
    usage := 100.0 * float64((total2-total1)-(idle2-idle1)) / float64(total2-total1)
    return fmt.Sprintf("%.2f", usage)
}

// Public function to collect CPU metrics
func CheckCPU() map[string]string {
    load := getCPULoad()
    usage := getCPUUsage()
    metrics := map[string]string{
        "usage_percent": usage,
    }
    // Merge load averages
    for k, v := range load {
        metrics["load_"+k] = v
    }
    return metrics
}