// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/snowdreamtech/unistack/internal/registry"
	"github.com/spf13/cobra"
)

var registryCmd = &cobra.Command{
	Use:    "registry",
	Short:  "Manage the local package registry",
	Hidden: true,
}

var registryBuildCmd = &cobra.Command{
	Use:   "build [DIR]",
	Short: "Build the registry SQLite database from a directory of package YAML files",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()

		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		absDir, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("invalid directory path: %w", err)
		}

		slog.Info(fmt.Sprintf("Starting registry build from directory: %s", absDir))

		dbPath := filepath.Join(absDir, "packages.db")
		zstPath := dbPath + ".zst"

		// Initialize registry builder
		builder, err := registry.NewBuilder(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize registry builder: %w", err)
		}
		defer builder.Close()

		ctx := context.Background()
		if err := builder.Build(ctx, absDir); err != nil {
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

		duration := time.Since(start)
		slog.Info(fmt.Sprintf("Registry build completed successfully in %v", duration))
		return nil
	},
}

func init() {
	registryCmd.AddCommand(registryBuildCmd)
	rootCmd.AddCommand(registryCmd)
}
