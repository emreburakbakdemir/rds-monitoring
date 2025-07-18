package systemchecks

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
   	"github.com/shirou/gopsutil/v3/load"
	"time"
)

func percentages() ([]float64, error) {
    percentages, err := cpu.Percent(time.Second, false)
    if err != nil {
        return nil, fmt.Errorf("cpu.Percent failed: %v", err)
    }
    if len(percentages) == 0 {
        return nil, fmt.Errorf("cpu.Percent returned empty slice")
    }
    return percentages, nil
}

// Public function to collect CPU metrics
func CheckCPU() map[string]string {
    result := make(map[string]string)

    percentages, err := percentages()
    if err != nil {
        result["error"] = fmt.Sprintf("Failed to get CPU percentages: %v", err)
        return result
    }

    loads, err := load.Avg()
    if err != nil {
        result["error"] = fmt.Sprintf("Failed to get load averages: %v", err)
        return result
    } 

    times, err := cpu.Times(false) 
    if err != nil {
        result["error"] = fmt.Sprintf("Failed to get CPU times: %v", err)
        return result
    }
    if len(times) == 0 {
        result["error"] = "CPU times returned empty slice"
        return result
    }
    
    result["time_spent_user"] = fmt.Sprintf("%.2f", times[0].User)
    result["time_spent_system"] = fmt.Sprintf("%.2f", times[0].System)
    result["idle"] = fmt.Sprintf("%.2f", times[0].Idle)
    // result["iowait"] = fmt.Sprintf("%.2f", times[0].Iowait)
    // result["irq"] = fmt.Sprintf("%.2f", times[0].Irq)
    // result["softirq"] = fmt.Sprintf("%.2f", times[0].Softirq)
    // result["steal"] = fmt.Sprintf("%.2f", times[0].Steal)
    // result["guest"] = fmt.Sprintf("%.2f", times[0].Guest)
    // result["guest_nice"] = fmt.Sprintf("%.2f", times[0].GuestNice)
	result["usage_percent"] = fmt.Sprintf("%.2f", percentages[0])
	result["load_1min"] = fmt.Sprintf("%.2f", loads.Load1)
	result["load_5min"] = fmt.Sprintf("%.2f", loads.Load5)
	result["load_15min"] = fmt.Sprintf("%.2f", loads.Load15)

    return result
}