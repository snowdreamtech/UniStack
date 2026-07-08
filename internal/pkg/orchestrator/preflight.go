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

	// 2. Disk Space Check: Require at least 1GB (1024 * 1024 * 1024 bytes) of free space
	freeSpace, err := getFreeDiskSpace(filepath.Dir(rootDir))
	if err == nil {
		if freeSpace < 1024*1024*1024 {
			return fmt.Errorf("🚨 FATAL: Insufficient disk space. Required: 1024 MB (1GB), Available: %d MB. Boot aborted to prevent corruption", freeSpace/(1024*1024))
		}
	} else {
		fmt.Printf("⚠️ Warning: Failed to check disk space: %v\n", err)
	}

	// 3. Physical Memory Check: Require at least 1GB (1024 * 1024 * 1024 bytes)
	totalMemory, err := getTotalMemory()
	if err == nil {
		if totalMemory < 1024*1024*1024 {
			return fmt.Errorf("🚨 FATAL: Insufficient physical memory (%d MB). Ansible execution requires at least 1024 MB (1GB) to prevent OOM crashes", totalMemory/(1024*1024))
		}
	} else {
		fmt.Printf("⚠️ Warning: Failed to check physical memory: %v\n", err)
	}

	return nil
}
