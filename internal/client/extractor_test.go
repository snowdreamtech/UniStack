// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package client

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func createTestTarball(t *testing.T, path string, files map[string]string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create tarball file: %v", err)
	}
	defer file.Close()

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("failed to write header for %s: %v", name, err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatalf("failed to write content for %s: %v", name, err)
		}
	}
}

func TestExtractTarGz(t *testing.T) {
	tmpDir := t.TempDir()
	tarballPath := filepath.Join(tmpDir, "test.tar.gz")
	destDir := filepath.Join(tmpDir, "extracted")

	files := map[string]string{
		"hello.txt":       "Hello World",
		"bin/script.sh":   "#!/bin/sh\necho 'hello'",
		"nested/file.txt": "nested content",
	}

	createTestTarball(t, tarballPath, files)

	err := ExtractTarGz(tarballPath, destDir)
	if err != nil {
		t.Fatalf("ExtractTarGz failed: %v", err)
	}

	for name, expectedContent := range files {
		extractedPath := filepath.Join(destDir, name)
		content, err := os.ReadFile(extractedPath)
		if err != nil {
			t.Errorf("failed to read extracted file %s: %v", name, err)
		} else if string(content) != expectedContent {
			t.Errorf("content mismatch for %s: expected %q, got %q", name, expectedContent, string(content))
		}
	}
}

func TestCreateSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "target.txt")
	linkPath := filepath.Join(tmpDir, "bin", "target-link")

	if err := os.WriteFile(targetPath, []byte("target"), 0644); err != nil {
		t.Fatalf("failed to write target file: %v", err)
	}

	if err := CreateSymlink(targetPath, linkPath); err != nil {
		t.Fatalf("CreateSymlink failed: %v", err)
	}

	// Verify symlink
	linkTarget, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("failed to read symlink: %v", err)
	}
	if linkTarget != targetPath {
		t.Errorf("symlink target mismatch: expected %q, got %q", targetPath, linkTarget)
	}

	// Test replacing existing symlink
	newTargetPath := filepath.Join(tmpDir, "target2.txt")
	if err := os.WriteFile(newTargetPath, []byte("target2"), 0644); err != nil {
		t.Fatalf("failed to write new target file: %v", err)
	}

	if err := CreateSymlink(newTargetPath, linkPath); err != nil {
		t.Fatalf("CreateSymlink (overwrite) failed: %v", err)
	}

	linkTarget2, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("failed to read overwritten symlink: %v", err)
	}
	if linkTarget2 != newTargetPath {
		t.Errorf("overwritten symlink target mismatch: expected %q, got %q", newTargetPath, linkTarget2)
	}
}
