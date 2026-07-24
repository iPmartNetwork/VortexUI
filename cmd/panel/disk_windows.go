//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// getFreeDiskMB returns the available disk space in MB for the given path.
// On Windows this uses wmic to query disk free space.
func getFreeDiskMB(path string) (uint64, error) {
	// Use PowerShell to get free space on the system drive
	out, err := exec.Command("powershell", "-Command",
		"(Get-PSDrive C).Free").Output()
	if err != nil {
		return 0, fmt.Errorf("unable to query disk space: %w", err)
	}
	freeStr := strings.TrimSpace(string(out))
	freeBytes, err := strconv.ParseUint(freeStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse disk space: %w", err)
	}
	return freeBytes / (1024 * 1024), nil
}
