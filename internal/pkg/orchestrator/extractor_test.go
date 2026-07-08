// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractAnsibleFS(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("UNISTACK_DATA_DIR", tempDir)

	// First extraction
	ansibleDir, err := extractAnsibleFS()
	if err != nil {
		t.Fatalf("extractAnsibleFS failed: %v", err)
	}

	// Verify directory exists
	if _, err := os.Stat(ansibleDir); os.IsNotExist(err) {
		t.Fatalf("Expected directory %s to be created", ansibleDir)
	}

	// Verify manifest exists
	manifestFile := filepath.Join(ansibleDir, ".unistack_manifest.json")
	if _, err := os.Stat(manifestFile); os.IsNotExist(err) {
		t.Fatalf("Expected manifest file %s to be created", manifestFile)
	}

	// Second extraction (should be fast path)
	ansibleDir2, err := extractAnsibleFS()
	if err != nil {
		t.Fatalf("extractAnsibleFS (fast path) failed: %v", err)
	}
	if ansibleDir2 != ansibleDir {
		t.Fatalf("Expected same directory on second extraction, got %s", ansibleDir2)
	}

	// Tamper with manifest to test force re-extraction
	// Corrupt a known file
	// Read manifest
	manifestData, err := os.ReadFile(manifestFile)
	if err != nil {
		t.Fatalf("Failed to read manifest: %v", err)
	}
	var m manifest
	if err := json.Unmarshal(manifestData, &m); err != nil {
		t.Fatalf("Failed to unmarshal manifest: %v", err)
	}

	// Find any file in the manifest to corrupt
	var fileToCorrupt string
	for k := range m.Files {
		fileToCorrupt = k
		break
	}

	if fileToCorrupt != "" {
		fullPath := filepath.Join(ansibleDir, fileToCorrupt)
		// Alter content
		os.WriteFile(fullPath, []byte("tampered content"), 0644)
		
		// Third extraction (should detect tampering and re-extract)
		ansibleDir3, err := extractAnsibleFS()
		if err != nil {
			t.Fatalf("extractAnsibleFS (tampered) failed: %v", err)
		}
		if ansibleDir3 != ansibleDir {
			t.Fatalf("Expected same directory on third extraction, got %s", ansibleDir3)
		}

		// Verify file was restored
		data, _ := os.ReadFile(fullPath)
		h := sha256.Sum256(data)
		expectedHash := m.Files[fileToCorrupt]
		if hex.EncodeToString(h[:]) != expectedHash {
			t.Fatalf("File was not restored correctly during tamper recovery")
		}
	}
}

func TestGetVersionID(t *testing.T) {
	id, err := getVersionID()
	if err != nil {
		t.Fatalf("getVersionID failed: %v", err)
	}
	if id == "" {
		t.Fatalf("getVersionID returned empty string")
	}
}
