package collector

import (
    "os/exec"
    "strings"
    "strconv"
    "fmt"
)

// FindProcessesByUptime returns a map of hour thresholds to process info slices
func FindProcessesByUptime(thresholds []int) (map[string][]string, error) {
    cmd := exec.Command("ps", "-eo", "pid,etime,comm", "--no-headers")
    out, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to run ps: %v", err)
    }

    result := make(map[string][]string)
    for _, t := range thresholds {
        result[fmt.Sprintf("%dh", t)] = []string{}
    }

    lines := strings.Split(string(out), "\n")
    for _, line := range lines {
        fields := strings.Fields(line)
        if len(fields) < 3 {
            continue
        }
        etime := fields[1]
        hours := parseElapsedHours(etime)
        for _, t := range thresholds {
            if hours >= t {
                key := fmt.Sprintf("%dh", t)
                result[key] = append(result[key], fmt.Sprintf("PID:%s CMD:%s ETIME:%s", fields[0], fields[2], etime))
            }
        }
    }
    return result, nil
}

// parseElapsedHours parses etime ([[dd-]hh:]mm:ss) and returns total hours
func parseElapsedHours(etime string) int {
    days := 0
    hours := 0
    rest := etime
    if strings.Contains(etime, "-") {
        parts := strings.SplitN(etime, "-", 2)
        days, _ = strconv.Atoi(parts[0])
        rest = parts[1]
    }
    timeParts := strings.Split(rest, ":")
    if len(timeParts) == 3 {
        h, _ := strconv.Atoi(timeParts[0])
        hours = h
    }
    totalHours := days*24 + hours
    return totalHours
}