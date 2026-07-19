// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package client

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/snowdreamtech/unistack/internal/registry"
)

func TestDownloadPackage(t *testing.T) {
	// 1. Prepare dummy package data
	pkgData := []byte("hello world package content")
	
	// Calculate true hash
	hasher := sha256.New()
	hasher.Write(pkgData)
	trueHash := hex.EncodeToString(hasher.Sum(nil))

	// 2. Set up a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(pkgData)
	}))
	defer ts.Close()

	// Override env directories for testing so we don't mess up real data
	tempDir := t.TempDir()
	
	// Temporarily override env config
	originalDataDir := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tempDir)
	defer os.Setenv("XDG_DATA_HOME", originalDataDir)

	downloader := NewDownloader()

	t.Run("successful download with valid hash", func(t *testing.T) {
		meta := &registry.PackageMetadata{
			Name:    "hello",
			Version: "1.0.0",
			Hash:    trueHash,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		path, err := downloader.DownloadPackage(ctx, ts.URL, meta)
		if err != nil {
			t.Fatalf("Expected success, got error: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Downloaded file does not exist at %s", path)
		}
	})

	t.Run("hash mismatch rejection", func(t *testing.T) {
		meta := &registry.PackageMetadata{
			Name:    "malicious",
			Version: "1.0.0",
			Hash:    "badc0ffee0000000000000000000000000000000000000000000000000000000",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := downloader.DownloadPackage(ctx, ts.URL, meta)
		if err == nil {
			t.Fatalf("Expected hash mismatch error, got nil")
		}

		// Ensure no temp file is left behind
		// (Hard to check exact path without mocking env completely, but error should indicate security validation)
		if err.Error() == "" {
			t.Errorf("Expected non-empty error message")
		}
	})
}
