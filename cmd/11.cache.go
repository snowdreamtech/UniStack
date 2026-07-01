// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(cacheCmd)
		cacheCmd.AddCommand(cacheClearCmd)
	}
}

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage UniGo cache",
	Long:  "Query or clear cached files downloaded by the application.",
}

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all cached data",
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Debug("Clearing cache...")
		// Future implementation: call env.GetCacheDir() and os.RemoveAll()
		fmt.Println("Placeholder: Cache cleared successfully.")
		return nil
	},
}
