// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"log/slog"

	"github.com/snowdreamtech/unistack/internal/client"
	"github.com/snowdreamtech/unistack/internal/registry"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download [package_name]",
	Short: "Download a package for offline installation",
	Long:  `Queries the local registry database for the package metadata and downloads its .tar.gz tarball to the local cache.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pkgName := args[0]

		registryBaseURL := "http://localhost:8080"
		if envURL := cmd.Flags().Lookup("url").Value.String(); envURL != "" {
			registryBaseURL = envURL
		}

		slog.Info("Querying local registry for package", "package", pkgName)

		meta, err := registry.QueryPackage(cmd.Context(), pkgName)
		if err != nil {
			return fmt.Errorf("failed to resolve package metadata: %w", err)
		}

		slog.Info("Found package metadata", "version", meta.Version, "hash", meta.Hash)

		downloader := client.NewDownloader()
		finalPath, err := downloader.DownloadPackage(cmd.Context(), registryBaseURL, meta)
		if err != nil {
			return fmt.Errorf("download failed: %w", err)
		}

		fmt.Printf("Successfully downloaded %s to %s\n", pkgName, finalPath)
		return nil
	},
}

func init() {
	// For testing, assuming the URL serves packages at the root directory
	downloadCmd.Flags().String("url", "http://localhost:8080", "Base URL for the registry files")
	rootCmd.AddCommand(downloadCmd)
}
