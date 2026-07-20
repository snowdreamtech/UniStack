// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/snowdreamtech/unistack/internal/client"
	"github.com/snowdreamtech/unistack/internal/registry"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade <package_name>",
	Short: "Upgrade a package to the latest version",
	Long: `Upgrade a package to the latest version available in the registry.

Examples:
  unistack upgrade hello
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pkgName := args[0]
		installer := client.NewInstaller()

		pkgs, err := installer.ListInstalledPackages()
		if err != nil {
			return fmt.Errorf("failed to list installed packages: %w", err)
		}

		var installedVersionStr string
		var isInstalled bool
		for _, p := range pkgs {
			if p.Metadata.Name == pkgName {
				isInstalled = true
				installedVersionStr = p.Metadata.Version
				break
			}
		}

		if !isInstalled {
			return fmt.Errorf("package %q is not currently installed", pkgName)
		}

		// Query registry for latest
		ctx := context.Background()
		meta, err := registry.QueryPackage(ctx, pkgName)
		if err != nil {
			return fmt.Errorf("failed to query package: %w", err)
		}
		if meta == nil {
			return fmt.Errorf("package %q not found in registry", pkgName)
		}

		vInstalled, err := semver.NewVersion(installedVersionStr)
		if err != nil {
			return fmt.Errorf("invalid installed version %q: %w", installedVersionStr, err)
		}

		vLatest, err := semver.NewVersion(meta.Version)
		if err != nil {
			return fmt.Errorf("invalid registry version %q: %w", meta.Version, err)
		}

		if vLatest.LessThan(vInstalled) || vLatest.Equal(vInstalled) {
			fmt.Printf("Package %s is already up-to-date (version %s).\n", pkgName, installedVersionStr)
			return nil
		}

		fmt.Printf("Upgrading %s from %s to %s...\n", pkgName, installedVersionStr, meta.Version)

		// 1. Uninstall old version
		fmt.Printf("Uninstalling old version...\n")
		if err := installer.Uninstall(pkgName); err != nil {
			return fmt.Errorf("failed to uninstall old version: %w", err)
		}

		// 2. Install new version using the unified installer logic
		fmt.Printf("Installing %s...\n", meta.Name)
		if err := installer.InstallPackage(ctx, pkgName); err != nil {
			return fmt.Errorf("failed to install new package: %w", err)
		}

		fmt.Println("Upgrade completed successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
