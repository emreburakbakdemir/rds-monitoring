package systemchecks

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/disk"
)

type PartitionInfo struct {
	Device       string `json:"device"`
	Mountpoint   string `json:"mountpoint"`
	Fstype       string `json:"fstype"`
	Total        string `json:"total"`
	Used         string `json:"used"`
	Free         string `json:"free"`
	UsedPercent  string `json:"used_percent"`
	Error        string `json:"error,omitempty"`
}

type DiskReport struct {
	Partitions []PartitionInfo `json:"partitions"`
}

func formatSize(bytes uint64) string {
	gb := float64(bytes) / (1024 * 1024 * 1024)
	kb := bytes / 1024
	return fmt.Sprintf("%.2f GB (%d KB)", gb, kb)
}

func CheckDisk(paths []string, mounted bool) DiskReport {
	var results []PartitionInfo

	// Build a map from mountpoint to device name
	partMap := make(map[string]string)
	partitions, _ := disk.Partitions(!mounted) // if seeOnlyMounted==true, only mounted; else, all
	for _, part := range partitions {
		partMap[part.Mountpoint] = part.Device
	}

	for _, mount := range paths {
		usage, err := disk.Usage(mount)
		deviceName := partMap[mount]
		if err != nil {
			results = append(results, PartitionInfo{
				Device:      deviceName,
				Mountpoint:  mount,
				Fstype:      "",
				Total:       "",
				Used:        "",
				Free:        "",
				UsedPercent: "",
				Error:       fmt.Sprintf("Error getting disk usage: %v", err),
			})
			continue
		}

		info := PartitionInfo{
			Device:      deviceName,
			Mountpoint:  usage.Path,
			Fstype:      usage.Fstype,
			Total:       formatSize(usage.Total),
			Used:        formatSize(usage.Used),
			Free:        formatSize(usage.Free),
			UsedPercent: fmt.Sprintf("%.2f%%", usage.UsedPercent),
		}
		results = append(results, info)
	}

	return DiskReport{Partitions: results}
}
