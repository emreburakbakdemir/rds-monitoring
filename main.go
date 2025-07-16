package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"time"
	"fmt"
	"github.com/emreburakbakdemir/rds-monitoring/collector"

)

type Config struct {
	Interval_seconds int `json:"interval"`
	Paths_to_watch []string `json:"paths_to_watch"`
	Services []string `json:"services"`
	Log_file string `json:log_file`
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
	fmt.Println(string(data))
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

// func check_service(service string) string {
// 	cmd := exec.Command("supervisorctl", "status", service)
// 	out, err := cmd.Output()
// 	// fmt.Println(string(out),err)
// 	if err != nil {
// 		return "error"
// 	}

// 	if strings.Contains(string(out), "RUNNING") {
// 		return string(out)
// 	}

// 	return "not running"
// }

func check_supervisor() string {
	cmd := exec.Command("systemctl", "is-active", "supervisor")
	out, err := cmd.Output()
	if err != nil {return "error"}

	return string(out)
}

func collect_metrics(config Config) Metrics {
	return Metrics {
		Timestamp: time.Now().Format(time.RFC3339),
		Host: get_hostname(),
		Disk: map[string]string{"/":"TODO"},
		Memory: map[string]string{"used": "TODO", "free": "TODO"},
		Services: map[string]string{
			"supervisor": check_supervisor(),
			"php-fpm": collector.Check_service("php-fpm"),
			"nginx": collector.Check_service("nginx"),
		},
	}
}

func main() {
	fmt.Println("saa")
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

