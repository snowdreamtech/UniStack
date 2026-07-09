// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// createFakeExecutable generates a dummy executable file at the specified path
// that simply exits with exitCode and optionally prints standard output.
func createFakeExecutable(t *testing.T, path string, exitCode int, stdout string) {
	t.Helper()

	// Ensure the parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create dir for fake executable %s: %v", path, err)
	}

	var content []byte
	if runtime.GOOS == "windows" {
		// Create a batch file on Windows
		content = []byte(fmt.Sprintf("@echo off\r\n"))
		if stdout != "" {
			content = append(content, []byte(fmt.Sprintf("echo %s\r\n", stdout))...)
		}
		content = append(content, []byte(fmt.Sprintf("exit /b %d\r\n", exitCode))...)
	} else {
		// Create a shell script on Unix
		content = []byte(fmt.Sprintf("#!/bin/sh\n"))
		if stdout != "" {
			content = append(content, []byte(fmt.Sprintf("echo \"%s\"\n", stdout))...)
		}
		content = append(content, []byte(fmt.Sprintf("exit %d\n", exitCode))...)
	}

	if err := os.WriteFile(path, content, 0755); err != nil {
		t.Fatalf("Failed to write fake executable %s: %v", path, err)
	}
}

// prependPath adds a directory to the front of the PATH environment variable
// for the duration of the test.
func prependPath(t *testing.T, dir string) {
	t.Helper()

	currentPath := os.Getenv("PATH")
	var newPath string
	if runtime.GOOS == "windows" {
		newPath = fmt.Sprintf("%s;%s", dir, currentPath)
	} else {
		newPath = fmt.Sprintf("%s:%s", dir, currentPath)
	}

	t.Setenv("PATH", newPath)
}

// createSmartFakePython generates a fake python executable that simulates "python -m venv <dir>"
// by actually creating the necessary dummy bin/pip and bin/ansible-galaxy files.
func createSmartFakePython(t *testing.T, path string) {
	t.Helper()
	var pyScript string
	if runtime.GOOS == "windows" {
		pyScript = `@echo off
if "%1"=="-m" if "%2"=="venv" (
	mkdir "%3\Scripts" 2>nul
	echo @echo off > "%3\Scripts\pip.bat"
	echo exit /b 0 >> "%3\Scripts\pip.bat"
	echo @echo off > "%3\Scripts\ansible-galaxy.bat"
	echo exit /b 0 >> "%3\Scripts\ansible-galaxy.bat"
	echo @echo off > "%3\Scripts\ansible-playbook.bat"
	echo exit /b 0 >> "%3\Scripts\ansible-playbook.bat"
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
	os.WriteFile(path, []byte(pyScript), 0755)
}
