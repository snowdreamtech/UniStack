// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var manpageDir string

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate various artifacts",
}

var manpageCmd = &cobra.Command{
	Use:   "manpage",
	Short: "Generate man pages",
	RunE: func(cmd *cobra.Command, args []string) error {
		if manpageDir == "" {
			return fmt.Errorf("output directory is required (use -d/--dir)")
		}
		if err := os.MkdirAll(manpageDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		header := &doc.GenManHeader{
			Title:   "UNISTACK",
			Section: "1",
		}

		err := doc.GenManTree(rootCmd, header, manpageDir)
		if err != nil {
			return fmt.Errorf("failed to generate manpages: %w", err)
		}

		return nil
	},
}

func init() {
	manpageCmd.Flags().StringVarP(&manpageDir, "dir", "d", "", "Output directory for manpages")
	generateCmd.AddCommand(manpageCmd)
	rootCmd.AddCommand(generateCmd)
}
