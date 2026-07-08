// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateDependenciesHash(t *testing.T) {
	// Create a temporary directory to act as the workDir
	tempDir, err := os.MkdirTemp("", "unistack_ansible_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Write mock requirements.txt
	reqTxtContent := "ansible==9.1.0\n"
	err = os.WriteFile(filepath.Join(tempDir, "requirements.txt"), []byte(reqTxtContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write requirements.txt: %v", err)
	}

	// Write mock requirements.yml
	reqYmlContent := "- src: some_role\n  version: 1.0.0\n"
	err = os.WriteFile(filepath.Join(tempDir, "requirements.yml"), []byte(reqYmlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write requirements.yml: %v", err)
	}

	// Calculate hash manually to compare
	hash := sha256.New()
	hash.Write([]byte(reqTxtContent))
	hash.Write([]byte(reqYmlContent))
	expectedHash := hex.EncodeToString(hash.Sum(nil))

	// Call the function
	actualHash, err := calculateDependenciesHash(tempDir)
	if err != nil {
		t.Fatalf("calculateDependenciesHash returned error: %v", err)
	}

	if actualHash != expectedHash {
		t.Errorf("Expected hash %s, got %s", expectedHash, actualHash)
	}

	// Test missing files (should still return a valid empty hash without errors)
	emptyDir, _ := os.MkdirTemp("", "unistack_ansible_test_empty")
	defer os.RemoveAll(emptyDir)

	emptyHash, err := calculateDependenciesHash(emptyDir)
	if err != nil {
		t.Fatalf("calculateDependenciesHash returned error on empty dir: %v", err)
	}

	hashEmpty := sha256.New()
	expectedEmptyHash := hex.EncodeToString(hashEmpty.Sum(nil))

	if emptyHash != expectedEmptyHash {
		t.Errorf("Expected empty hash %s, got %s", expectedEmptyHash, emptyHash)
	}
}
