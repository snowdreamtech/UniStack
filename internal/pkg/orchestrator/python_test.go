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
		
		// Setup fake apt-get
		fakeApt := filepath.Join(tempDirAuto, "apt-get")
		createFakeExecutable(t, fakeApt, 0, "")

		// Setup fake sudo
		fakeSudo := filepath.Join(tempDirAuto, "sudo")
		createFakeExecutable(t, fakeSudo, 0, "")

		// Setup fake sh (success case)
		fakeSh := filepath.Join(tempDirAuto, "sh")
		shScript := `#!/bin/sh
echo "#!/bin/sh" > ` + filepath.Join(tempDirAuto, "python3") + `
echo "exit 0" >> ` + filepath.Join(tempDirAuto, "python3") + `
/bin/chmod +x ` + filepath.Join(tempDirAuto, "python3") + `
exit 0
`
		os.WriteFile(fakeSh, []byte(shScript), 0755)
		
		t.Setenv("PATH", tempDirAuto)
		
		cmd, err = EnsurePythonInstalled(ctx)
		if err != nil {
			t.Fatalf("Expected auto-installation to succeed, got %v", err)
		}
		if cmd != "python3" {
			t.Fatalf("Expected python3 after installation, got %s", cmd)
		}

		// 3. Test missing sudo
		os.Remove(fakeSudo)
		os.Remove(filepath.Join(tempDirAuto, "python3"))
		// We expect EnsurePythonInstalled to fail because it requires sudo
		// (unless the test is run as root, but we assume it's not)
		if os.Getuid() != 0 {
			_, err = EnsurePythonInstalled(ctx)
			if err == nil || !strings.Contains(err.Error(), "requires root privileges") {
				t.Fatalf("Expected failure due to missing sudo, got %v", err)
			}
		}

		// Restore sudo for remaining tests
		createFakeExecutable(t, fakeSudo, 0, "")

		// 4. Test installation command failure
		os.Remove(filepath.Join(tempDirAuto, "python3"))
		// Modify fake sh to fail
		os.WriteFile(fakeSh, []byte("#!/bin/sh\nexit 1\n"), 0755)
		
		_, err = EnsurePythonInstalled(ctx)
		if err == nil || !strings.Contains(err.Error(), "failed to install Python 3 automatically") {
			t.Fatalf("Expected failure due to command failure, got %v", err)
		}

		// 5. Test installation success but python still missing
		// Modify fake sh to succeed, but NOT create python3
		os.WriteFile(fakeSh, []byte("#!/bin/sh\nexit 0\n"), 0755)
		
		_, err = EnsurePythonInstalled(ctx)
		if err == nil || !strings.Contains(err.Error(), "still not in PATH") {
			t.Fatalf("Expected failure due to missing python after install, got %v", err)
		}

		// 6. Test no supported package manager found
		os.Remove(fakeApt)
		_, err = EnsurePythonInstalled(ctx)
		if err == nil || !strings.Contains(err.Error(), "could not detect a supported package manager") {
			t.Fatalf("Expected failure due to missing package manager, got %v", err)
		}
	} else {
		// 7. Test Windows missing python early exit
		t.Setenv("PATH", t.TempDir()) // Empty PATH
		_, err = EnsurePythonInstalled(ctx)
		if err == nil || !strings.Contains(err.Error(), "must be installed manually on Windows") {
			t.Fatalf("Expected early failure on Windows, got %v", err)
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

func TestSetupVirtualEnvironmentFailure(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	var pyName string
	if runtime.GOOS == "windows" {
		pyName = "python.bat"
	} else {
		pyName = "python"
	}
	fakePy := filepath.Join(tempDir, pyName)
	createFakeExecutable(t, fakePy, 1, "venv creation error")

	prependPath(t, tempDir)

	_, err := SetupVirtualEnvironment(ctx, pyName, filepath.Join(tempDir, "venv2"))
	if err == nil || !strings.Contains(err.Error(), "failed to create venv") {
		t.Fatalf("Expected SetupVirtualEnvironment to fail, got %v", err)
	}
}
