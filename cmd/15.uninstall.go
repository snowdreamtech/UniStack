// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"

	"github.com/snowdreamtech/unistack/internal/client"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <package_id>",
	Short: "Uninstall a package",
	Long: `Uninstall an existing package from the local UniStack environment.

Examples:
  unistack uninstall hello-1.0.0
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pkgID := args[0]
		installer := client.NewInstaller()

		fmt.Printf("Uninstalling %s...\n", pkgID)
		if err := installer.Uninstall(pkgID); err != nil {
			return fmt.Errorf("uninstall failed: %w", err)
		}

		fmt.Println("Uninstallation completed successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
