package systemchecks

import (
	"fmt"
	"sort"
	"time"
	"os/exec"
	"strings"

	"github.com/emreburakbakdemir/rds-monitoring/types"

	"github.com/shirou/gopsutil/process"
)

func hoursSince(epochMillis int64) time.Duration {
	createdAt := time.UnixMilli(epochMillis)
	return time.Since(createdAt)
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h!=0 {
		return fmt.Sprintf("%dh%dm", h, m)}
	return fmt.Sprintf("%dm", m)
}

func getTTY(pid int32) (string, error) {
	cmd := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-o", "tty=")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

type processInfo struct {
	PID           int32   `json:"pid"`
	Name          string  `json:"name"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float32 `json:"memory_percent"`
	CreateTime    string  `json:"create_time"`
	Status        string  `json:"status"`
	ParentPID     int32   `json:"parent_pid"`
	RunningHours  string  `json:"running_hours"`
	TTY           string  `json:"tty"`
}

func CheckProcesses(cfg types.Config) interface{} {
	var result []map[string]interface{}

	for _, procCfg := range cfg.Processes {
		if !procCfg.Enabled {
			continue
		}
		procs, err := process.Processes()
		if err != nil {
			return map[string]interface{}{"error": err.Error()}
		}

		filter := procCfg.Filter

		filtered := []processInfo{}
		for _, p := range procs {
			status, _ := p.Status()
			parent, _ := p.Ppid()
			cpu, _ := p.CPUPercent()
			mem, _ := p.MemoryPercent()
			ct, _ := p.CreateTime()
			name, _ := p.Name()
			tty, _ := getTTY(p.Pid)

			if filter.State != "" && filter.State != status {
				continue
			}
			if filter.ParentPID != 0 && filter.ParentPID != parent {
				continue
			}
			if filter.TTY != "" && filter.TTY != tty {
				continue
			}
			if filter.RunningHourThreshold != 0 {
				dur := hoursSince(ct)
				if dur.Hours() < float64(filter.RunningHourThreshold) {
					continue
				}
			}

			createTimeFormatted := time.UnixMilli(ct).Format(time.RFC3339)
			runningHours := formatDuration(hoursSince(ct))

			filtered = append(filtered, processInfo{
				PID:           p.Pid,
				Name:          name,
				CPUPercent:    cpu,
				MemoryPercent: mem,
				CreateTime:    createTimeFormatted,
				Status:        status,
				ParentPID:     parent,
				RunningHours:  runningHours,
				TTY:           tty,
			})
		}

		sortByCPU := func(procs []processInfo) {
			sort.Slice(procs, func(i, j int) bool {
				return procs[i].CPUPercent > procs[j].CPUPercent
			})
		}
		sortByMemory := func(procs []processInfo) {
			sort.Slice(procs, func(i, j int) bool {
				return procs[i].MemoryPercent > procs[j].MemoryPercent
			})
		}
		sortByRunningTime := func(procs []processInfo) {
			sort.Slice(procs, func(i, j int) bool {
				return procs[i].CreateTime < procs[j].CreateTime
			})
		}

		if filter.TopMemoryUsage > 0 {
			sortByMemory(filtered)
			if len(filtered) > filter.TopMemoryUsage {
				filtered = filtered[:filter.TopMemoryUsage]
			}
		}
		if filter.TopCPUUsage > 0 {
			sortByCPU(filtered)
			if len(filtered) > filter.TopCPUUsage {
				filtered = filtered[:filter.TopCPUUsage]
			}
		}
		if filter.TopRunningTime > 0 {
			sortByRunningTime(filtered)
			if len(filtered) > filter.TopRunningTime {
				filtered = filtered[:filter.TopRunningTime]
			}
		}

		// Add limit feature
		if filter.Limit > 0 && len(filtered) > filter.Limit {
			filtered = filtered[:filter.Limit]
		}

		result = append(result, map[string]interface{}{
			"filter":    filter,
			"processes": filtered,
		})
	}
	return result
}