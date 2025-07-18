package main

import (
	"encoding/json"
	"fmt"
	"github.com/emreburakbakdemir/rds-monitoring/systemchecks"
	"github.com/emreburakbakdemir/rds-monitoring/servicechecks"
	"github.com/emreburakbakdemir/rds-monitoring/logger"
	"github.com/shirou/gopsutil/v3/disk"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"time"
	"github.com/emreburakbakdemir/rds-monitoring/types"
)

func load_config(path string) (types.Config, error) {
    var config types.Config
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



func collect_metrics(config types.Config) types.Metrics {
    // Use the first enabled process config
    // var procConf types.ProcessMonitoringConfig
    // for _, p := range config.Processes {
    //     if p.Enabled {
    //         procConf = p.Filter
    //         break
    //     }
    // }

    processes := systemchecks.CheckProcesses(config)    

    // Use the first enabled disk config
    var diskPaths []string
    var diskMounted bool
    for _, d := range config.Disk {
        if d.Enabled {
            diskPaths = d.Disk_metrics.PathsToWatchDisk
            diskMounted = d.Disk_metrics.Mounted
            break
        }
    }
    if len(diskPaths) == 1 && diskPaths[0] == "*" {
        partitions, err := disk.Partitions(!diskMounted)
        if err == nil {
            for _, part := range partitions {
                diskPaths = append(diskPaths, part.Mountpoint)
            }
        } else {
            diskPaths = []string{"/"}
        }
    }

    // Use the first enabled memory config
    var memoryMetrics map[string]bool
    for _, m := range config.Memory {
        if m.Enabled {
            memoryMetrics = m.MemoryMetrics
            break
        }
    }

    return types.Metrics{
        Timestamp:     time.Now().Format(time.RFC3339),
        Host:          get_hostname(),
        Disk:          systemchecks.CheckDisk(diskPaths, diskMounted),
        Memory:        systemchecks.CheckMemory(memoryMetrics),
        CPU:           systemchecks.CheckCPU(),
        // ProcessSortBy: procConf.SortBy,
        Processes:     processes,
        Services: map[string]string{
            "supervisor": check_supervisor(),
            "php-fpm":    servicechecks.CheckService("php-fpm"),
            "nginx":      servicechecks.CheckService("nginx"),
        },
		Permissions: systemchecks.CheckPermissions(config),
    }
}

func main() {
    config, err := load_config("config/conf.yaml")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
        os.Exit(1)
    }

    err = logger.Init(config.General.Log_file)
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
        time.Sleep(time.Duration(config.General.Interval_seconds) * time.Second)
    }
}
