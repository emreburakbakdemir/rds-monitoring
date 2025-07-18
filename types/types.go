package types



type DiskMetricsConfig struct {
	Mounted          bool     `yaml:"mounted"`
	PathsToWatchDisk []string `yaml:"paths_to_watch_disk"`
}	

type ProcessMonitoringConfig struct {
	Limit            int      `yaml:"limit"`
	SortBy           string   `yaml:"sort_by"`
	UptimeThresholds []int    `yaml:"uptime_thresholds"`
}	

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

type Metrics struct {
    Timestamp     string                         `json:"timestamp"`
    Host          string                         `json:"host"`
    Disk          DiskReport        `json:"disk"`
    Memory        map[string]string              `json:"memory"`
    CPU           map[string]string              `json:"cpu"`
    Services      map[string]string              `json:"services"`
    // ProcessSortBy string                         `json:"process_sort_by"`
    Processes     interface{}     `json:"processes"`
	Permissions  interface{}             `json:"permissions,omitempty"`
}

type Config struct {
	General struct {
		Interval_seconds   int `yaml:"check_interval"`
		Log_file                 string                  `yaml:"log_file"`
	} `yaml:"general"`

	Memory []struct {
		MemoryMetrics map[string]bool `yaml:"memory_metrics"`
		Enabled bool `yaml:"enabled"`
	} `yaml:"memory"`
	
	Services []struct {
		Enabled bool `yaml:"enabled"`
		Services []string `yaml:"services"`
	} `yaml:"services"`

	CPU struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"cpu"`

	Disk []struct {
		Enabled bool `yaml:"enabled"`
		Filter  struct {
			SortBy				string	`yaml:"sort_by"`
			TopDiskSize 		int		`yaml:"top_disk_size"`
			TopDiskUsage    	int		`yaml:"top_disk_usage"`
			TopDiskUsagePercent	int		`yaml:"top_disk_usage_percent"`
			TopFreeSpace		int		`yaml:"top_free_space"`
		} `yaml:"filter"`
		Disk_metrics DiskMetricsConfig `yaml:"disk_metrics"`
	} `yaml:"disk"`

	Processes []struct {
		Enabled bool `yaml:"enabled"`
		Filter  struct {
			Limit int `yaml:"limit"`
			RunningHourThreshold int    `yaml:"running_hour_threshold"`
			TopMemoryUsage       int    `yaml:"top_memory_usage"`
			TopCPUUsage          int    `yaml:"top_cpu_usage"`
			State                string `yaml:"state"`
			ParentPID            int32  `yaml:"parent_pid"`
			TTY                  string `yaml:"tty"`
			TopRunningTime       int    `yaml:"top_running_time"`
		} `yaml:"filter"`
	} `yaml:"processes"`

	Permissions struct {
		Enabled bool     `yaml:"enabled"`
		Paths   []string `yaml:"paths"`
	} `yaml:"permissions"`

	Network struct {
		Enabled         bool `yaml:"enabled"`
		CheckOpenPorts  bool `yaml:"check_open_ports"`
		CheckInterfaces bool `yaml:"check_interfaces"`
	} `yaml:"network"`
}