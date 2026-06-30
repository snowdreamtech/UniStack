// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/snowdreamtech/unigo/internal/pkg/env"
	"github.com/snowdreamtech/unigo/internal/utils"
	"github.com/spf13/cobra"
)

var dfHumanReadable bool

var dfCmd = &cobra.Command{
	Use:   "df",
	Short: "Display the disk usage of UniGo data directories",
	Long:  `Display the disk usage of various folders within the UniGo data directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir := env.GetDataDir()

		entries, err := os.ReadDir(dataDir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("Data directory does not exist yet.")
				return nil
			}
			return fmt.Errorf("failed to read data directory: %w", err)
		}

		var totalSize int64

		fmt.Printf("Data Directory: %s\n\n", dataDir)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "Directory\tSize\t")
		fmt.Fprintln(w, "---------\t----\t")

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			dirPath := filepath.Join(dataDir, entry.Name())
			size, err := utils.CalculateDirectorySize(dirPath)
			if err != nil {
				continue
			}

			totalSize += size

			sizeStr := fmt.Sprintf("%d", size)
			if dfHumanReadable {
				sizeStr = utils.FormatBytes(size)
			}

			fmt.Fprintf(w, "%s\t%s\t\n", entry.Name(), sizeStr)
		}

		totalStr := fmt.Sprintf("%d", totalSize)
		if dfHumanReadable {
			totalStr = utils.FormatBytes(totalSize)
		}

		fmt.Fprintln(w, "---------\t----\t")
		fmt.Fprintf(w, "TOTAL\t%s\t\n", totalStr)
		w.Flush()

		return nil
	},
}

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(dfCmd)
	}
	dfCmd.Flags().BoolVarP(&dfHumanReadable, "human-readable", "h", false, "print sizes in powers of 1024 (e.g., 1023M)")
}
