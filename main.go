package main

import (
	"encoding/json"
	"fmt"
	"github.com/emreburakbakdemir/rds-monitoring/collector"
	"log"
	"os"
	"os/exec"
	"time"
	"strings"
)

type Config struct {
	Interval_seconds int      `json:"interval"`
	Paths_to_watch   []string `json:"paths_to_watch"`
	Services          []string `json:"services"`
	Log_file          string   `json:log_file`
	Uptime_thresholds []int    `json:"uptime_thresholds"`
}

type Metrics struct {
	Timestamp string            `json:"timestamp"`
	Host      string            `json:"host"`
	Disk      map[string]string `json:"disk"`
	Memory    map[string]string `json:"memory"`
	CPU       map[string]string `json:"cpu"`
	Services  map[string]string `json:"services"`
	Processes map[string][]string `json:"processes"`
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
	return config, err
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
	var processes map[string][]string
	if config.Uptime_thresholds != nil && len(config.Uptime_thresholds) > 0 {
		processes, _ = collector.FindProcessesByUptime(config.Uptime_thresholds)
	} else {
		processes, _ = collector.FindProcessesByUptime([]int{24}) // default to 24h if not set
	}

	return Metrics{
		Timestamp: time.Now().Format(time.RFC3339),
		Host:      get_hostname(),
		Disk:      collector.CheckDisk("/"),
		Memory:    collector.CheckMemory(),
		CPU:       collector.CheckCPU(),
		Processes: processes,
		Services: map[string]string{
			"supervisor": check_supervisor(),
			"php-fpm":    collector.Check_service("php-fpm"),
			"nginx":      collector.Check_service("nginx"),
		},
	}
}


func main() {
	config, err := load_config("config/conf.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	logDir := config.Log_file[:strings.LastIndex(config.Log_file, "/")]
	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating log directory: %v\n", err)
		os.Exit(1)
	}

	logFile, err := os.OpenFile(config.Log_file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)

	for {
		metrics := collect_metrics(config)
		json_data, err := json.MarshalIndent(metrics, "", " ")
		if err != nil {
			logger.Printf("Error marshalling metrics: %v\n", err)
			continue
		}
		logger.Println(string(json_data))
		fmt.Println(string(json_data))
		time.Sleep(time.Duration(config.Interval_seconds) * time.Second)
	}
}
