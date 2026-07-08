// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/flock"
	"github.com/snowdreamtech/unistack/ansible"
	"github.com/snowdreamtech/unistack/internal/pkg/env"
)

// extractAnsibleFS extracts the embedded Ansible files to the root data directory idempotently.
// It uses an atomic directory swap pattern to ensure extraction is never left in a half-finished state.
func extractAnsibleFS() (string, error) {
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
	if err := os.MkdirAll(tmpDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create temporary extraction directory: %w", err)
	}
	// Bypass any restrictive host umask to guarantee we can write into our own temp dir
	os.Chmod(tmpDir, 0700)

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
			if err := os.MkdirAll(targetPath, 0700); err != nil {
				return err
			}
			os.Chmod(targetPath, 0700) // Immunity against host umask
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

		// Preserve executable bits (critical for dynamic inventories) but lock down permissions to the current user (0700 mask)
		perm := info.Mode().Perm() & 0700
		if perm == 0 {
			perm = 0600 // Default to readable/writable by owner if no perms
		}

		// Windows compilation destroys POSIX executable bits in embed.FS.
		// Auto-detect shebang to retroactively repair executable permissions!
		if len(data) >= 2 && data[0] == '#' && data[1] == '!' {
			perm |= 0100 // Add owner execute bit (u+x)
		}

		if err := os.WriteFile(targetPath, data, perm); err != nil {
			return fmt.Errorf("failed to write file %s: %w", targetPath, err)
		}
		
		// Force permissions strictly to bypass any overly restrictive host umask
		os.Chmod(targetPath, perm)

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to extract ansible files: %w", err)
	}

	// 5. Write the hash file to indicate completion inside the temporary directory
	hashFilePath := filepath.Join(tmpDir, ".unistack_hash")
	if err := os.WriteFile(hashFilePath, []byte(embeddedHash), 0600); err != nil {
		return "", fmt.Errorf("failed to write hash file: %w", err)
	}
	os.Chmod(hashFilePath, 0600) // Guarantee readability in future runs regardless of umask

	// 6. Atomically swap the directories
	// To safely swap without relying purely on POSIX rename-over-nonempty semantics:
	oldDir := filepath.Join(rootDir, "ansible.old")
	_ = os.RemoveAll(oldDir) // Ensure old backup is gone

	// Clean up any stale tmp dirs from previous crashed runs (e.g. killed during extraction)
	if matches, err := filepath.Glob(filepath.Join(rootDir, "ansible.tmp.*")); err == nil {
		for _, m := range matches {
			if m != tmpDir {
				_ = os.RemoveAll(m)
			}
		}
	}

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

// PrepareEnvironment sets up the data directory, safely extracts files, and ensures
// the python virtual environment is perfectly bootstrapped. It uses a unified global lock
// to protect both extraction and pip bootstrapping against concurrent executions.
func PrepareEnvironment(ctx context.Context, pipIndexUrl string) (string, string, []string, error) {
	rootDir := env.GetDataDir()
	
	// Ensure root directory exists before attempting to create the lock file, locked down to owner only
	if err := os.MkdirAll(rootDir, 0700); err != nil {
		return "", "", nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	os.Chmod(rootDir, 0700) // Retroactively secure if it already existed with loose permissions

	// Pre-create the logs directory so Ansible doesn't complain that it can't write to ansible.log
	ansibleDir := filepath.Join(rootDir, ".ansible")
	logsDir := filepath.Join(ansibleDir, "logs")
	if err := os.MkdirAll(logsDir, 0700); err != nil {
		return "", "", nil, fmt.Errorf("failed to create logs directory: %w", err)
	}
	// Explicitly tighten permissions on the entire Ansible state namespace
	os.Chmod(ansibleDir, 0700)
	os.Chmod(logsDir, 0700)

	lockFile := filepath.Join(rootDir, ".init.lock")
	// Pre-create the lock file with strict permissions (0600) to prevent cross-user DoS attacks
	if f, err := os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0600); err == nil {
		f.Close()
	}
	os.Chmod(lockFile, 0600) // Immunity against umask removing read/write bits
	fileLock := flock.New(lockFile)

	// 1. Acquire OS-level file lock
	lockAcquired := false
	for i := 0; i < 60; i++ {
		locked, err := fileLock.TryLock()
		if err != nil {
			return "", "", nil, fmt.Errorf("failed to attempt global lock: %w", err)
		}
		if locked {
			lockAcquired = true
			break
		}
		if i == 0 {
			fmt.Println("⏳ Another UniStack process is initializing. Waiting for global lock...")
		}
		
		select {
		case <-ctx.Done():
			return "", "", nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	if !lockAcquired {
		return "", "", nil, fmt.Errorf("timeout waiting for global lock at %s", lockFile)
	}
	defer fileLock.Unlock()

	// 2. Safely extract files (protected by global lock)
	workDir, err := extractAnsibleFS()
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to extract embedded files: %w", err)
	}

	// 3. Bootstrapping dependencies (protected by global lock)
	binary, venvEnv, err := ensureAnsibleInstalled(workDir, pipIndexUrl)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to ensure dependencies: %w", err)
	}

	return workDir, binary, venvEnv, nil
}
