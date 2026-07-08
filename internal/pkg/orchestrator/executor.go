// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ensureAnsibleInstalled checks if ansible-playbook exists. If not, it creates a venv and installs it.
func ensureAnsibleInstalled(workDir, pipIndexUrl string) (string, []string, error) {
	// 1. Check global PATH
	if binary, err := exec.LookPath("ansible-playbook"); err == nil {
		return binary, nil, nil
	}

	// 2. Define venv paths
	// workDir is ~/.local/share/unistack/ansible, so venv is ~/.local/share/unistack/.venv
	baseDir := filepath.Dir(workDir)
	venvDir := filepath.Join(baseDir, ".venv")
	venvBin := filepath.Join(venvDir, "bin", "ansible-playbook")

	// 3. Check if already installed in venv
	if _, err := os.Stat(venvBin); err == nil {
		return venvBin, buildVenvEnv(venvDir), nil
	}

	// 4. Need to bootstrap. Check for python3.
	pythonBin, err := exec.LookPath("python3")
	if err != nil {
		return "", nil, fmt.Errorf("python3 is required to bootstrap Ansible but was not found in PATH")
	}

	fmt.Println("🚀 Bootstrapping Python Virtual Environment for Ansible...")

	// Create Venv
	cmd := exec.Command(pythonBin, "-m", "venv", venvDir)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return "", nil, fmt.Errorf("failed to create virtual environment: %w", err)
	}

	pipBin := filepath.Join(venvDir, "bin", "pip")

	// Set Pip Index URL if provided
	if pipIndexUrl != "" {
		fmt.Printf("📦 Configuring pip mirror: %s\n", pipIndexUrl)
		cmd = exec.Command(pipBin, "config", "set", "global.index-url", pipIndexUrl)
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️ Warning: Failed to set pip index-url: %v\n", err)
		}
	}

	// Install requirements
	reqFile := filepath.Join(workDir, "requirements.txt")
	fmt.Println("📦 Installing Ansible dependencies...")
	cmd = exec.Command(pipBin, "install", "-r", reqFile)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return "", nil, fmt.Errorf("failed to install requirements via pip: %w", err)
	}

	if _, err := os.Stat(venvBin); err != nil {
		return "", nil, fmt.Errorf("ansible-playbook not found in venv after installation")
	}

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
func ExecutePlaybook(workDir, playbook string, inventory string, pipIndexUrl string) error {
	binary, venvEnv, err := ensureAnsibleInstalled(workDir, pipIndexUrl)
	if err != nil {
		return fmt.Errorf("ansible-playbook setup failed: %w", err)
	}

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
