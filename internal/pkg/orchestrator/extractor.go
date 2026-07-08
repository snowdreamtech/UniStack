// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/snowdreamtech/unistack/ansible"
	"github.com/snowdreamtech/unistack/internal/pkg/env"
)

// ExtractAnsibleFS extracts the embedded Ansible files to the root data directory idempotently.
// It uses an atomic directory swap pattern to ensure extraction is never left in a half-finished state.
func ExtractAnsibleFS() (string, error) {
	rootDir := env.GetDataDir()
	ansibleDir := filepath.Join(rootDir, "ansible")

	// 1. Get the unique identifier for this binary build
	embeddedHash, err := getVersionID()
	if err != nil {
		return "", fmt.Errorf("failed to get binary version id: %w", err)
	}

	// 2. Check if the existing directory already has this exact version
	hashFile := filepath.Join(ansibleDir, ".unistack_hash")
	if existingHash, err := os.ReadFile(hashFile); err == nil {
		if strings.TrimSpace(string(existingHash)) == embeddedHash {
			// Fast path: Idempotent skip. Files are already fully extracted.
			return ansibleDir, nil
		}
	}

	// 3. We need to extract. Create a unique temporary directory to avoid conflicts.
	// This guarantees that if the process dies halfway, the main ansible/ directory is untouched.
	tmpDir := filepath.Join(rootDir, fmt.Sprintf("ansible.tmp.%d", time.Now().UnixNano()))
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temporary extraction directory: %w", err)
	}

	// Make sure we clean up the temporary directory if something fails during extraction
	extractionSuccessful := false
	defer func() {
		if !extractionSuccessful {
			_ = os.RemoveAll(tmpDir)
		}
	}()

	// 4. Extract all files into the temporary directory
	err = fs.WalkDir(ansible.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == "." {
			return nil
		}

		targetPath := filepath.Join(tmpDir, path)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		data, err := fs.ReadFile(ansible.FS, path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", targetPath, err)
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to extract ansible files: %w", err)
	}

	// 5. Write the hash file to indicate completion inside the temporary directory
	if err := os.WriteFile(filepath.Join(tmpDir, ".unistack_hash"), []byte(embeddedHash), 0644); err != nil {
		return "", fmt.Errorf("failed to write hash file: %w", err)
	}

	// 6. Atomically swap the directories
	// To safely swap without relying purely on POSIX rename-over-nonempty semantics:
	oldDir := filepath.Join(rootDir, "ansible.old")
	_ = os.RemoveAll(oldDir) // Ensure old backup is gone

	// If the current ansible dir exists, move it out of the way
	if _, err := os.Stat(ansibleDir); err == nil {
		if err := os.Rename(ansibleDir, oldDir); err != nil {
			return "", fmt.Errorf("failed to move existing ansible directory: %w", err)
		}
	}

	// Move the fully prepared temporary directory into the final location
	if err := os.Rename(tmpDir, ansibleDir); err != nil {
		// Attempt rollback if rename fails
		_ = os.Rename(oldDir, ansibleDir)
		return "", fmt.Errorf("failed to atomically activate new ansible directory: %w", err)
	}

	// Clean up the backup directory
	_ = os.RemoveAll(oldDir)

	extractionSuccessful = true
	return ansibleDir, nil
}

// getVersionID generates a unique identifier for the currently running binary.
// In production, it prefers the Git commit hash injected at build time.
// In local development, it instantly stats the executable file (Size + ModTime).
// This completely avoids the overhead of hashing embedded files on every run.
func getVersionID() (string, error) {
	if env.CommitHashFull != "" && env.CommitHashFull != "N/A" {
		return env.CommitHashFull, nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	info, err := os.Stat(exePath)
	if err != nil {
		return "", fmt.Errorf("failed to stat executable: %w", err)
	}

	// Format: size-timestamp
	return fmt.Sprintf("%d-%d", info.Size(), info.ModTime().UnixNano()), nil
}
