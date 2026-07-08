// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"os"
	"testing"
)

func TestGetFreeDiskSpace(t *testing.T) {
	// Use the current directory or temp directory
	tempDir := t.TempDir()

	space, err := getFreeDiskSpace(tempDir)
	if err != nil {
		t.Fatalf("getFreeDiskSpace failed: %v", err)
	}

	// We expect *some* free space to exist on the disk running tests
	if space == 0 {
		t.Errorf("Expected free space > 0, got 0")
	}

	// Test with a non-existent path
	_, err = getFreeDiskSpace(tempDir + string(os.PathSeparator) + "nonexistent_directory_for_test")
	if err == nil {
		t.Errorf("Expected getFreeDiskSpace to fail for a non-existent directory")
	}
}

func TestGetTotalMemory(t *testing.T) {
	mem, err := getTotalMemory()
	if err != nil {
		t.Fatalf("getTotalMemory failed: %v", err)
	}

	// We expect *some* memory to exist on the machine running tests
	if mem == 0 {
		t.Errorf("Expected total memory > 0, got 0")
	}
}
