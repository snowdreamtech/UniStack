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
	"gopkg.in/yaml.v3"
)

// Installer handles installing packages from local tarballs or remote registry.
type Installer struct {
	PackagesDir string
	BinDir      string
}

// PackageMetadata holds the inner metadata for a package.yml
type PackageMetadata struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// PackageManifest represents the structure of package.yml
type PackageManifest struct {
	ApiVersion string          `yaml:"apiVersion"`
	Kind       string          `yaml:"kind"`
	Metadata   PackageMetadata `yaml:"metadata"`
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

	// 1. Extract to a temporary directory first to read package.yml
	base := filepath.Base(pkgPath)
	tmpExtractDir := filepath.Join(i.PackagesDir, ".tmp-"+base)
	
	if err := os.MkdirAll(i.PackagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create packages directory: %w", err)
	}

	_ = os.RemoveAll(tmpExtractDir)

	if err := ExtractTarGz(pkgPath, tmpExtractDir); err != nil {
		_ = os.RemoveAll(tmpExtractDir)
		return fmt.Errorf("failed to extract package: %w", err)
	}

	// 2. Read package.yml
	manifestPath := filepath.Join(tmpExtractDir, "package.yml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		_ = os.RemoveAll(tmpExtractDir)
		return fmt.Errorf("failed to read package.yml: %w", err)
	}

	var manifest PackageManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		_ = os.RemoveAll(tmpExtractDir)
		return fmt.Errorf("invalid package.yml format: %w", err)
	}

	if manifest.Metadata.Name == "" {
		_ = os.RemoveAll(tmpExtractDir)
		return fmt.Errorf("package.yml is missing metadata.name")
	}

	pkgID := fmt.Sprintf("%s-%s", manifest.Metadata.Name, manifest.Metadata.Version)
	finalDir := filepath.Join(i.PackagesDir, pkgID)

	// Check if already installed
	if _, err := os.Stat(finalDir); err == nil {
		_ = os.RemoveAll(tmpExtractDir)
		return fmt.Errorf("package %s is already installed", pkgID)
	}

	if err := os.Rename(tmpExtractDir, finalDir); err != nil {
		_ = os.RemoveAll(tmpExtractDir)
		return fmt.Errorf("failed to move package to final directory: %w", err)
	}

	// 3. Execute Ansible if tasks/main.yml exists
	tasksMainPath := filepath.Join(finalDir, "tasks", "main.yml")
	if _, err := os.Stat(tasksMainPath); err == nil {
		if err := runAnsibleRole(finalDir, ""); err != nil {
			return fmt.Errorf("ansible-playbook failed: %w", err)
		}
	}

	// 4. Symlink
	execName := manifest.Metadata.Name
	execPath := filepath.Join(finalDir, "bin", execName)
	if _, err := os.Stat(execPath); os.IsNotExist(err) {
		execPath = filepath.Join(finalDir, execName)
	}
	
	// We do not fail if executable is not found, because meta packages or Ansible-only packages might not have a direct binary in bin/
	if _, err := os.Stat(execPath); err == nil {
		linkPath := filepath.Join(i.BinDir, execName)
		if err := CreateSymlink(execPath, linkPath); err != nil {
			return fmt.Errorf("failed to create symlink for %s: %w", execName, err)
		}
	}

	return nil
}

// runAnsibleRole generates a temporary playbook and executes the ansible role in the package path.
// state can be empty (install) or "absent" (uninstall).
func runAnsibleRole(pkgPath string, state string) error {
	playbookContent := `
- hosts: localhost
  connection: local
  roles:
    - role: .
`
	playbookPath := filepath.Join(pkgPath, "_unistack_playbook.yml")
	if err := os.WriteFile(playbookPath, []byte(playbookContent), 0644); err != nil {
		return fmt.Errorf("failed to create temporary playbook: %w", err)
	}
	defer os.Remove(playbookPath)

	fmt.Printf("Detected Ansible role in %s. Executing (state=%s)...\n", pkgPath, state)
	
	args := []string{"-i", "localhost,", "-c", "local", playbookPath, "-e", fmt.Sprintf("app_source_path=%s", pkgPath)}
	if state != "" {
		args = append(args, "-e", fmt.Sprintf("state=%s", state))
	}
	
	cmd := exec.Command("ansible-playbook", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ListInstalledPackages returns a list of installed PackageManifests parsed from their package.yml files.
func (i *Installer) ListInstalledPackages() ([]PackageManifest, error) {
	entries, err := os.ReadDir(i.PackagesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []PackageManifest{}, nil
		}
		return nil, err
	}
	
	var manifests []PackageManifest
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			manifestPath := filepath.Join(i.PackagesDir, e.Name(), "package.yml")
			data, err := os.ReadFile(manifestPath)
			if err != nil {
				continue // skip invalid directories without package.yml
			}
			
			var manifest PackageManifest
			if err := yaml.Unmarshal(data, &manifest); err == nil && manifest.Metadata.Name != "" {
				manifests = append(manifests, manifest)
			}
		}
	}
	return manifests, nil
}

// Uninstall removes a package from the system, invoking Ansible uninstall logic if present.
func (i *Installer) Uninstall(pkgName string) error {
	pkgs, err := i.ListInstalledPackages()
	if err != nil {
		return err
	}
	
	var installedPkg *PackageManifest
	var finalDir string
	
	for _, p := range pkgs {
		if p.Metadata.Name == pkgName {
			installedPkg = &p
			pkgID := fmt.Sprintf("%s-%s", p.Metadata.Name, p.Metadata.Version)
			finalDir = filepath.Join(i.PackagesDir, pkgID)
			break
		}
	}
	
	if installedPkg == nil {
		return fmt.Errorf("package %s is not installed", pkgName)
	}

	// 1. Ansible absent if playbook exists
	tasksMainPath := filepath.Join(finalDir, "tasks", "main.yml")
	if _, err := os.Stat(tasksMainPath); err == nil {
		if err := runAnsibleRole(finalDir, "absent"); err != nil {
			return fmt.Errorf("ansible-playbook uninstall failed: %w", err)
		}
	}

	// 2. Remove symlink
	linkPath := filepath.Join(i.BinDir, pkgName)
	if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove symlink for %s: %w", pkgName, err)
	}

	// 3. Remove package directory
	if err := os.RemoveAll(finalDir); err != nil {
		return fmt.Errorf("failed to remove package directory %s: %w", finalDir, err)
	}

	return nil
}
