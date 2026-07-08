//go:build linux

package orchestrator

import (
	"golang.org/x/sys/unix"
)

// getFreeDiskSpace returns the available disk space in bytes for the given path
func getFreeDiskSpace(path string) (uint64, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return 0, err
	}
	// Bavail is the free blocks available to unprivileged user
	return uint64(stat.Bavail) * uint64(stat.Bsize), nil
}

// getTotalMemory returns the total physical memory in bytes
func getTotalMemory() (uint64, error) {
	var info unix.Sysinfo_t
	if err := unix.Sysinfo(&info); err != nil {
		return 0, err
	}
	return uint64(info.Totalram) * uint64(info.Unit), nil
}
