package main

import (
	"encoding/json"
	"fmt"
	"github.com/emreburakbakdemir/rds-monitoring/systemchecks"
	"github.com/emreburakbakdemir/rds-monitoring/servicescheck"
	"github.com/emreburakbakdemir/rds-monitoring/logger"
	"github.com/shirou/gopsutil/v3/disk"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"time"
)

type ProcessMonitoringConfig struct {
	Limit            int      `yaml:"limit"`
	// SortingOptions   []string `yaml:"sorting_options"`
	SortBy           string   `yaml:"sort_by"`
	UptimeThresholds []int    `yaml:"uptime_thresholds"`
}

type DiskMetricsConfig struct {
	Mounted          bool     `yaml:"mounted"`
	PathsToWatchDisk []string `yaml:"paths_to_watch_disk"`
}

type Config struct {
	Interval_seconds         int                     `yaml:"check_interval"`
	Disk_metrics             DiskMetricsConfig       `yaml:"disk_metrics"`
	Services                 []string                `yaml:"services"`
	Log_file                 string                  `yaml:"log_file"`
	Uptime_thresholds        []int                   `yaml:"cpu_uptime_thresholds"`
	MemoryMetrics            map[string]bool         `yaml:"memory_metrics"`
	ProcessMonitoringConfig  ProcessMonitoringConfig `yaml:"process_monitoring"`
}

type Metrics struct {
	Timestamp string `json:"timestamp"`
	Host      string `json:"host"`
	Disk      collector.DiskReport `json:"disk"`
	Memory    map[string]string `json:"memory"`
	CPU       map[string]string `json:"cpu"`
	Services  map[string]string `json:"services"`
	ProcessSortBy string `json:"process_sort_by"`
	Processes []collector.ProcessInfo `json:"processes"`
}

func load_config(path string) (Config, error) {
	var config Config
	data, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse config file: %w", err)
	}
	return config, nil
}

func get_hostname() string {
	host, err := os.Hostname()
	if err != nil {
		return "unknown host"
	}
	return host
}

func check_supervisor() string {
	cmd := exec.Command("systemctl", "is-active", "supervisor")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return string(out)
}

func collect_metrics(config Config) Metrics {
	// Collect detailed process metrics
	processes, err := collector.CheckProcesses(config.ProcessMonitoringConfig.Limit,
		 										config.ProcessMonitoringConfig.SortBy,
												config.ProcessMonitoringConfig.UptimeThresholds)
	if err != nil {
		logger.Log.Printf("Error collecting detailed process metrics: %v", err)
		processes = []collector.ProcessInfo{}
	}

	var diskPaths []string
	if len(config.Disk_metrics.PathsToWatchDisk) == 1 && config.Disk_metrics.PathsToWatchDisk[0] == "*" {
		partitions, err := disk.Partitions(!config.Disk_metrics.Mounted)
		if err == nil {
			for _, part := range partitions {
				diskPaths = append(diskPaths, part.Mountpoint)
			}
		} else {
			// monitor root only if error
			diskPaths = []string{"/"}
		}
	} else {
		diskPaths = config.Disk_metrics.PathsToWatchDisk
	}

	// diskReport := collector.CheckDisk(diskPaths, config.disk_metrics.mounted)

	return Metrics{
		Timestamp: time.Now().Format(time.RFC3339),
		Host:      get_hostname(),
		Disk:      collector.CheckDisk(diskPaths, config.Disk_metrics.Mounted),
		Memory:    collector.CheckMemory(config.MemoryMetrics),
		CPU:       collector.CheckCPU(),
		ProcessSortBy: config.ProcessMonitoringConfig.SortBy,
		Processes: processes,
		Services: map[string]string{
			"supervisor": check_supervisor(),
			"php-fpm":    collector.CheckService("php-fpm"),
			"nginx":      collector.CheckService("nginx"),
		},
	}
}

func main() {
	config, err := load_config("config/conf.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	err = logger.Init(config.Log_file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		os.Exit(1)
	}

	for {
		metrics := collect_metrics(config)
		json_data, err := json.MarshalIndent(metrics, "", " ")
		if err != nil {
			logger.Log.Printf("Error marshalling metrics: %v\n", err)
			continue
		}
		logger.Log.Println(string(json_data)) // log to file
		fmt.Println(string(json_data))        // also print to console
		time.Sleep(time.Duration(config.Interval_seconds) * time.Second)
	}
}
