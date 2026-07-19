// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unistack/internal/registry"
)

func TestInstaller_InstallFromRemote(t *testing.T) {
	// 1. Setup local registry DB and server
	tmpDir := t.TempDir()
	tarballPath := filepath.Join(tmpDir, "hello-1.0.0.tar.gz")
	
	files := map[string]string{
		"bin/hello": "#!/bin/sh\necho 'hello from remote'",
	}
	createTestTarball(t, tarballPath, files)
	tarBytes, err := os.ReadFile(tarballPath)
	if err != nil {
		t.Fatalf("failed to read created tarball: %v", err)
	}
	
	// Create a test server to serve the tarball
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/hello-1.0.0.tar.gz" {
			w.Write(tarBytes)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	// 2. Setup Installer
	packagesDir := filepath.Join(tmpDir, "packages")
	binDir := filepath.Join(tmpDir, "bin")
	
	installer := &Installer{
		PackagesDir: packagesDir,
		BinDir:      binDir,
	}

	downloader := NewDownloader()

	meta := &registry.PackageMetadata{
		Name:    "hello",
		Version: "1.0.0",
		// Omit Hash for simplicity in this integration test
	}

	// 3. Download and Install
	ctx := context.Background()
	downloadedPath, err := downloader.DownloadPackage(ctx, ts.URL, meta)
	if err != nil {
		t.Fatalf("DownloadPackage failed: %v", err)
	}
	
	if err := installer.InstallFromLocal(downloadedPath); err != nil {
		t.Fatalf("InstallFromLocal failed: %v", err)
	}

	// 4. Verify
	extractedFile := filepath.Join(packagesDir, "hello-1.0.0", "bin", "hello")
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Errorf("extracted file %s does not exist", extractedFile)
	}
	
	linkPath := filepath.Join(binDir, "hello")
	linkTarget, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("failed to read symlink %s: %v", linkPath, err)
	}
	
	if linkTarget != extractedFile {
		t.Errorf("symlink target mismatch: expected %q, got %q", extractedFile, linkTarget)
	}
}
