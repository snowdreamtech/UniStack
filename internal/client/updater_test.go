// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package client

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/snowdreamtech/unistack/internal/env"
	"github.com/snowdreamtech/unistack/internal/registry"
)

func TestUpdateRegistry(t *testing.T) {
	// Set up a mock sqlite DB content (just dummy data)
	dummyDBContent := []byte("SQLite format 3\x00...")

	// Compress it using zstd
	var buf bytes.Buffer
	enc, err := zstd.NewWriter(&buf)
	if err != nil {
		t.Fatalf("Failed to create zstd writer: %v", err)
	}
	enc.Write(dummyDBContent)
	enc.Close()

	compressedData := buf.Bytes()

	// Calculate hash
	h := sha256.New()
	h.Write(compressedData)
	hashStr := "sha256:" + hex.EncodeToString(h.Sum(nil))

	repomd := registry.RepoMd{
		Timestamp: time.Now().Unix(),
		Hash:      hashStr,
		Path:      "repodata/packages.db.zst",
	}

	// Create a test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repodata/repomd.json" {
			json.NewEncoder(w).Encode(repomd)
			return
		}
		if r.URL.Path == "/repodata/packages.db.zst" {
			w.Write(compressedData)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	// Override env directories for testing so we don't mess up real data
	tempDir := t.TempDir()

	// Temporarily override env config
	originalCacheHome := os.Getenv("XDG_CACHE_HOME")
	os.Setenv("XDG_CACHE_HOME", tempDir)
	defer os.Setenv("XDG_CACHE_HOME", originalCacheHome)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = UpdateSource(ctx, "test_source", ts.URL)
	if err != nil {
		t.Fatalf("UpdateSource failed: %v", err)
	}

	// Verify that the file was created and contains the uncompressed data
	dbPath := env.GetSourceDatabasePath("test_source")
	content, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("Failed to read decompressed DB file: %v", err)
	}

	if !bytes.Equal(content, dummyDBContent) {
		t.Errorf("Decompressed content mismatch. Got %d bytes, want %d bytes", len(content), len(dummyDBContent))
	}

	// Verify that local repomd.json was saved
	localRepomdPath := filepath.Join(env.GetRegistryCacheDir(), "test_source_repomd.json")
	if _, err := os.Stat(localRepomdPath); os.IsNotExist(err) {
		t.Errorf("Local repomd.json was not saved")
	}

	// Run it again, it should skip downloading because hash matches
	err = UpdateSource(ctx, "test_source", ts.URL)
	if err != nil {
		t.Fatalf("Second UpdateSource (skip) failed: %v", err)
	}
}
