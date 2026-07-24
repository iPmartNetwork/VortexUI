//go:build !windows

package main

import "syscall"

// getFreeDiskMB returns the available disk space in MB for the given path.
func getFreeDiskMB(path string) (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}
	freeBytes := stat.Bavail * uint64(stat.Bsize)
	return freeBytes / (1024 * 1024), nil
}
