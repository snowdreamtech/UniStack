// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package client

import (
	"fmt"
	"os"
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

	// 2. Symlink
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
