// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build netbsd

package orchestrator

import (
	"golang.org/x/sys/unix"
)

// getFreeDiskSpace returns the available disk space in bytes for the given path
func getFreeDiskSpace(path string) (uint64, error) {
	var stat unix.Statvfs_t
	if err := unix.Statvfs(path, &stat); err != nil {
		return 0, err
	}
	// F_bavail is the free blocks available to unprivileged user
	return uint64(stat.F_bavail) * uint64(stat.F_frsize), nil
}

// getTotalMemory returns the total physical memory in bytes
func getTotalMemory() (uint64, error) {
	if val, err := unix.SysctlUint64("hw.physmem64"); err == nil {
		return val, nil
	}
	if val, err := unix.SysctlUint64("hw.physmem"); err == nil {
		return val, nil
	}
	if val, err := unix.SysctlUint32("hw.physmem"); err == nil {
		return uint64(val), nil
	}
	return 0, unix.EINVAL
}
