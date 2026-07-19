// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"

	"github.com/snowdreamtech/unistack/internal/client"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <package_name>",
	Short: "Uninstall a package",
	Long: `Uninstall a package and its associated files and symlinks.

Examples:
  unistack uninstall hello
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pkgName := args[0]
		installer := client.NewInstaller()

		if err := installer.Uninstall(pkgName); err != nil {
			return fmt.Errorf("failed to uninstall package %q: %w", pkgName, err)
		}

		fmt.Printf("Successfully uninstalled %s.\n", pkgName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
