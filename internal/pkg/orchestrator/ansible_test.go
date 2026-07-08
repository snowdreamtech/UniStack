// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEnsureAnsibleInstalled(t *testing.T) {
	// Scenario 1: System ansible-playbook found
	tempDirFast := t.TempDir()
	var ansibleName string
	if runtime.GOOS == "windows" {
		ansibleName = "ansible-playbook.bat"
	} else {
		ansibleName = "ansible-playbook"
	}
	fakeAnsible := filepath.Join(tempDirFast, ansibleName)
	createFakeExecutable(t, fakeAnsible, 0, "")

	t.Setenv("PATH", tempDirFast)

	cmd, _, err := ensureAnsibleInstalled(tempDirFast, "")
	if err != nil {
		t.Fatalf("Expected fast path to succeed, got %v", err)
	}
	// On Windows, LookPath returns the full path with extension if in PATH, or just the executable name if found directly.
	// We'll just verify no error occurred and cmd is not empty.
	if cmd == "" {
		t.Fatalf("Expected ansible-playbook command path to be returned")
	}

	// Scenario 2: System missing, Venv cached (marker file matches hash)
	tempDataDir := t.TempDir()
	t.Setenv("UNISTACK_DATA_DIR", tempDataDir)
	
	// Ensure system PATH does not have ansible-playbook anymore
	t.Setenv("PATH", t.TempDir()) // empty PATH

	workDir := t.TempDir()
	hash, _ := calculateDependenciesHash(workDir)
	
	// Create marker file and venv binary
	venvDir := filepath.Join(tempDataDir, ".ansible", "venv")
	venvBinDir := filepath.Join(venvDir, "bin")
	if runtime.GOOS == "windows" {
		venvBinDir = filepath.Join(venvDir, "Scripts")
	}
	if err := os.MkdirAll(venvBinDir, 0755); err != nil {
		t.Fatalf("Failed to create venv bin dir: %v", err)
	}

	venvAnsible := filepath.Join(venvBinDir, ansibleName)
	if runtime.GOOS == "windows" && ansibleName == "ansible-playbook.bat" {
		// Actually the code specifically looks for "ansible-playbook.exe" on Windows, not .bat
		venvAnsible = filepath.Join(venvBinDir, "ansible-playbook.exe")
	}
	createFakeExecutable(t, venvAnsible, 0, "")

	markerFile := filepath.Join(venvDir, ".bootstrap_complete")
	os.WriteFile(markerFile, []byte(hash), 0644)

	cmd2, envVars, err2 := ensureAnsibleInstalled(workDir, "")
	if err2 != nil {
		t.Fatalf("Expected cached venv path to succeed, got %v", err2)
	}
	if cmd2 != venvAnsible {
		t.Fatalf("Expected venv ansible path %s, got %s", venvAnsible, cmd2)
	}
	if len(envVars) == 0 {
		t.Fatalf("Expected envVars to be returned for cached venv")
	}

	// Scenario 3: Missing marker file, triggers full pip installation
	os.Remove(markerFile)
	
	// Mock python, pip, ansible-galaxy in the PATH so they succeed
	tempDirInstall := t.TempDir()
	var pyName, pipName, galaxyName string
	if runtime.GOOS == "windows" {
		pyName = "python.bat"
		pipName = "pip.bat"
		galaxyName = "ansible-galaxy.bat"
	} else {
		pyName = "python"
		pipName = "pip"
		galaxyName = "ansible-galaxy"
	}

	// We need our fake python to ACTUALLY create the venv binaries when called as "python -m venv <venvDir>"
	// otherwise EnsureAnsibleInstalled will fail when trying to run the pip binary that should have been created!
	fakePythonPath := filepath.Join(tempDirInstall, pyName)
	var pyScript string
	if runtime.GOOS == "windows" {
		pyScript = `@echo off
if "%1"=="-m" if "%2"=="venv" (
	mkdir "%3\Scripts" 2>nul
	echo @echo off > "%3\Scripts\pip.exe"
	echo exit /b 0 >> "%3\Scripts\pip.exe"
	echo @echo off > "%3\Scripts\ansible-galaxy.exe"
	echo exit /b 0 >> "%3\Scripts\ansible-galaxy.exe"
	echo @echo off > "%3\Scripts\ansible-playbook.exe"
	echo exit /b 0 >> "%3\Scripts\ansible-playbook.exe"
)
exit /b 0
`
	} else {
		pyScript = `#!/bin/sh
if [ "$1" = "-m" ] && [ "$2" = "venv" ]; then
	/bin/mkdir -p "$3/bin"
	echo "#!/bin/sh" > "$3/bin/pip"
	echo "exit 0" >> "$3/bin/pip"
	/bin/chmod +x "$3/bin/pip"
	echo "#!/bin/sh" > "$3/bin/ansible-galaxy"
	echo "exit 0" >> "$3/bin/ansible-galaxy"
	/bin/chmod +x "$3/bin/ansible-galaxy"
	echo "#!/bin/sh" > "$3/bin/ansible-playbook"
	echo "exit 0" >> "$3/bin/ansible-playbook"
	/bin/chmod +x "$3/bin/ansible-playbook"
fi
exit 0
`
	}
	os.WriteFile(fakePythonPath, []byte(pyScript), 0755)

	createFakeExecutable(t, filepath.Join(tempDirInstall, pipName), 0, "")
	createFakeExecutable(t, filepath.Join(tempDirInstall, galaxyName), 0, "")

	// Prepend PATH
	prependPath(t, tempDirInstall)

	cmd3, _, err3 := ensureAnsibleInstalled(workDir, "")
	if err3 != nil {
		t.Fatalf("Expected full install path to succeed, got %v", err3)
	}
	if cmd3 != venvAnsible {
		t.Fatalf("Expected venv ansible path after install %s, got %s", venvAnsible, cmd3)
	}
}

