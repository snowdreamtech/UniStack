// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPrepareEnvironmentAndExecution(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	// Redirect data dir to avoid touching user system
	t.Setenv("UNISTACK_DATA_DIR", tempDir)

	// Set up our isolated fake binaries
	fakeBinDir := filepath.Join(tempDir, "fake_bin")
	os.MkdirAll(fakeBinDir, 0755)

	var pyName, playbookName, ansibleName string
	if runtime.GOOS == "windows" {
		pyName = "python3.bat"
		playbookName = "ansible-playbook.bat"
		ansibleName = "ansible.bat"
	} else {
		pyName = "python3"
		playbookName = "ansible-playbook"
		ansibleName = "ansible"
	}

	createFakeExecutable(t, filepath.Join(fakeBinDir, pyName), 0, "Python 3.12.0")
	createFakeExecutable(t, filepath.Join(fakeBinDir, playbookName), 0, "ansible-playbook 2.15.0")
	createFakeExecutable(t, filepath.Join(fakeBinDir, ansibleName), 0, "ansible 2.15.0")

	// In Windows EnsurePythonInstalled returns an error if neither python3 nor python is found
	if runtime.GOOS == "windows" {
		createFakeExecutable(t, filepath.Join(fakeBinDir, "python.bat"), 0, "Python 3.12.0")
	} else {
		createFakeExecutable(t, filepath.Join(fakeBinDir, "python"), 0, "Python 3.12.0")
	}

	prependPath(t, fakeBinDir)

	// 1. Test PrepareEnvironment
	workDir, binary, venvEnv, err := PrepareEnvironment(ctx, "")
	if err != nil {
		t.Fatalf("PrepareEnvironment failed: %v", err)
	}

	if workDir == "" {
		t.Fatalf("Expected workDir to be returned")
	}
	if binary == "" {
		t.Fatalf("Expected binary to be returned")
	}
	// 2. Test ExecutePlaybook
	playbookFile := filepath.Join(tempDir, "test_playbook.yml")
	os.WriteFile(playbookFile, []byte("- hosts: all\n"), 0644)

	err = ExecutePlaybook(workDir, playbookFile, "", binary, venvEnv)
	if err != nil {
		t.Fatalf("ExecutePlaybook failed: %v", err)
	}

	// 3. Test ExecuteAdHoc
	err = ExecuteAdHoc(workDir, "all", binary, venvEnv, "-m", "ping")
	if err != nil {
		t.Fatalf("ExecuteAdHoc failed: %v", err)
	}
}

func TestPrepareEnvironmentFailures(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	t.Setenv("UNISTACK_DATA_DIR", tempDir)

	// Test PrepareEnvironment with NO python
	t.Setenv("PATH", t.TempDir()) // empty PATH
	_, _, _, err := PrepareEnvironment(ctx, "")
	if err == nil {
		t.Fatalf("Expected PrepareEnvironment to fail without Python, got nil")
	}
}
