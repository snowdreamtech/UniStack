// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"log/slog"

	"github.com/snowdreamtech/unistack/internal/client"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the local package registry database",
	Long:  `Downloads the latest packages.db.zst from the remote registry and updates the local cache.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// In a real application, this URL might be configurable via ~/.unistack.toml or env vars.
		// For now, we will default to a local test server if UNISTACK_REGISTRY_URL is not set,
		// but ideally it should point to the production registry.
		// We'll let the user provide it via an environment variable for testing.

		registryURL := "https://registry.unistack.org"
		if envURL := cmd.Flags().Lookup("url").Value.String(); envURL != "" {
			registryURL = envURL
		}

		slog.Info("Starting registry update", "url", registryURL)

		if err := client.UpdateRegistry(cmd.Context(), registryURL); err != nil {
			slog.Error("Failed to update registry", "error", err)
			return err
		}

		slog.Info("Registry update complete")
		return nil
	},
}

func init() {
	updateCmd.Flags().String("url", "https://registry.unistack.org", "Registry Base URL")
	rootCmd.AddCommand(updateCmd)
}
