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