func TestCalculateDependenciesHash(t *testing.T) {
	// Create a temporary directory to act as the workDir
	tempDir, err := os.MkdirTemp("", "unistack_ansible_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Write mock requirements.txt
	reqTxtContent := "ansible==9.1.0\n"
	err = os.WriteFile(filepath.Join(tempDir, "requirements.txt"), []byte(reqTxtContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write requirements.txt: %v", err)
	}

	// Write mock requirements.yml
	reqYmlContent := "- src: some_role\n  version: 1.0.0\n"
	err = os.WriteFile(filepath.Join(tempDir, "requirements.yml"), []byte(reqYmlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write requirements.yml: %v", err)
	}

	// Calculate hash manually to compare
	hash := sha256.New()
	hash.Write([]byte(reqTxtContent))
	hash.Write([]byte(reqYmlContent))
	expectedHash := hex.EncodeToString(hash.Sum(nil))

	// Call the function
	actualHash, err := calculateDependenciesHash(tempDir)
	if err != nil {
		t.Fatalf("calculateDependenciesHash returned error: %v", err)
	}

	if actualHash != expectedHash {
		t.Errorf("Expected hash %s, got %s", expectedHash, actualHash)
	}

	// Test missing files (should still return a valid empty hash without errors)
	emptyDir, _ := os.MkdirTemp("", "unistack_ansible_test_empty")
	defer os.RemoveAll(emptyDir)

	emptyHash, err := calculateDependenciesHash(emptyDir)
	if err != nil {
		t.Fatalf("calculateDependenciesHash returned error on empty dir: %v", err)
	}

	hashEmpty := sha256.New()
	expectedEmptyHash := hex.EncodeToString(hashEmpty.Sum(nil))

	if emptyHash != expectedEmptyHash {
		t.Errorf("Expected empty hash %s, got %s", expectedEmptyHash, emptyHash)
	}
}
