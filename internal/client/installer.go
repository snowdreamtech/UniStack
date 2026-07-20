// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package client

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unistack/internal/config"
	"github.com/snowdreamtech/unistack/internal/env"
	"github.com/snowdreamtech/unistack/internal/registry"
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

	safeName := strings.ReplaceAll(manifest.Metadata.Name, "/", "_")
	pkgID := fmt.Sprintf("%s-%s", safeName, manifest.Metadata.Version)
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
	// Use filepath.Base so that namespaced packages like "snowdreamtech/hello" create a symlink named "hello"
	execName := filepath.Base(manifest.Metadata.Name)
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
			safeName := strings.ReplaceAll(p.Metadata.Name, "/", "_")
			pkgID := fmt.Sprintf("%s-%s", safeName, p.Metadata.Version)
			finalDir = filepath.Join(i.PackagesDir, pkgID)
			break
		}
	}

	if installedPkg == nil {
		return fmt.Errorf("package %s is not installed", pkgName)
	}

	// 0. Check for reverse dependencies
	db, err := registry.OpenRegistryDB()
	if err == nil {
		// It's possible the registry DB isn't initialized locally if only local packages were installed,
		// but if it exists, check for reverse dependencies.
		revDeps, err := registry.GetReverseDependencies(context.Background(), db, pkgName)
		db.Close()
		if err == nil && len(revDeps) > 0 {
			// Filter to see if any are currently installed
			installedMap := make(map[string]bool)
			for _, p := range pkgs {
				installedMap[p.Metadata.Name] = true
			}

			var installedRevDeps []string
			for _, rd := range revDeps {
				if installedMap[rd] {
					installedRevDeps = append(installedRevDeps, rd)
				}
			}

			if len(installedRevDeps) > 0 {
				return fmt.Errorf("cannot uninstall %s because the following installed packages depend on it: %v", pkgName, installedRevDeps)
			}
		}
	}

	// 1. Ansible absent if playbook exists
	tasksMainPath := filepath.Join(finalDir, "tasks", "main.yml")
	if _, err := os.Stat(tasksMainPath); err == nil {
		if err := runAnsibleRole(finalDir, "absent"); err != nil {
			return fmt.Errorf("ansible-playbook uninstall failed: %w", err)
		}
	}

	// 2. Remove symlink
	execName := filepath.Base(pkgName)
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

// InstallPackage resolves dependencies, downloads, and installs a package and its dependencies.
func (i *Installer) InstallPackage(ctx context.Context, targetPkg string) error {
	// 1. Open registry database
	db, err := registry.OpenRegistryDB()
	if err != nil {
		return fmt.Errorf("failed to open registry database: %w", err)
	}
	defer db.Close()

	// 2. Build dependency graph
	graph := NewDependencyGraph()
	fmt.Printf("Resolving dependencies for %s...\n", targetPkg)
	if err := graph.BuildGraph(ctx, db, targetPkg); err != nil {
		return fmt.Errorf("failed to build dependency graph: %w", err)
	}

	// 3. Topological Sort
	sorted, err := graph.TopologicalSort()
	if err != nil {
		if errors.Is(err, ErrCircularDependency) {
			return err
		}
		return fmt.Errorf("failed to sort dependencies: %w", err)
	}

	// 4. Check already installed packages (T009)
	installedPkgs, err := i.ListInstalledPackages()
	if err != nil {
		return fmt.Errorf("failed to list installed packages: %w", err)
	}
	installedMap := make(map[string]bool)
	for _, p := range installedPkgs {
		installedMap[p.Metadata.Name] = true
	}

	downloader := NewDownloader()

	// 5. Download and install in order
	for _, pkgName := range sorted {
		if installedMap[pkgName] {
			fmt.Printf("Package %s is already installed, skipping.\n", pkgName)
			continue
		}

		meta, err := registry.QueryPackage(ctx, pkgName)
		if err != nil {
			return fmt.Errorf("failed to query package %s from registry: %w", pkgName, err)
		}
		if meta == nil {
			return fmt.Errorf("package %q not found in registry", pkgName)
		}

		// Resolve source URL
		sources, err := config.LoadSources()
		if err != nil {
			return fmt.Errorf("failed to load sources config: %w", err)
		}
		var registryURL string
		for _, s := range sources {
			if s.Name == meta.Source {
				registryURL = s.URL
				break
			}
		}
		if registryURL == "" {
			return fmt.Errorf("registry source %q for package %s not found in configuration", meta.Source, pkgName)
		}

		fmt.Printf("Downloading %s version %s from %s...\n", meta.Name, meta.Version, meta.Source)
		downloadedPath, err := downloader.DownloadPackage(ctx, registryURL, meta)
		if err != nil {
			return fmt.Errorf("failed to download package %s: %w", pkgName, err)
		}

		fmt.Printf("Installing %s...\n", meta.Name)
		if err := i.InstallFromLocal(downloadedPath); err != nil {
			return fmt.Errorf("installation failed for %s: %w", pkgName, err)
		}
	}

	return nil
}
