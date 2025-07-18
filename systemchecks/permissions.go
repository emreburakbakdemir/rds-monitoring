package systemchecks

import (
	"os"
	"github.com/emreburakbakdemir/rds-monitoring/types"
)

func CheckPermissions(cfg types.Config) interface{} {
	if !cfg.Permissions.Enabled {
		return nil
	}
	results := make(map[string]interface{})
	for _, path := range cfg.Permissions.Paths {
		info, err := os.Stat(path)
		if err != nil {
			results[path] = map[string]interface{}{
				"error": err.Error(),
			}
			continue
		}
		mode := info.Mode()
		results[path] = map[string]interface{}{
			"permissions": mode.Perm().String(),
			"is_dir":      mode.IsDir(),
			"mode":        mode.String(),
		}
	}
	return results
}