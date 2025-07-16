package collector

import (
    "syscall"
    "fmt"
)

func humanReadableDisk(bytes uint64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    div, exp := unit, 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func CheckDisk(path string) map[string]string {
    fs := syscall.Statfs_t{}
    err := syscall.Statfs(path, &fs)
    if err != nil {
        return map[string]string{
            "error": fmt.Sprintf("failed to statfs %s: %v", path, err),
        }
    }
    total := fs.Blocks * uint64(fs.Bsize)
    free := fs.Bfree * uint64(fs.Bsize)
    used := total - free
    usedPercent := float64(used) / float64(total) * 100

    return map[string]string{
        "total": humanReadableDisk(total),
        "used": humanReadableDisk(used),
        "free": humanReadableDisk(free),
        "used_percent": fmt.Sprintf("%.2f", usedPercent),
    }
}