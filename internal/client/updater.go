// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package client

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/snowdreamtech/unistack/internal/env"
	"github.com/snowdreamtech/unistack/internal/registry"
)

// UpdateRegistry fetches repomd.json, checks for updates, and downloads/extracts the registry database if needed.
func UpdateRegistry(ctx context.Context, registryURL string) error {
	slog.Info("Checking for registry updates...")

	downloader := NewDownloader()

	// Ensure registry cache directory exists
	registryDir := env.GetRegistryCacheDir()
	if err := os.MkdirAll(registryDir, 0755); err != nil {
		return fmt.Errorf("failed to create registry cache directory: %w", err)
	}

	registryURL = strings.TrimSuffix(registryURL, "/")
	repomdURL := registryURL + "/repodata/repomd.json"

	// 1. Fetch remote repomd.json
	var remoteRepomd registry.RepoMd
	err := downloader.Download(ctx, repomdURL, func(body io.Reader) error {
		return json.NewDecoder(body).Decode(&remoteRepomd)
	})
	if err != nil {
		return fmt.Errorf("failed to download repomd.json: %w", err)
	}

	// 2. Read local repomd.json (if exists)
	localRepomdPath := filepath.Join(registryDir, "repomd.json")
	var localRepomd registry.RepoMd
	localBytes, err := os.ReadFile(localRepomdPath)
	if err == nil {
		if err := json.Unmarshal(localBytes, &localRepomd); err == nil {
			if localRepomd.Hash == remoteRepomd.Hash {
				slog.Info("Registry is already up-to-date")
				return nil
			}
		}
	}

	// 3. Download the actual database
	slog.Info("Downloading updated package registry...")
	dbPath := env.GetRegistryDatabasePath()
	zstPath := dbPath + ".zst.tmp"
	defer os.Remove(zstPath)

	zstURL := registryURL + "/" + remoteRepomd.Path

	err = downloader.Download(ctx, zstURL, func(body io.Reader) error {
		outFile, err := os.OpenFile(zstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("failed to create temporary zst file: %w", err)
		}
		defer outFile.Close()

		hasher := sha256.New()
		tee := io.TeeReader(body, hasher)

		if _, err := io.Copy(outFile, tee); err != nil {
			return fmt.Errorf("failed to download registry database: %w", err)
		}

		calculatedHash := "sha256:" + hex.EncodeToString(hasher.Sum(nil))
		if calculatedHash != remoteRepomd.Hash {
			return fmt.Errorf("hash mismatch: expected %s, got %s", remoteRepomd.Hash, calculatedHash)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// 4. Extract the zst file to a temporary database file
	tmpDbPath := dbPath + ".tmp"
	defer os.Remove(tmpDbPath)

	zstFile, err := os.Open(zstPath)
	if err != nil {
		return fmt.Errorf("failed to open downloaded zst file: %w", err)
	}
	defer zstFile.Close()

	decoder, err := zstd.NewReader(zstFile)
	if err != nil {
		return fmt.Errorf("failed to create zstd decoder: %w", err)
	}
	defer decoder.Close()

	outDbFile, err := os.OpenFile(tmpDbPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create temporary database file: %w", err)
	}

	if _, err := io.Copy(outDbFile, decoder); err != nil {
		outDbFile.Close()
		return fmt.Errorf("failed to decompress registry database: %w", err)
	}
	// Explicitly close before rename to avoid Windows file lock issues
	outDbFile.Close()

	// 5. Atomically rename the files
	if err := os.Rename(tmpDbPath, dbPath); err != nil {
		return fmt.Errorf("failed to move updated registry database into place: %w", err)
	}

	// 6. Save the new repomd.json
	remoteBytes, err := json.MarshalIndent(remoteRepomd, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal new repomd.json: %w", err)
	}
	if err := os.WriteFile(localRepomdPath, remoteBytes, 0644); err != nil {
		slog.Warn("Failed to save local repomd.json, update succeeded but next run may re-download", "error", err)
	}

	slog.Info("Successfully updated package registry", "path", dbPath)
	return nil
}
