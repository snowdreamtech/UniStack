// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/snowdreamtech/unistack/internal/client"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all installed packages",
	Long: `List all installed packages currently available in the local UniStack environment.

Examples:
  unistack list
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		installer := client.NewInstaller()

		pkgs, err := installer.ListInstalledPackages()
		if err != nil {
			return fmt.Errorf("failed to list installed packages: %w", err)
		}

		if len(pkgs) == 0 {
			fmt.Println("No packages installed.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
		fmt.Fprintln(w, "PACKAGE\tVERSION")

		for _, pkg := range pkgs {
			// Extract name and version from package ID (e.g. hello-1.0.0)
			// we assume the last hyphen separates name and version
			// if there is no hyphen, the version is unknown
			// To keep it simple, we just print the raw ID for now
			fmt.Fprintf(w, "%s\t\n", pkg)
		}

		w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
