package systemchecks

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/mem"
)

func CheckMemory(config map[string]bool) map[string]string {
	result := make(map[string]string)

	format := func(b uint64) string {
		gb := float64(b) / (1024 * 1024 * 1024)
		kb := b / 1024
		return fmt.Sprintf("%.2f GB (%d KB)", gb, kb)
	}

	// Physical memory
	vm, err := mem.VirtualMemory()
	if err != nil {
		result["error"] = fmt.Sprintf("VirtualMemory error: %v", err)
		return result
	}

	if config["total"] {
		result["total"] = format(vm.Total)
	}
	if config["available"] {
		result["available"] = format(vm.Available)
	}
	if config["used"] {
		result["used"] = format(vm.Used)
	}
	if config["free"] {
		result["free"] = format(vm.Free)
	}
	if config["used_percent"] {
		result["used_percent"] = fmt.Sprintf("%.2f%%", vm.UsedPercent)
	}
	if config["active"] {
		result["active"] = format(vm.Active)
	}
	if config["inactive"] {
		result["inactive"] = format(vm.Inactive)
	}
	if config["buffers"] {
		result["buffers"] = format(vm.Buffers)
	}
	if config["cached"] {
		result["cached"] = format(vm.Cached)
	}
	if config["shared"] {
		result["shared"] = format(vm.Shared)
	}
	if config["slab"] {
		result["slab"] = format(vm.Slab)
	}
	if config["dirty"] {
		result["dirty"] = format(vm.Dirty)
	}

	// Swap memory
	sm, err := mem.SwapMemory()
	if err == nil {
		if config["swap_total"] {
			result["swap_total"] = format(sm.Total)
		}
		if config["swap_used"] {
			result["swap_used"] = format(sm.Used)
		}
		if config["swap_free"] {
			result["swap_free"] = format(sm.Free)
		}
		if config["swap_used_percent"] {
			result["swap_used_percent"] = fmt.Sprintf("%.2f%%", sm.UsedPercent)
		}
		if config["swap_in"] {
			result["swap_in"] = fmt.Sprintf("%d pages", sm.Sin)
		}
		if config["swap_out"] {
			result["swap_out"] = fmt.Sprintf("%d pages", sm.Sout)
		}
	} else {
		result["swap_error"] = fmt.Sprintf("SwapMemory error: %v", err)
	}

	return result
}
