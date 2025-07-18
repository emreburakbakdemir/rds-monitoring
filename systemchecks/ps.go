package systemchecks

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/process"
	"sort"
	"time"
)

type ProcessInfo struct {
    PID int32 `json:"pid"`
    ParentPID int32 `json:"parent_pid"`
    Command string `json:"command"`
    Args []string `json:"args"`
    Name string `json:"name"`
    CPUPercent float64 `json:"cpu_percent"`
    MemoryPercent float32 `json:"memory_percent"`
    Status []string `json:"status"`
    User string `json:"user"`
    NiceValue int32 `json:"nice_value"`
    Uptime string `json:"uptime"`
    CreateTime string `json:"create_time"`
}

func CheckProcesses(limit int, sortBy string, uptime_thresholds []int) ([]ProcessInfo, error) {
    processes, err := process.Processes()
    if err != nil {
        return nil, fmt.Errorf("failed to get processes: %v", err)
    }

    var detailed []ProcessInfo

    for _, p := range processes {
        if limit > 0 && len(detailed) >= limit {
            break
        }

        name, _ := p.Name()
        exe, _ := p.Exe()
        cmdline, _ := p.CmdlineSlice()
        username, _ := p.Username()
        cpuPercent, _ := p.CPUPercent()
        memPercent, _ := p.MemoryPercent()
        status, _ := p.Status()
        nice, _ := p.Nice()
        createTimeMillis, _ := p.CreateTime()
        ppid, _ := p.Ppid()

        uptime := ""
        if createTimeMillis > 0 {
            uptimeDuration := time.Since(time.UnixMilli(createTimeMillis))
            uptime = uptimeDuration.String()
        }

        info := ProcessInfo{
            PID:           p.Pid,
            ParentPID:     ppid,
            Command:       exe,
            Args:          cmdline,
            Name:          name,
            CPUPercent:    cpuPercent,
            MemoryPercent: memPercent,
            Status:        status,
            User:          username,
            NiceValue:     nice,
            Uptime:        uptime,
            CreateTime:    time.UnixMilli(createTimeMillis).Format(time.RFC3339),
        }

        detailed = append(detailed, info)
    }

    if len(uptime_thresholds) > 0 {
        var filtered []ProcessInfo
        for _, p := range detailed {
            uptimeDuration, err := time.ParseDuration(p.Uptime)
            if err != nil {
                continue // skip if uptime cannot be parsed
            }
            hours := uptimeDuration.Hours()
            for _, threshold := range uptime_thresholds {
                if hours >= float64(threshold) {
                    filtered = append(filtered, p)
                    break // include process if it matches any threshold
                }
            }
        }
        detailed = filtered
    }

    switch sortBy {
        case "cpu_percent":
            sort.Slice(detailed, func(i, j int) bool {
                return detailed[i].CPUPercent > detailed[j].CPUPercent
            })
        case "memory_percent":
            sort.Slice(detailed, func(i, j int) bool {
                return detailed[i].MemoryPercent > detailed[j].MemoryPercent
            })
        case "uptime":
            sort.Slice(detailed, func(i, j int) bool {  
                uptimeI, _ := time.ParseDuration(detailed[i].Uptime)
                uptimeJ, _ := time.ParseDuration(detailed[j].Uptime)
                return uptimeI > uptimeJ
            })
        default:
            sort.Slice(detailed, func(i, j int) bool {
                return detailed[i].PID < detailed[j].PID
            })
    }
	if len(detailed) > limit {
		detailed = detailed[:limit]
	}

    return detailed, nil
}