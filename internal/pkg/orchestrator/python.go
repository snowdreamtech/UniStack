// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package orchestrator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// EnsurePythonInstalled checks for python3 (or python on Windows).
// If missing on POSIX, it attempts to detect the package manager and install it.
func EnsurePythonInstalled(ctx context.Context) (string, error) {
	// 1. Fast Path: Check if python3 (or python) is already available
	pythonCmd := "python3"
	if _, err := exec.LookPath(pythonCmd); err != nil {
		if _, err := exec.LookPath("python"); err == nil {
			pythonCmd = "python"
		} else {
			pythonCmd = ""
		}
	}

	if pythonCmd != "" {
		fmt.Printf("✅ Found system Python at %s\n", pythonCmd)
		return pythonCmd, nil
	}

	// 2. Python not found. Handle Windows separately.
	if runtime.GOOS == "windows" {
		return "", fmt.Errorf("🚨 Python 3 is missing. The only strict requirement is that Python 3 must be installed manually on Windows. Please install it and ensure it is in your PATH")
	}

	// 3. Attempt POSIX auto-installation
	fmt.Println("⚠️ Python 3 not found. Attempting automatic installation via system package manager...")

	// Detect package manager
	packageManagers := []string{
		"apk", "apt-get", "microdnf", "dnf", "yum", "pacman", 
		"zypper", "xbps-install", "emerge", "pkg", "pkg_add", "brew", "port",
		"swupd", "eopkg", "nix-env", "opkg", "tdnf", "urpmi", "slackpkg",
	}

	var detectedPM string
	for _, pm := range packageManagers {
		if _, err := exec.LookPath(pm); err == nil {
			detectedPM = pm
			break
		}
	}

	if detectedPM == "" {
		return "", fmt.Errorf("could not detect a supported package manager for auto-installation. Please install Python 3 manually")
	}

	fmt.Printf("📦 Detected package manager: %s\n", detectedPM)

	// Map package manager to installation shell command
	var installCmd string
	switch detectedPM {
	case "apk":
		installCmd = "apk update || true && apk add --no-cache python3 py3-pip"
	case "apt-get":
		installCmd = "export DEBIAN_FRONTEND=noninteractive && apt-get update -y && apt-get install -y python3 python3-pip python3-venv python3-distutils python3-setuptools python3-wheel"
	case "microdnf":
		installCmd = "microdnf install -y python3 python3-pip || microdnf install -y python3"
	case "dnf":
		installCmd = "dnf install -y python3 python3-pip || dnf install -y python3 python3-pip-wheel || dnf install -y python3 python3-wheel python3-pip"
	case "yum":
		installCmd = "yum install -y python3 python3-pip || yum install -y python3 python3-pip-wheel || yum install -y python3 python3-wheel python3-pip"
	case "pacman":
		installCmd = "pacman -Sy --noconfirm python python-pip"
	case "zypper":
		installCmd = "zypper --non-interactive install python3 python3-pip"
	case "xbps-install":
		installCmd = "xbps-install -Sy python3 python3-pip"
	case "emerge":
		installCmd = "emerge -q dev-lang/python dev-python/pip || emerge -q dev-lang/python"
	case "pkg":
		installCmd = "pkg install -y python3 py39-pip || pkg install -y python3 py38-pip"
	case "pkg_add":
		installCmd = "pkg_add -I python%3 || pkg_add -I python3"
	case "brew":
		installCmd = "brew install python3"
	case "port":
		installCmd = "port selfupdate || true && port install python311 || port install python310 || port install python3"
	case "swupd":
		installCmd = "swupd bundle-add python3-basic"
	case "eopkg":
		installCmd = "eopkg install -y python3"
	case "nix-env":
		installCmd = "nix-env -iA nixpkgs.python3"
	case "opkg":
		installCmd = "opkg update && opkg install python3"
	case "tdnf":
		installCmd = "tdnf install -y python3"
	case "urpmi":
		installCmd = "urpmi --auto python3"
	case "slackpkg":
		installCmd = "slackpkg update && slackpkg install python3"
	}

	// Sudo integration
	if os.Getuid() != 0 && detectedPM != "brew" {
		// Verify sudo is available
		if _, err := exec.LookPath("sudo"); err == nil {
			fmt.Println("🔑 Root privileges required for installation. Attempting to use sudo (you may be prompted for your password)...")
			installCmd = "sudo sh -c '" + strings.ReplaceAll(installCmd, "'", "'\\''") + "'"
		} else {
			return "", fmt.Errorf("installation requires root privileges, but 'sudo' is not installed. Please run as root or install Python 3 manually")
		}
	} else {
		installCmd = "sh -c '" + strings.ReplaceAll(installCmd, "'", "'\\''") + "'"
	}

	// Execute installation
	cmd := exec.CommandContext(ctx, "sh", "-c", installCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to install Python 3 automatically: %w", err)
	}

	fmt.Println("✅ Automatic Python 3 installation complete")

	// 4. Re-verify Python 3 installation
	if _, err := exec.LookPath("python3"); err == nil {
		return "python3", nil
	} else if _, err := exec.LookPath("python"); err == nil {
		return "python", nil
	}

	return "", fmt.Errorf("python installation appeared to succeed, but 'python3' (or 'python') is still not in PATH")
}

// buildVenvEnv creates environment variables needed to run binaries inside a virtualenv
func buildVenvEnv(venvDir string) []string {
	envVars := os.Environ()
	
	venvBinDir := filepath.Join(venvDir, "bin")
	if runtime.GOOS == "windows" {
		venvBinDir = filepath.Join(venvDir, "Scripts")
	}

	pathUpdated := false
	for i, e := range envVars {
		if strings.HasPrefix(strings.ToUpper(e), "PATH=") {
			envVars[i] = fmt.Sprintf("PATH=%s%c%s", venvBinDir, os.PathListSeparator, e[5:])
			pathUpdated = true
			break
		}
	}

	if !pathUpdated {
		envVars = append(envVars, fmt.Sprintf("PATH=%s", venvBinDir))
	}
	envVars = append(envVars, fmt.Sprintf("VIRTUAL_ENV=%s", venvDir))
	return envVars
}

// SetupVirtualEnvironment creates the Python venv and sets up the environment variables
func SetupVirtualEnvironment(ctx context.Context, pythonCmd, venvDir string) ([]string, error) {
	fmt.Println("🚀 Bootstrapping Python Virtual Environment for Ansible...")

	// Create venv
	cmd := exec.CommandContext(ctx, pythonCmd, "-m", "venv", venvDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create venv: %w", err)
	}

	return buildVenvEnv(venvDir), nil
}
