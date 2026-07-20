// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

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
	"time"

	"github.com/snowdreamtech/unistack/internal/registry"
	"github.com/spf13/cobra"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage the local package registry",
}

var registryBuildCmd = &cobra.Command{
	Use:   "build [SOURCE_DIR] [DEST_DIR]",
	Short: "Build the registry SQLite database from a directory of package archives, auto-arranging them",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()

		sourceDir := "."
		destDir := "."
		if len(args) == 1 {
			sourceDir = args[0]
			destDir = args[0]
		} else if len(args) == 2 {
			sourceDir = args[0]
			destDir = args[1]
		}

		absSource, err := filepath.Abs(sourceDir)
		if err != nil {
			return fmt.Errorf("invalid source path: %w", err)
		}

		absDest, err := filepath.Abs(destDir)
		if err != nil {
			return fmt.Errorf("invalid destination path: %w", err)
		}

		return buildRegistry(absSource, absDest, start)
	},
}

var registryPackCmd = &cobra.Command{
	Use:   "pack [SOURCE_DIR] [DEST_DIR]",
	Short: "Pack source roles into the registry structure and build the database",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()

		sourceDir := "ansible/roles"
		if len(args) > 0 {
			sourceDir = args[0]
		}

		destDir := "public"
		if len(args) > 1 {
			destDir = args[1]
		}

		absSource, err := filepath.Abs(sourceDir)
		if err != nil {
			return fmt.Errorf("invalid source path: %w", err)
		}

		absDest, err := filepath.Abs(destDir)
		if err != nil {
			return fmt.Errorf("invalid destination path: %w", err)
		}

		slog.Info(fmt.Sprintf("Packaging roles from %s to %s", absSource, absDest))

		ctx := context.Background()
		if err := registry.Pack(ctx, absSource, absDest); err != nil {
			return fmt.Errorf("failed to pack roles: %w", err)
		}

		// Proceed to build the registry using the generated packages
		return buildRegistry(absDest, absDest, start)
	},
}

func buildRegistry(absSource, absDest string, start time.Time) error {
	repodataDir := filepath.Join(absDest, "repodata")
	if err := os.MkdirAll(repodataDir, 0755); err != nil {
		return fmt.Errorf("failed to create repodata directory: %w", err)
	}

	slog.Info(fmt.Sprintf("Starting registry build from source %s to %s", absSource, absDest))

	dbPath := filepath.Join(repodataDir, "packages.db")
	zstPath := dbPath + ".zst"

	// Initialize registry builder
	builder, err := registry.NewBuilder(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize registry builder: %w", err)
	}
	defer builder.Close()

	ctx := context.Background()
	if err := builder.Build(ctx, absSource, absDest); err != nil {
		return fmt.Errorf("registry build failed: %w", err)
	}

	// The db file is still open by modernc.org/sqlite, close it before compressing
	builder.Close()

	slog.Info(fmt.Sprintf("Successfully built SQLite database at: %s", dbPath))

	// Compress the database
	slog.Info(fmt.Sprintf("Compressing database to %s...", zstPath))
	if err := registry.CompressZstd(dbPath, zstPath); err != nil {
		return fmt.Errorf("compression failed: %w", err)
	}

	// Calculate hash of zst file
	zstFile, err := os.Open(zstPath)
	if err != nil {
		return fmt.Errorf("failed to open compressed db: %w", err)
	}
	defer zstFile.Close()

	h := sha256.New()
	if _, err := io.Copy(h, zstFile); err != nil {
		return fmt.Errorf("failed to hash compressed db: %w", err)
	}
	zstHash := "sha256:" + hex.EncodeToString(h.Sum(nil))

	// Write repomd.json
	repomdPath := filepath.Join(repodataDir, "repomd.json")
	repomd := registry.RepoMd{
		Timestamp: time.Now().Unix(),
		Hash:      zstHash,
		Path:      "repodata/packages.db.zst",
	}
	repomdBytes, err := json.MarshalIndent(repomd, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode repomd.json: %w", err)
	}
	if err := os.WriteFile(repomdPath, repomdBytes, 0644); err != nil {
		return fmt.Errorf("failed to write repomd.json: %w", err)
	}

	slog.Info(fmt.Sprintf("Generated repomd.json at: %s", repomdPath))

	duration := time.Since(start)
	slog.Info(fmt.Sprintf("Registry build completed successfully in %v", duration))
	return nil
}

func init() {
	registryCmd.AddCommand(registryBuildCmd)
	registryCmd.AddCommand(registryPackCmd)
	rootCmd.AddCommand(registryCmd)
}
