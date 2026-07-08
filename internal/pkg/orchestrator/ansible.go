// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/snowdreamtech/unistack/internal/pkg/env"
)

// ensureAnsibleInstalled checks for ansible and installs it in a venv if missing
func ensureAnsibleInstalled(workDir string, pipIndexUrl string) (string, []string, error) {
	// First check if ansible-playbook is already in the system PATH
	sysBin, err := exec.LookPath("ansible-playbook")
	if err == nil {
		slog.Debug(fmt.Sprintf("✅ Found system Ansible at %s\n", sysBin))
		return sysBin, nil, nil
	}

	// Paths for local venv - placed OUTSIDE workDir so it survives atomic file extractions
	// when the UniStack binary is upgraded but python dependencies remain unchanged.
	venvDir := filepath.Join(env.GetDataDir(), ".ansible", "venv")

	venvBinDir := filepath.Join(venvDir, "bin")
	if runtime.GOOS == "windows" {
		venvBinDir = filepath.Join(venvDir, "Scripts")
	}
	venvBin := filepath.Join(venvBinDir, "ansible-playbook.exe")
	if runtime.GOOS != "windows" {
		venvBin = filepath.Join(venvBinDir, "ansible-playbook")
	}
	markerFile := filepath.Join(venvDir, ".bootstrap_complete")

	// Calculate dependency hash to detect version upgrades
	currentHash, _ := calculateDependenciesHash(workDir)

	// If atomic marker exists, check hash and binary
	if markerData, err := os.ReadFile(markerFile); err == nil {
		if string(markerData) == currentHash {
			if _, err := os.Stat(venvBin); err == nil {
				return venvBin, buildVenvEnv(venvDir), nil
			}
		} else {
			slog.Debug("🔄 Dependencies have changed (binary upgrade detected). Rebuilding virtual environment...")
		}
	}

	// The global lock is now held by PrepareEnvironment, so we can proceed directly.

	// Double check marker after acquiring lock (not strictly needed now, but safe)
	if markerData, err := os.ReadFile(markerFile); err == nil {
		if string(markerData) == currentHash {
			if _, err := os.Stat(venvBin); err == nil {
				return venvBin, buildVenvEnv(venvDir), nil
			}
		}
	}

	// We are going to bootstrap. Remove incomplete venv if exists (Scorched Earth)
	os.RemoveAll(venvDir)
	os.Remove(markerFile)

	// Define robust execution closure with scorched earth on final failure
	bootstrapSuccess := false
	defer func() {
		if !bootstrapSuccess {
			slog.Debug("💥 Bootstrap failed or was interrupted. Executing scorched earth rollback...")
			os.RemoveAll(venvDir)
			os.Remove(markerFile)
		}
	}()

	// Global context for all network operations (10 minute timeout), wrapped in a signal trap for Ctrl+C
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(sigCtx, 10*time.Minute)
	defer cancel()

	// Delegate Python discovery and auto-installation to the python module
	pythonCmd, err := EnsurePythonInstalled(ctx)
	if err != nil {
		return "", nil, err
	}

	// Delegate venv creation to the python module
	venvEnv, err := SetupVirtualEnvironment(ctx, pythonCmd, venvDir)
	if err != nil {
		return "", nil, err
	}

	pipBin := filepath.Join(venvBinDir, "pip")

	if pipIndexUrl != "" {
		slog.Debug(fmt.Sprintf("📦 Configuring pip mirror: %s\n", pipIndexUrl))
		cmd := exec.CommandContext(ctx, pipBin, "config", "set", "global.index-url", pipIndexUrl)
		if err := cmd.Run(); err != nil {
			slog.Debug(fmt.Sprintf("⚠️ Warning: failed to set pip mirror: %v\n", err))
		}
	}

	// Helper function for command retry with context
	runWithRetry := func(name string, createCmd func(context.Context) *exec.Cmd, maxRetries int, delay time.Duration) error {
		var lastErr error
		for i := 0; i < maxRetries; i++ {
			if i > 0 {
				slog.Debug(fmt.Sprintf("⚠️ %s failed, retrying in %v (attempt %d/%d)...\n", name, delay, i+1, maxRetries))
				select {
				case <-time.After(delay):
				case <-ctx.Done():
					return fmt.Errorf("context timeout during %s", name)
				}
			}
			cmd := createCmd(ctx)
			cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
			lastErr = cmd.Run()
			if lastErr == nil {
				return nil
			}
		}
		return fmt.Errorf("%s failed after %d attempts: %w", name, maxRetries, lastErr)
	}

	// Install requirements via pip
	reqFile := filepath.Join(workDir, "requirements.txt")
	slog.Debug("📦 Installing Ansible dependencies via pip...")
	err = runWithRetry("pip install", func(c context.Context) *exec.Cmd {
		return exec.CommandContext(c, pipBin, "install", "-r", reqFile)
	}, 3, 3*time.Second)
	if err != nil {
		return "", nil, err
	}

	// Install Ansible Galaxy Collections and Roles
	galaxyReqFile := filepath.Join(workDir, "requirements.yml")
	if _, err := os.Stat(galaxyReqFile); err == nil {
		slog.Debug("🌌 Installing Ansible Galaxy Dependencies (Collections & Roles)...")
		galaxyBin := filepath.Join(venvBinDir, "ansible-galaxy")

		// Install Collections
		err = runWithRetry("ansible-galaxy collection install", func(c context.Context) *exec.Cmd {
			cCmd := exec.CommandContext(c, galaxyBin, "collection", "install", "-r", galaxyReqFile)
			cCmd.Dir = workDir
			env := venvEnv
			env = append(env, fmt.Sprintf("ANSIBLE_CONFIG=%s", filepath.Join(workDir, "ansible.cfg")))
			cCmd.Env = env
			return cCmd
		}, 3, 3*time.Second)
		if err != nil {
			return "", nil, err
		}

		// Install Roles (ignore errors if no roles are defined in requirements.yml)
		_ = runWithRetry("ansible-galaxy role install", func(c context.Context) *exec.Cmd {
			cCmd := exec.CommandContext(c, galaxyBin, "role", "install", "-r", galaxyReqFile)
			cCmd.Dir = workDir
			env := venvEnv
			env = append(env, fmt.Sprintf("ANSIBLE_CONFIG=%s", filepath.Join(workDir, "ansible.cfg")))
			cCmd.Env = env
			return cCmd
		}, 3, 3*time.Second)
	}

	// Successfully finished everything. Write atomic marker with the hash.
	if file, err := os.OpenFile(markerFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600); err == nil {
		file.WriteString(currentHash)
		file.Close()
	}
	os.Chmod(markerFile, 0600) // Immunity against umask stripping read permissions (which would break future fast paths)

	bootstrapSuccess = true
	return venvBin, venvEnv, nil
}

// calculateDependenciesHash computes a SHA-256 hash of the content of requirements.txt and requirements.yml.
// This allows us to detect when the binary is upgraded and dependencies change, triggering a fresh bootstrap.
func calculateDependenciesHash(workDir string) (string, error) {
	hash := sha256.New()

	reqFile := filepath.Join(workDir, "requirements.txt")
	reqData, err := os.ReadFile(reqFile)
	if err == nil {
		hash.Write(reqData)
	}

	galaxyReqFile := filepath.Join(workDir, "requirements.yml")
	galaxyReqData, err := os.ReadFile(galaxyReqFile)
	if err == nil {
		hash.Write(galaxyReqData)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
