// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/snowdreamtech/unistack/internal/client"
	"github.com/snowdreamtech/unistack/internal/registry"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <package_or_file>",
	Short: "Install a package",
	Long: `Install a package from the registry or from a local .tar.gz file.

Examples:
  # Install from a local file
  unistack install ./hello-1.0.0.tar.gz

  # Install from registry (not yet fully implemented)
  unistack install hello
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		installer := client.NewInstaller()

		// If it looks like a local file
		if strings.HasSuffix(target, ".tar.gz") {
			if _, err := os.Stat(target); err == nil {
				fmt.Printf("Installing from local file: %s\n", target)
				if err := installer.InstallFromLocal(target); err != nil {
					return fmt.Errorf("installation failed: %w", err)
				}
				fmt.Println("Installation completed successfully.")
				return nil
			}
		}

		// Remote registry installation (Phase 3)
		ctx := context.Background()

		// 1. Query the registry for the package
		meta, err := registry.QueryPackage(ctx, target)
		if err != nil {
			return fmt.Errorf("failed to query package: %w", err)
		}
		if meta == nil {
			return fmt.Errorf("package %q not found in registry", target)
		}

		// 2. Download the package
		downloader := client.NewDownloader()
		// Hardcoded registry URL for now, or from config. Let's use a default localhost for testing or 
		// "http://localhost:8080" as used in downloader tests
		registryURL := "http://localhost:8080" 
		
		fmt.Printf("Downloading %s version %s...\n", meta.Name, meta.Version)
		downloadedPath, err := downloader.DownloadPackage(ctx, registryURL, meta)
		if err != nil {
			return fmt.Errorf("failed to download package: %w", err)
		}

		// 3. Install it
		fmt.Printf("Installing %s...\n", meta.Name)
		if err := installer.InstallFromLocal(downloadedPath); err != nil {
			return fmt.Errorf("installation failed: %w", err)
		}
		
		fmt.Println("Installation completed successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
