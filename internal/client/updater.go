// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package client

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/klauspost/compress/zstd"
	"github.com/snowdreamtech/unistack/internal/env"
)

// UpdateRegistry downloads the latest packages.db.zst and extracts it to packages.db
func UpdateRegistry(ctx context.Context, registryURL string) error {
	slog.Info("Updating package registry...")

	downloader := NewDownloader()

	// Ensure registry cache directory exists
	registryDir := env.GetRegistryCacheDir()
	if err := os.MkdirAll(registryDir, 0755); err != nil {
		return fmt.Errorf("failed to create registry cache directory: %w", err)
	}

	dbPath := env.GetRegistryDatabasePath()

	// Create a temporary file for atomic write
	tmpPath := dbPath + ".tmp"
	defer os.Remove(tmpPath)

	err := downloader.Download(ctx, registryURL, func(body io.Reader) error {
		// Set up zstd reader
		decoder, err := zstd.NewReader(body)
		if err != nil {
			return fmt.Errorf("failed to create zstd decoder: %w", err)
		}
		defer decoder.Close()

		outFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("failed to create temporary database file: %w", err)
		}
		defer outFile.Close()

		if _, err := io.Copy(outFile, decoder); err != nil {
			return fmt.Errorf("failed to decompress registry database: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Rename temp file to actual file atomically
	if err := os.Rename(tmpPath, dbPath); err != nil {
		return fmt.Errorf("failed to move updated registry database into place: %w", err)
	}

	slog.Info("Successfully updated package registry", "path", dbPath)
	return nil
}
