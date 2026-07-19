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
		urlFlag := cmd.Flags().Lookup("url").Value.String()
		if urlFlag != "https://registry.unistack.org" && urlFlag != "" {
			slog.Info("Starting single-source registry update", "url", urlFlag)
			if err := client.UpdateSource(cmd.Context(), "default", urlFlag); err != nil {
				slog.Error("Failed to update registry", "error", err)
				return err
			}
			return nil
		}

		slog.Info("Starting registry update from all configured sources")
		if err := client.UpdateRegistry(cmd.Context()); err != nil {
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
