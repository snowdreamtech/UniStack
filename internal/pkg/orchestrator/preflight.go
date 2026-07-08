// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// RunPreflightChecks inspects the host operating system and hardware resources
// to ensure the environment meets the minimum requirements for running Ansible.
func RunPreflightChecks(rootDir string) error {
	fmt.Println("🚀 Running Pre-flight System Checks...")

	// 1. Operating System Check
	if runtime.GOOS == "windows" {
		return fmt.Errorf("🚨 FATAL: Ansible Control Node cannot run natively on Windows. Please use WSL (Windows Subsystem for Linux) to run UniStack")
	}

	// 2. Disk Space Check: Require at least 500MB (500 * 1024 * 1024 bytes) of free space
	freeSpace, err := getFreeDiskSpace(filepath.Dir(rootDir))
	if err == nil {
		if freeSpace < 500*1024*1024 {
			return fmt.Errorf("🚨 FATAL: Insufficient disk space. Required: 500MB, Available: %d MB. Boot aborted to prevent corruption", freeSpace/(1024*1024))
		}
	} else {
		fmt.Printf("⚠️ Warning: Failed to check disk space: %v\n", err)
	}

	// 3. Physical Memory Check: Warn if less than 512MB
	totalMemory, err := getTotalMemory()
	if err == nil {
		if totalMemory < 512*1024*1024 {
			fmt.Printf("⚠️ WARNING: Host has very low memory (%d MB). Ansible execution may fail with Out-Of-Memory (OOM) errors. Recommended: 1024 MB+\n", totalMemory/(1024*1024))
		}
	} else {
		fmt.Printf("⚠️ Warning: Failed to check physical memory: %v\n", err)
	}

	return nil
}
