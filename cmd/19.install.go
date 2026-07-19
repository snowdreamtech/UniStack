// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/snowdreamtech/unistack/internal/client"
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
		registryURL := "http://localhost:8080" // Default for now

		if err := installer.InstallPackage(ctx, target, registryURL); err != nil {
			if err.Error() == client.ErrCircularDependency.Error() {
				return fmt.Errorf("installation aborted: %w", err)
			}
			return fmt.Errorf("installation failed: %w", err)
		}

		fmt.Println("Installation completed successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
