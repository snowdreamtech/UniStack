// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRunPostflightChecks(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	pythonName := "python"
	ansibleName := "ansible-playbook"

	if runtime.GOOS == "windows" {
		pythonName += ".bat"
		ansibleName += ".bat"
	} else {
		pythonName += ".sh"
		ansibleName += ".sh"
	}

	fakePython := filepath.Join(tempDir, pythonName)
	fakeAnsible := filepath.Join(tempDir, ansibleName)

	// Create fake executables that output valid version strings
	createFakeExecutable(t, fakePython, 0, "Python 3.12.0")
	createFakeExecutable(t, fakeAnsible, 0, "ansible-playbook 2.15.0")

	err := RunPostflightChecks(ctx, nil, tempDir, fakePython, fakeAnsible)
	if err != nil {
		t.Fatalf("Expected RunPostflightChecks to succeed, got: %v", err)
	}

	// Test with a failing python executable
	fakeFailPython := filepath.Join(tempDir, "fail_"+pythonName)
	createFakeExecutable(t, fakeFailPython, 1, "Error")

	err = RunPostflightChecks(ctx, nil, tempDir, fakeFailPython, fakeAnsible)
	if err == nil {
		t.Fatalf("Expected RunPostflightChecks to fail when python check fails")
	}

	// Test with a failing ansible executable
	fakeFailAnsible := filepath.Join(tempDir, "fail_"+ansibleName)
	createFakeExecutable(t, fakeFailAnsible, 1, "Error")

	err = RunPostflightChecks(ctx, nil, tempDir, fakePython, fakeFailAnsible)
	if err == nil {
		t.Fatalf("Expected RunPostflightChecks to fail when ansible check fails")
	}
}
