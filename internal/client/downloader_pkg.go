// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package client

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/snowdreamtech/unistack/internal/env"
	"github.com/snowdreamtech/unistack/internal/registry"
)

// DownloadPackage downloads a package tarball given its metadata, validates its hash,
// and saves it to the local cache.
func (d *Downloader) DownloadPackage(ctx context.Context, registryBaseURL string, meta *registry.PackageMetadata) (string, error) {
	// Construct the URL. e.g. http://localhost:8080/hello-1.0.0.tar.gz
	pkgFilename := fmt.Sprintf("%s-%s.tar.gz", meta.Name, meta.Version)
	url := fmt.Sprintf("%s/%s", registryBaseURL, pkgFilename)

	slog.Info("Downloading package", "name", meta.Name, "version", meta.Version, "url", url)

	downloadsDir := env.GetDownloadsDir()
	if err := os.MkdirAll(downloadsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create downloads directory: %w", err)
	}

	finalPath := filepath.Join(downloadsDir, pkgFilename)
	tmpPath := finalPath + ".tmp"
	
	// Clean up tmp file in case of failure
	defer os.Remove(tmpPath)

	err := d.Download(ctx, url, func(body io.Reader) error {
		outFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("failed to create temporary package file: %w", err)
		}
		defer outFile.Close()

		if _, err := io.Copy(outFile, body); err != nil {
			return fmt.Errorf("failed to write package data: %w", err)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	// Validate hash if expected hash is provided
	if meta.Hash != "" {
		slog.Debug("Validating package hash", "expected", meta.Hash)
		tmpFile, err := os.Open(tmpPath)
		if err != nil {
			return "", fmt.Errorf("failed to open downloaded file for hashing: %w", err)
		}
		
		err = ValidateFileHash(tmpFile, meta.Hash)
		tmpFile.Close() // Close immediately after hashing
		
		if err != nil {
			// Hash mismatch, tmpPath is removed by defer
			return "", fmt.Errorf("security validation failed for %s: %w", pkgFilename, err)
		}
	} else {
		slog.Warn("No hash provided for package, skipping security validation", "package", meta.Name)
	}

	// Move to final path
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return "", fmt.Errorf("failed to move downloaded package to final path: %w", err)
	}

	slog.Info("Successfully downloaded package", "path", finalPath)
	return finalPath, nil
}
