// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
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
