// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBuildVenvEnv(t *testing.T) {
	venvDir := filepath.Join("tmp", "unistack", "venv")

	// Determine the expected binary path
	expectedBinDir := filepath.Join(venvDir, "bin")
	if runtime.GOOS == "windows" {
		expectedBinDir = filepath.Join(venvDir, "Scripts")
	}

	env := buildVenvEnv(venvDir)

	// Verify VIRTUAL_ENV is present
	foundVirtualEnv := false
	for _, e := range env {
		if e == "VIRTUAL_ENV="+venvDir {
			foundVirtualEnv = true
			break
		}
	}
	if !foundVirtualEnv {
		t.Errorf("VIRTUAL_ENV was not set correctly in the resulting env slice")
	}

	// Verify PATH has been manipulated to include the venv bin directory first
	foundPath := false
	for _, e := range env {
		if strings.HasPrefix(strings.ToUpper(e), "PATH=") {
			foundPath = true
			// The bin directory should be right after "PATH="
			pathValue := e[5:]
			if !strings.HasPrefix(pathValue, expectedBinDir) {
				t.Errorf("PATH does not start with the venv bin directory. Got: %s", pathValue)
			}
			break
		}
	}

	if !foundPath {
		t.Errorf("PATH was not found in the resulting env slice")
	}
}

func TestEnsurePythonInstalled(t *testing.T) {
	ctx := context.Background()

	// 1. Test fast path: python3 exists
	tempDirFast := t.TempDir()
	var pyName string
	if runtime.GOOS == "windows" {
		pyName = "python3.bat"
	} else {
		pyName = "python3"
	}
	
	fakePy3 := filepath.Join(tempDirFast, pyName)
	createFakeExecutable(t, fakePy3, 0, "")
	
	// Create an isolated environment where ONLY our fake python3 exists in PATH
	t.Setenv("PATH", tempDirFast)
	
	cmd, err := EnsurePythonInstalled(ctx)
	if err != nil {
		t.Fatalf("Expected fast path to succeed, got %v", err)
	}
	// The function returns just the command name "python3"
	if cmd != "python3" && cmd != "python" {
		t.Fatalf("Expected python3, got %s", cmd)
	}

	// 2. Test auto-installation path (POSIX only, Windows will fail early)
	if runtime.GOOS != "windows" {
		tempDirAuto := t.TempDir()
		// No python in PATH, but we provide a fake package manager
		// And a fake `sh` to successfully pretend installation happened
		fakeApt := filepath.Join(tempDirAuto, "apt-get")
		createFakeExecutable(t, fakeApt, 0, "")

		// Provide a fake sudo so the root check doesn't fail
		fakeSudo := filepath.Join(tempDirAuto, "sudo")
		createFakeExecutable(t, fakeSudo, 0, "")

		fakeSh := filepath.Join(tempDirAuto, "sh")
		// The fake sh will act as the installer, and then it MUST make `python3` available in PATH
		// so the subsequent check `exec.LookPath("python3")` passes.
		// Our fake sh will literally touch "python3" in the tempDirAuto and make it executable.
		shScript := `#!/bin/sh
echo "#!/bin/sh" > ` + filepath.Join(tempDirAuto, "python3") + `
echo "exit 0" >> ` + filepath.Join(tempDirAuto, "python3") + `
/bin/chmod +x ` + filepath.Join(tempDirAuto, "python3") + `
exit 0
`
		os.WriteFile(fakeSh, []byte(shScript), 0755)

		t.Setenv("PATH", tempDirAuto)
		
		// Run EnsurePythonInstalled
		// It won't find python3 initially, will find apt-get, will run `sh -c ...`
		// which will create python3 in tempDirAuto (which is in PATH).
		cmd, err = EnsurePythonInstalled(ctx)
		if err != nil {
			t.Fatalf("Expected auto-installation to succeed, got %v", err)
		}
		if cmd != "python3" {
			t.Fatalf("Expected python3 after installation, got %s", cmd)
		}
	}
}

func TestSetupVirtualEnvironment(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	// Mock python executable
	var pyName string
	if runtime.GOOS == "windows" {
		pyName = "python.bat"
	} else {
		pyName = "python"
	}
	fakePy := filepath.Join(tempDir, pyName)
	createFakeExecutable(t, fakePy, 0, "")

	venvDir := filepath.Join(tempDir, "myvenv")

	// Prepend PATH so that the exec.CommandContext(pythonCmd, ...) finds our fake python
	prependPath(t, tempDir)

	envVars, err := SetupVirtualEnvironment(ctx, pyName, venvDir)
	if err != nil {
		t.Fatalf("Expected SetupVirtualEnvironment to succeed, got %v", err)
	}
	if len(envVars) == 0 {
		t.Fatalf("Expected envVars to be returned")
	}
}
