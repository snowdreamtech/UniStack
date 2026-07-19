// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package cmd

import (
	"fmt"
	"log/slog"

	"github.com/snowdreamtech/unistack/internal/client"
	"github.com/snowdreamtech/unistack/internal/config"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the local package registry database",
	Long:  `Downloads the latest packages.db.zst from the remote registry and updates the local cache.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		urlFlag := cmd.Flags().Lookup("url").Value.String()
		nameFlag := cmd.Flags().Lookup("name").Value.String()

		if nameFlag != "" {
			sources, err := config.LoadSources()
			if err != nil {
				return fmt.Errorf("failed to load sources: %w", err)
			}
			
			var targetURL string
			for _, s := range sources {
				if s.Name == nameFlag {
					targetURL = s.URL
					break
				}
			}
			
			if targetURL == "" {
				return fmt.Errorf("source '%s' not found in configuration", nameFlag)
			}
			
			slog.Info("Starting update for specific source", "name", nameFlag, "url", targetURL)
			if err := client.UpdateSource(cmd.Context(), nameFlag, targetURL); err != nil {
				slog.Error("Failed to update source", "name", nameFlag, "error", err)
				return err
			}
			return nil
		}

		if urlFlag != "" {
			// One-off url flag overriding
			name := "default"
			if nameFlag != "" {
				name = nameFlag
			}
			slog.Info("Starting single-source registry update", "url", urlFlag)
			if err := client.UpdateSource(cmd.Context(), name, urlFlag); err != nil {
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
	updateCmd.Flags().String("url", "", "One-off registry base URL to update from")
	updateCmd.Flags().String("name", "", "Update a specific registry source by name")
	rootCmd.AddCommand(updateCmd)
}
