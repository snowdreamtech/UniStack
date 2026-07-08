// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
	"github.com/snowdreamtech/unistack/internal/pkg/env"
)

// PrepareEnvironment is the unified entry point that prepares the entire execution environment.
// It handles:
// 1. Concurrency locking to prevent race conditions
// 2. Unpacking embedded files to disk
// 3. Detecting and deploying Python (if missing)
// 4. Detecting and deploying Ansible
func PrepareEnvironment(ctx context.Context, pipIndexUrl string) (string, string, []string, error) {
	rootDir := env.GetDataDir()

	// Ensure root directory exists before attempting to create the lock file, locked down to owner only
	if err := os.MkdirAll(rootDir, 0700); err != nil {
		return "", "", nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	if err := os.Chmod(rootDir, 0700); err != nil {
		return "", "", nil, fmt.Errorf("SECURITY ABORT: cannot secure root data directory %s (possible squatting attack): %w", rootDir, err)
	}

	// Pre-create the logs directory so Ansible doesn't complain that it can't write to ansible.log
	ansibleDir := filepath.Join(rootDir, ".ansible")
	logsDir := filepath.Join(ansibleDir, "logs")
	if err := os.MkdirAll(logsDir, 0700); err != nil {
		return "", "", nil, fmt.Errorf("failed to create logs directory: %w", err)
	}
	// Explicitly tighten permissions on the entire Ansible state namespace
	if err := os.Chmod(ansibleDir, 0700); err != nil {
		return "", "", nil, fmt.Errorf("SECURITY ABORT: cannot secure .ansible directory: %w", err)
	}
	if err := os.Chmod(logsDir, 0700); err != nil {
		return "", "", nil, fmt.Errorf("SECURITY ABORT: cannot secure logs directory: %w", err)
	}

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
			slog.Debug("⏳ Another UniStack process is initializing. Waiting for global lock...")
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

	// Pre-flight Systems Check (OS, Disk, Memory)
	if err := RunPreflightChecks(rootDir); err != nil {
		return "", "", nil, err
	}

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

	// 4. Post-flight Health Checks
	// Extract python binary path from environment
	pythonBin := filepath.Join(workDir, ".venv", "bin", "python") // fallback
	// To be perfectly precise, we should resolve pythonBin based on venvEnv or assume standard layout.
	// Since ensureAnsibleInstalled guarantees standard layout (workDir/.ansible/venv/bin/python),
	// we will pass the exact paths.

	// Wait, ensureAnsibleInstalled returns the ansible binary path, we can derive python from it
	pythonBin = filepath.Join(filepath.Dir(binary), "python")

	if err := RunPostflightChecks(ctx, venvEnv, workDir, pythonBin, binary); err != nil {
		return "", "", nil, err
	}

	return workDir, binary, venvEnv, nil
}

// ExecutePlaybook is the unified entry point to run the prepared Ansible environment (ansible-playbook).
func ExecutePlaybook(workDir, playbook, inventory, binary string, venvEnv []string, extraArgs ...string) error {
	var args []string
	if inventory != "" {
		args = append(args, "-i", inventory)
	}
	args = append(args, playbook)
	args = append(args, extraArgs...)
	return runAnsibleCommand("ansible-playbook", workDir, binary, venvEnv, args...)
}

// ExecuteAdHoc is the unified entry point to run the prepared Ansible environment (ansible).
func ExecuteAdHoc(workDir, pattern string, binary string, venvEnv []string, extraArgs ...string) error {
	args := []string{pattern}
	args = append(args, extraArgs...)
	return runAnsibleCommand("ansible", workDir, binary, venvEnv, args...)
}

// runAnsibleCommand handles the shared logic of invoking the Ansible environment.
func runAnsibleCommand(binaryName, workDir, binary string, venvEnv []string, args ...string) error {
	// Derive the target binary (ansible vs ansible-playbook) from the known binary path
	cmdPath := filepath.Join(filepath.Dir(binary), binaryName)

	// In some Edge cases on Windows, we might need .exe suffix
	if filepath.Ext(binary) == ".exe" && filepath.Ext(cmdPath) == "" {
		cmdPath += ".exe"
	}

	cmd := exec.Command(cmdPath, args...)
	cmd.Dir, _ = os.Getwd()

	// Stream standard input, output and error directly to the console
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start with the venv environment if we are using the venv, otherwise system environment
	var envVars []string
	if len(venvEnv) > 0 {
		envVars = venvEnv
	} else {
		envVars = os.Environ()
	}

	// Set ANSIBLE_CONFIG
	envVars = append(envVars, fmt.Sprintf("ANSIBLE_CONFIG=%s", filepath.Join(workDir, "ansible.cfg")))
	cmd.Env = envVars

	// Provide a masked representation of the command for logging
	slog.Debug(fmt.Sprintf("🚀 Executing: %s %v in %s\n", cmdPath, args, workDir))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s execution failed: %w", binaryName, err)
	}

	return nil
}
