// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package client

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstaller_InstallFromLocal(t *testing.T) {
	tmpDir := t.TempDir()

	packagesDir := filepath.Join(tmpDir, "packages")
	binDir := filepath.Join(tmpDir, "bin")

	installer := &Installer{
		PackagesDir: packagesDir,
		BinDir:      binDir,
	}

	tarballPath := filepath.Join(tmpDir, "hello-1.0.0.tar.gz")

	files := map[string]string{
		"bin/hello":   "#!/bin/sh\necho 'hello'",
		"package.yml": "apiVersion: v1alpha1\nkind: package\nmetadata:\n  name: hello\n  version: 1.0.0\n",
	}

	createTestTarball(t, tarballPath, files)

	if err := installer.InstallFromLocal(tarballPath); err != nil {
		t.Fatalf("InstallFromLocal failed: %v", err)
	}

	// Verify extracted files
	extractedFile := filepath.Join(packagesDir, "hello-1.0.0", "bin", "hello")
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Errorf("extracted file %s does not exist", extractedFile)
	}

	// Verify symlink
	linkPath := filepath.Join(binDir, "hello")
	linkTarget, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("failed to read symlink %s: %v", linkPath, err)
	}

	if linkTarget != extractedFile {
		t.Errorf("symlink target mismatch: expected %q, got %q", extractedFile, linkTarget)
	}
}

func TestInstaller_InstallFromLocal_InvalidPackage(t *testing.T) {
	installer := &Installer{}
	err := installer.InstallFromLocal("invalid.zip")
	if err == nil {
		t.Error("expected error for non .tar.gz file")
	}
}
