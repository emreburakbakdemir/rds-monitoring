package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"time"
	"fmt"
	"strings"
	"syscall"
)

type Config struct {
	Interval_seconds int `json:"interval"`
	Paths_to_watch []string `json:"paths_to_watch"`
}

type Metrics struct {
	Timestamp string `json:"timestamp"`
	Host string `json:"host"`
	Disk map[string]string `json:"disk"`
	Memory map[string]string `json:"memory"`
	Services map[string]string `json:"services"`
}

func load_config(path string) (Config, error) {
	var config Config
	data, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	return config, err
}

func get_hostname() string {
	host, err := os.Hostname()
	if err != nil {
		return "unknown host"
	}
	return host
}

func check_service(service string) string {
	cmd := exec.Command("supervisorctl", "status", service)
	out, err := cmd.Output()
	if err != nil {
		return "error"
	}

	if strings.Contains(string(out), "RUNNING") {
		return "running"
	}

	return "not running"
}

func collect_metrics(config Config) Metrics {
	return Metrics {
		Timestamp: time.Now().Format(time.RFC3339),
		Host: get_hostname(),
		Disk: map[string]string{"/":"TODO"},
		Memory: map[string]string{"used": "TODO", "free": "TODO"},
		Services: map[string]string{
			"php-fpm": check_service("php-fpm"),
			"nginx": check_service("nginx"),
		},
	}
}

func main() {
	config, err := load_config("conf.json")
	if err != nil {
		fmt.Println("Failed to read the config file.", err)
		os.Exit(1)
	}

	for {
		metrics := collect_metrics(config)
		json_data, _ := json.MarshalIndent(metrics, "", " ")
		fmt.Println(string(json_data))
		time.Sleep(time.Duration(config.Interval_seconds) * time.Second)
	}
}

