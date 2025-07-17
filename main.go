package main

import (
	"encoding/json"
	"fmt"
	"github.com/emreburakbakdemir/rds-monitoring/collector"
	"github.com/emreburakbakdemir/rds-monitoring/logger"
	"os"
	"os/exec"
	"time"
)

type Config struct {
	Interval_seconds  int      `json:"interval"`
	Paths_to_watch    []string `json:"paths_to_watch"`
	Services          []string `json:"services"`
	Log_file          string   `json:"log_file"`
	Uptime_thresholds []int    `json:"uptime_thresholds"`
	Paths_to_watch_disk []string `json:"paths_to_watch_disk"`
}

type Metrics struct {
	Timestamp string                         `json:"timestamp"`
	Host      string                         `json:"host"`
	Disk      map[string]map[string]string   `json:"disk"`
	Memory    map[string]string              `json:"memory"`
	CPU       map[string]string              `json:"cpu"`
	Services  map[string]string              `json:"services"`
	Processes map[string][]string            `json:"processes"`
}

func load_config(path string) (Config, error) {
	var config Config
	data, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}
	err = json.Unmarshal(data, &config)
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
	// Collect disk metrics
	diskMetrics := make(map[string]map[string]string)
	for i, path := range config.Paths_to_watch_disk {
		diskMetrics[path] = collector.CheckDisk(config.Paths_to_watch_disk[i]) // assuming the first path is the main one
	}

	// Collect memory metrics
	var processes map[string][]string
	if config.Uptime_thresholds != nil && len(config.Uptime_thresholds) > 0 {
		processes, _ = collector.FindProcessesByUptime(config.Uptime_thresholds)
	} else {
		processes, _ = collector.FindProcessesByUptime([]int{24}) // default to 24h
	}

	return Metrics{
		Timestamp: time.Now().Format(time.RFC3339),
		Host:      get_hostname(),
		Disk:      diskMetrics, // now supports multiple directories
		Memory:    collector.CheckMemory(),
		CPU:       collector.CheckCPU(),
		Processes: processes,
		Services: map[string]string{
			"supervisor": check_supervisor(),
			"php-fpm":    collector.CheckService("php-fpm"),
			"nginx":      collector.CheckService("nginx"),
		},
	}
}

func main() {
	config, err := load_config("config/conf.json")
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
