// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/snowdreamtech/unistack/internal/pkg/env"
)

// ensureAnsibleInstalled checks for ansible and installs it in a venv if missing
func ensureAnsibleInstalled(workDir string, pipIndexUrl string) (string, []string, error) {
	// First check if ansible-playbook is already in the system PATH
	sysBin, err := exec.LookPath("ansible-playbook")
	if err == nil {
		fmt.Printf("✅ Found system Ansible at %s\n", sysBin)
		return sysBin, nil, nil
	}

	// Paths for local venv - placed OUTSIDE workDir so it survives atomic file extractions
	// when the UniStack binary is upgraded but python dependencies remain unchanged.
	venvDir := filepath.Join(env.GetDataDir(), "venv")
	venvBin := filepath.Join(venvDir, "bin", "ansible-playbook")
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
			fmt.Println("🔄 Dependencies have changed (binary upgrade detected). Rebuilding virtual environment...")
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

	// Disk Pre-flight Check: Require at least 500MB (500 * 1024 * 1024 bytes) of free space
	freeSpace, err := getFreeDiskSpace(filepath.Dir(workDir))
	if err == nil && freeSpace < 500*1024*1024 {
		return "", nil, fmt.Errorf("🚨 FATAL: Insufficient disk space. Required: 500MB, Available: %d MB. Bootstrap aborted to prevent corruption", freeSpace/(1024*1024))
	} else if err != nil {
		fmt.Printf("⚠️ Warning: failed to check disk space: %v\n", err)
	}

	// We are going to bootstrap. Remove incomplete venv if exists (Scorched Earth)
	os.RemoveAll(venvDir)
	os.Remove(markerFile)

	// Define robust execution closure with scorched earth on final failure
	bootstrapSuccess := false
	defer func() {
		if !bootstrapSuccess {
			fmt.Println("💥 Bootstrap failed or was interrupted. Executing scorched earth rollback...")
			os.RemoveAll(venvDir)
			os.Remove(markerFile)
		}
	}()

	fmt.Println("🚀 Bootstrapping Python Virtual Environment for Ansible...")

	// Find python3
	pythonCmd, err := exec.LookPath("python3")
	if err != nil {
		return "", nil, fmt.Errorf("python3 not found in PATH, required for bootstrapping")
	}

	// Global context for all network operations (10 minute timeout), wrapped in a signal trap for Ctrl+C
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(sigCtx, 10*time.Minute)
	defer cancel()

	// Create venv
	cmd := exec.CommandContext(ctx, pythonCmd, "-m", "venv", venvDir)
	if err := cmd.Run(); err != nil {
		return "", nil, fmt.Errorf("failed to create venv: %w", err)
	}

	pipBin := filepath.Join(venvDir, "bin", "pip")

	// Set pip mirror if provided
	if pipIndexUrl != "" {
		fmt.Printf("📦 Configuring pip mirror: %s\n", pipIndexUrl)
		cmd = exec.CommandContext(ctx, pipBin, "config", "set", "global.index-url", pipIndexUrl)
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️ Warning: failed to set pip mirror: %v\n", err)
		}
	}

	// Helper function for command retry with context
	runWithRetry := func(name string, createCmd func(context.Context) *exec.Cmd, maxRetries int, delay time.Duration) error {
		var lastErr error
		for i := 0; i < maxRetries; i++ {
			if i > 0 {
				fmt.Printf("⚠️ %s failed, retrying in %v (attempt %d/%d)...\n", name, delay, i+1, maxRetries)
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
	fmt.Println("📦 Installing Ansible dependencies via pip...")
	err = runWithRetry("pip install", func(c context.Context) *exec.Cmd {
		return exec.CommandContext(c, pipBin, "install", "-r", reqFile)
	}, 3, 3*time.Second)
	if err != nil {
		return "", nil, err
	}

	// Install Ansible Galaxy Collections
	galaxyReqFile := filepath.Join(workDir, "requirements.yml")
	if _, err := os.Stat(galaxyReqFile); err == nil {
		fmt.Println("🌌 Installing Ansible Galaxy Collections...")
		galaxyBin := filepath.Join(venvDir, "bin", "ansible-galaxy")
		
		err = runWithRetry("ansible-galaxy install", func(c context.Context) *exec.Cmd {
			cCmd := exec.CommandContext(c, galaxyBin, "collection", "install", "-r", galaxyReqFile)
			cCmd.Dir = workDir
			env := buildVenvEnv(venvDir)
			env = append(env, fmt.Sprintf("ANSIBLE_CONFIG=%s", filepath.Join(workDir, "ansible.cfg")))
			cCmd.Env = env
			return cCmd
		}, 3, 3*time.Second)
		if err != nil {
			return "", nil, err
		}
	}

	// Successfully finished everything. Write atomic marker with the hash.
	if file, err := os.Create(markerFile); err == nil {
		file.WriteString(currentHash)
		file.Close()
	}

	bootstrapSuccess = true
	return venvBin, buildVenvEnv(venvDir), nil
}

// buildVenvEnv creates environment variables needed to run binaries inside a virtualenv
func buildVenvEnv(venvDir string) []string {
	env := os.Environ()
	pathIdx := -1
	for i, e := range env {
		if len(e) > 5 && e[:5] == "PATH=" {
			pathIdx = i
			break
		}
	}

	venvBinDir := filepath.Join(venvDir, "bin")
	if pathIdx != -1 {
		env[pathIdx] = fmt.Sprintf("PATH=%s:%s", venvBinDir, env[pathIdx][5:])
	} else {
		env = append(env, fmt.Sprintf("PATH=%s", venvBinDir))
	}
	env = append(env, fmt.Sprintf("VIRTUAL_ENV=%s", venvDir))
	return env
}

// ExecutePlaybook runs ansible-playbook from the given working directory.
func ExecutePlaybook(workDir, playbook, inventory, binary string, venvEnv []string) error {

	// Prepare the command
	cmd := exec.Command(binary, "-i", inventory, playbook)
	cmd.Dir = workDir
	
	// Stream standard output and error directly to the console
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start with the venv environment if we are using the venv, otherwise system environment
	var env []string
	if len(venvEnv) > 0 {
		env = venvEnv
	} else {
		env = os.Environ()
	}
	
	// Set ANSIBLE_CONFIG
	env = append(env, fmt.Sprintf("ANSIBLE_CONFIG=%s", filepath.Join(workDir, "ansible.cfg")))
	cmd.Env = env

	fmt.Printf("🚀 Executing: %s -i %s %s in %s\n", binary, inventory, playbook, workDir)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("playbook execution failed: %w", err)
	}

	return nil
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
