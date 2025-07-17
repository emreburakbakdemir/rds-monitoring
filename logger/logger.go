package logger

import (
	"log"
	"os"
	"path/filepath"
)

var Log *log.Logger

// Init sets up the logger and creates directories if necessary.
func Init(logPath string) error {
	dir := filepath.Dir(logPath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	Log = log.New(logFile, "", log.LstdFlags)
	return nil
}
