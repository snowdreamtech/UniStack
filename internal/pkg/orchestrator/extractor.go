// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/snowdreamtech/unistack/ansible"
	"github.com/snowdreamtech/unistack/internal/pkg/env"
)

type manifest struct {
	Version string            `json:"version"`
	Files   map[string]string `json:"files"`
}

// extractAnsibleFS extracts the embedded Ansible files to the root data directory idempotently.
// It uses an atomic directory swap pattern to ensure extraction is never left in a half-finished state.
func extractAnsibleFS() (string, error) {
	rootDir := env.GetDataDir()
	ansibleDir := filepath.Join(rootDir, "ansible")
	oldDir := filepath.Join(rootDir, "ansible.old")

	// 0. Crash Recovery: Check if previous swap was interrupted
	if _, errOld := os.Stat(oldDir); errOld == nil {
		if _, errNew := os.Stat(ansibleDir); os.IsNotExist(errNew) {
			slog.Debug("⚠️ Recovering from incomplete previous extraction swap...")
			_ = os.Rename(oldDir, ansibleDir)
		}
	}

	// 1. Get the unique identifier for this binary build
	embeddedHash, err := getVersionID()
	if err != nil {
		return "", fmt.Errorf("failed to get binary version id: %w", err)
	}

	// 2. Check if the existing directory already has this exact version and files are untampered
	manifestFile := filepath.Join(ansibleDir, ".unistack_manifest.json")
	if manifestData, err := os.ReadFile(manifestFile); err == nil {
		var m manifest
		if err := json.Unmarshal(manifestData, &m); err == nil {
			if m.Version == embeddedHash {
				// Fast path: Verify all files
				tampered := false
				for path, expectedHash := range m.Files {
					fullPath := filepath.Join(ansibleDir, path)
					data, err := os.ReadFile(fullPath)
					if err != nil {
						tampered = true
						break
					}
					h := sha256.Sum256(data)
					if hex.EncodeToString(h[:]) != expectedHash {
						tampered = true
						break
					}
				}
				if !tampered {
					return ansibleDir, nil
				}
				slog.Debug("⚠️ Tampering or file corruption detected. Forcing re-extraction...")
			}
		}
	}

	// 3. We need to extract. Create a unique temporary directory to avoid conflicts.
	tmpDir := filepath.Join(rootDir, fmt.Sprintf("ansible.tmp.%d", time.Now().UnixNano()))
	if err := os.MkdirAll(tmpDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create temporary extraction directory: %w", err)
	}
	os.Chmod(tmpDir, 0700)

	extractionSuccessful := false
	defer func() {
		if !extractionSuccessful {
			_ = os.RemoveAll(tmpDir)
		}
	}()

	fileHashes := make(map[string]string)

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
			if err := os.MkdirAll(targetPath, 0700); err != nil {
				return err
			}
			os.Chmod(targetPath, 0700)

			// Fsync directory to ensure metadata is flushed
			if dirF, err := os.Open(targetPath); err == nil {
				dirF.Sync()
				dirF.Close()
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		data, err := fs.ReadFile(ansible.FS, path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		h := sha256.Sum256(data)
		fileHashes[path] = hex.EncodeToString(h[:])

		perm := info.Mode().Perm() & 0700
		if perm == 0 {
			perm = 0600
		}
		if len(data) >= 2 && data[0] == '#' && data[1] == '!' {
			perm |= 0100
		}

		// Military-grade write with Fsync
		f, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", targetPath, err)
		}
		if _, err := f.Write(data); err != nil {
			f.Close()
			return err
		}
		if err := f.Sync(); err != nil {
			f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		os.Chmod(targetPath, perm)

		// Sync parent directory
		if parentF, err := os.Open(filepath.Dir(targetPath)); err == nil {
			parentF.Sync()
			parentF.Close()
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to extract ansible files: %w", err)
	}

	// 5. Write the manifest file
	m := manifest{
		Version: embeddedHash,
		Files:   fileHashes,
	}
	mData, _ := json.MarshalIndent(m, "", "  ")
	manifestPath := filepath.Join(tmpDir, ".unistack_manifest.json")
	if f, err := os.OpenFile(manifestPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600); err == nil {
		f.Write(mData)
		f.Sync()
		f.Close()
	}
	os.Chmod(manifestPath, 0600)

	// Final sync on the tmp root directory
	if tmpRootF, err := os.Open(tmpDir); err == nil {
		tmpRootF.Sync()
		tmpRootF.Close()
	}

	// 6. Atomically swap the directories
	_ = os.RemoveAll(oldDir)
	if matches, err := filepath.Glob(filepath.Join(rootDir, "ansible.tmp.*")); err == nil {
		for _, m := range matches {
			if m != tmpDir {
				_ = os.RemoveAll(m)
			}
		}
	}

	if _, err := os.Stat(ansibleDir); err == nil {
		if err := os.Rename(ansibleDir, oldDir); err != nil {
			return "", fmt.Errorf("failed to move existing ansible directory: %w", err)
		}
	}

	if err := os.Rename(tmpDir, ansibleDir); err != nil {
		_ = os.Rename(oldDir, ansibleDir)
		return "", fmt.Errorf("failed to atomically activate new ansible directory: %w", err)
	}

	_ = os.RemoveAll(oldDir)
	extractionSuccessful = true
	return ansibleDir, nil
}

// getVersionID generates a unique identifier for the currently running binary.
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
	return fmt.Sprintf("%d-%d", info.Size(), info.ModTime().UnixNano()), nil
}
