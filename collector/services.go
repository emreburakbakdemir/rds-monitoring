package collector

import (
	"os/exec"
)

func Check_service(service string) string {
	cmd := exec.Command("supervisorctl", "status", service)
	out, err := cmd.Output()
	if err != nil {return "error"}

	return string(out)
}