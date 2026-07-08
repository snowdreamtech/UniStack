// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ExecutePlaybook runs ansible-playbook from the given working directory.
func ExecutePlaybook(workDir, playbook string, inventory string) error {
	// Ensure ansible-playbook is installed on the system
	binary, err := exec.LookPath("ansible-playbook")
	if err != nil {
		return fmt.Errorf("ansible-playbook not found in PATH, is ansible installed? %w", err)
	}

	// Prepare the command
	cmd := exec.Command(binary, "-i", inventory, playbook)
	cmd.Dir = workDir
	
	// Stream standard output and error directly to the console
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Optional: Setting environment variables (like ANSIBLE_CONFIG)
	// Because ansible.cfg is in the workDir, Ansible usually picks it up automatically.
	// But we can be explicit.
	env := os.Environ()
	env = append(env, fmt.Sprintf("ANSIBLE_CONFIG=%s", filepath.Join(workDir, "ansible.cfg")))
	cmd.Env = env

	fmt.Printf("🚀 Executing: %s -i %s %s in %s\n", binary, inventory, playbook, workDir)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("playbook execution failed: %w", err)
	}

	return nil
}
