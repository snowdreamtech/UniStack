// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package client

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unistack/internal/env"
)

// Installer handles installing packages from local tarballs or remote registry.
type Installer struct {
	PackagesDir string
	BinDir      string
}

// NewInstaller creates a new Installer with default paths.
func NewInstaller() *Installer {
	return &Installer{
		PackagesDir: env.GetInstallPackagesDir(),
		BinDir:      env.GetInstallBinDir(),
	}
}

// InstallFromLocal installs a package from a local .tar.gz file.
// It extracts it atomically and creates a symlink to its executable.
func (i *Installer) InstallFromLocal(pkgPath string) error {
	if !strings.HasSuffix(pkgPath, ".tar.gz") {
		return fmt.Errorf("only .tar.gz packages are supported currently")
	}

	// Extract name and version from file name (e.g. hello-1.0.0.tar.gz -> hello-1.0.0)
	base := filepath.Base(pkgPath)
	pkgID := strings.TrimSuffix(base, ".tar.gz")

	// Try to guess the executable name (e.g., hello from hello-1.0.0)
	parts := strings.Split(pkgID, "-")
	execName := parts[0]
	if len(parts) > 1 {
		execName = strings.Join(parts[:len(parts)-1], "-") // Handle names with hyphens
	}

	finalDir := filepath.Join(i.PackagesDir, pkgID)
	tmpDir := filepath.Join(i.PackagesDir, ".tmp-"+pkgID)

	// Check if already installed
	if _, err := os.Stat(finalDir); err == nil {
		return fmt.Errorf("package %s is already installed", pkgID)
	}

	// Cleanup tmp dir if it exists from a previous failed run
	_ = os.RemoveAll(tmpDir)

	if err := os.MkdirAll(i.PackagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create packages directory: %w", err)
	}

	// 1. Extract atomically
	if err := ExtractTarGz(pkgPath, tmpDir); err != nil {
		_ = os.RemoveAll(tmpDir)
		return fmt.Errorf("failed to extract package: %w", err)
	}

	if err := os.Rename(tmpDir, finalDir); err != nil {
		_ = os.RemoveAll(tmpDir)
		return fmt.Errorf("failed to move package to final directory: %w", err)
	}

	// 2. Execute Ansible if app_loader.yml exists
	appLoaderPath := filepath.Join(finalDir, "app_loader.yml")
	if _, err := os.Stat(appLoaderPath); err == nil {
		if err := runAnsiblePlaybook(finalDir); err != nil {
			return fmt.Errorf("ansible-playbook failed: %w", err)
		}
	}

	// 3. Symlink
	// Try to find the executable: either in bin/ or root
	execPath := filepath.Join(finalDir, "bin", execName)
	if _, err := os.Stat(execPath); os.IsNotExist(err) {
		execPath = filepath.Join(finalDir, execName)
		if _, err := os.Stat(execPath); os.IsNotExist(err) {
			return fmt.Errorf("could not find executable %q in extracted package", execName)
		}
	}

	linkPath := filepath.Join(i.BinDir, execName)
	if err := CreateSymlink(execPath, linkPath); err != nil {
		return fmt.Errorf("failed to create symlink for %s: %w", execName, err)
	}

	return nil
}

// runAnsiblePlaybook executes the app_loader.yml playbook located in the package path.
func runAnsiblePlaybook(pkgPath string) error {
	appLoaderPath := filepath.Join(pkgPath, "app_loader.yml")
	fmt.Printf("Detected Ansible playbook %s. Executing...\n", appLoaderPath)
	cmd := exec.Command("ansible-playbook", "-i", "localhost,", "-c", "local", appLoaderPath, "-e", fmt.Sprintf("app_source_path=%s", pkgPath))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ListInstalledPackages returns a list of installed packages in the form of "pkgID" strings (e.g. "hello-1.0.0")
func (i *Installer) ListInstalledPackages() ([]string, error) {
	entries, err := os.ReadDir(i.PackagesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var pkgs []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			pkgs = append(pkgs, e.Name())
		}
	}
	return pkgs, nil
}

// Uninstall removes a package from the system, invoking Ansible uninstall logic if present.
func (i *Installer) Uninstall(pkgID string) error {
	finalDir := filepath.Join(i.PackagesDir, pkgID)
	
	if _, err := os.Stat(finalDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("package %s is not installed", pkgID)
		}
		return err
	}

	// Guess exec name similarly to InstallFromLocal
	parts := strings.Split(pkgID, "-")
	execName := parts[0]
	if len(parts) > 1 {
		execName = strings.Join(parts[:len(parts)-1], "-")
	}

	// 1. Ansible absent if playbook exists
	appLoaderPath := filepath.Join(finalDir, "app_loader.yml")
	if _, err := os.Stat(appLoaderPath); err == nil {
		fmt.Printf("Detected Ansible playbook %s. Executing uninstall...\n", appLoaderPath)
		cmd := exec.Command("ansible-playbook", "-i", "localhost,", "-c", "local", appLoaderPath, "-e", fmt.Sprintf("app_source_path=%s", finalDir), "-e", "state=absent")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("ansible-playbook uninstall failed: %w", err)
		}
	}

	// 2. Remove symlink
	linkPath := filepath.Join(i.BinDir, execName)
	if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove symlink for %s: %w", execName, err)
	}

	// 3. Remove package directory
	if err := os.RemoveAll(finalDir); err != nil {
		return fmt.Errorf("failed to remove package directory %s: %w", finalDir, err)
	}

	return nil
}
