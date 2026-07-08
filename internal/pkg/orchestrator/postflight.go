// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"path/filepath"
)

// RunPostflightChecks validates that the initialized environment (Python, Ansible)
// is functional and capable of executing payloads on the host system.
func RunPostflightChecks(ctx context.Context, venvEnv []string, workDir, pythonBin, ansibleBin string) error {
	fmt.Println("🚀 Running Post-flight System Checks...")

	// Create unified environment map for command execution
	var envVars []string
	if len(venvEnv) > 0 {
		envVars = venvEnv
	} else {
		envVars = os.Environ()
	}
	envVars = append(envVars, fmt.Sprintf("ANSIBLE_CONFIG=%s", filepath.Join(workDir, "ansible.cfg")))

	// 1. Python Health Check
	pyCmd := exec.CommandContext(ctx, pythonBin, "-V")
	pyCmd.Env = envVars
	pyOutput, err := pyCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("🚨 FATAL: Python health check failed. The environment is broken or incompatible with this architecture.\nError: %w\nOutput: %s", err, string(pyOutput))
	}
	pyVersion := strings.TrimSpace(string(pyOutput))
	fmt.Printf("✅ [Postflight] Python interpreter is healthy (%s)\n", pyVersion)

	// 2. Ansible Health Check
	ansibleCmd := exec.CommandContext(ctx, ansibleBin, "--version")
	ansibleCmd.Env = envVars
	ansibleOutput, err := ansibleCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("🚨 FATAL: Ansible health check failed. Modules may be missing or corrupted.\nError: %w\nOutput: %s", err, string(ansibleOutput))
	}
	
	// Just print the first line of ansible --version to keep it clean
	ansibleVersionLines := strings.Split(string(ansibleOutput), "\n")
	if len(ansibleVersionLines) > 0 {
		fmt.Printf("✅ [Postflight] Ansible core is healthy (%s)\n", strings.TrimSpace(ansibleVersionLines[0]))
	}

	return nil
}
