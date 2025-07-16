package collector

import (
	"fmt"
	"os/exec"
)

func Check_service(service string) string {
	cmd := exec.Command("supervisorctl", "status", service)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return string(out)
}