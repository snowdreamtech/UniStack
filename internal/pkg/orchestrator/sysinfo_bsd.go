// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build freebsd || openbsd || netbsd || dragonfly

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
	// FreeBSD, OpenBSD, NetBSD typically use hw.physmem
	// Try 64-bit first (e.g. OpenBSD hw.physmem64 if it exists, or 64-bit hw.physmem)
	if val, err := unix.SysctlUint64("hw.physmem64"); err == nil {
		return val, nil
	}
	if val, err := unix.SysctlUint64("hw.physmem"); err == nil {
		return val, nil
	}
	// Fallback for 32-bit BSD kernels where hw.physmem is a uint32
	if val, err := unix.SysctlUint32("hw.physmem"); err == nil {
		return uint64(val), nil
	}
	return 0, unix.EINVAL
}
