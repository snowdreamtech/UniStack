// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build windows

package orchestrator

import (
	"golang.org/x/sys/windows"
)

// getFreeDiskSpace returns the available disk space in bytes for the given path
func getFreeDiskSpace(path string) (uint64, error) {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}

	var freeBytesAvailableToCaller uint64
	var totalNumberOfBytes uint64
	var totalNumberOfFreeBytes uint64

	err = windows.GetDiskFreeSpaceEx(
		pathPtr,
		&freeBytesAvailableToCaller,
		&totalNumberOfBytes,
		&totalNumberOfFreeBytes,
	)
	if err != nil {
		return 0, err
	}

	return freeBytesAvailableToCaller, nil
}

// getTotalMemory returns the total physical memory in bytes
func getTotalMemory() (uint64, error) {
	// Not needed on Windows since preflight blocks Windows execution immediately.
	// Returning 0 to satisfy the compiler.
	return 0, nil
}
