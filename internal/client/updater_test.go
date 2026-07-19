// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package client

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/snowdreamtech/unistack/internal/env"
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

	// Create a test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(compressedData)
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

	err = UpdateRegistry(ctx, ts.URL)
	if err != nil {
		t.Fatalf("UpdateRegistry failed: %v", err)
	}

	// Verify that the file was created and contains the uncompressed data
	dbPath := env.GetRegistryDatabasePath()
	content, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("Failed to read decompressed DB file: %v", err)
	}

	if !bytes.Equal(content, dummyDBContent) {
		t.Errorf("Decompressed content mismatch. Got %d bytes, want %d bytes", len(content), len(dummyDBContent))
	}
}
