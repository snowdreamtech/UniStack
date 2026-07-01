// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"

	"github.com/snowdreamtech/unigo/internal/pkg/env"
	"github.com/snowdreamtech/unigo/internal/pkg/errors"
	"github.com/snowdreamtech/unigo/internal/utils"
	"github.com/spf13/cobra"
)

var dfCmd = &cobra.Command{
	Use:   "df",
	Short: "Display disk usage for UniGo directories",
	Long:  `Show the size of the UniGo configuration, data, and cache directories.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dirs := []struct {
			Name string
			Path string
		}{
			{"Config", env.GetConfigDir()},
			{"Data", env.GetDataDir()},
			{"Cache", env.GetCacheDir()},
		}

		fmt.Printf("%-10s %-50s %s\n", "Type", "Path", "Size")
		fmt.Println("-------------------------------------------------------------------------")

		var totalSize int64
		for _, dir := range dirs {
			if _, err := os.Stat(dir.Path); os.IsNotExist(err) {
				continue
			}

			size, err := utils.CalculateDirectorySize(dir.Path)
			if err != nil {
				return errors.NewSystemError(fmt.Sprintf("failed to calculate size for %s", dir.Path), err)
			}

			totalSize += size
			fmt.Printf("%-10s %-50s %s\n", dir.Name, dir.Path, utils.FormatBytes(size))
		}

		fmt.Println("-------------------------------------------------------------------------")
		fmt.Printf("%-61s %s\n", "Total:", utils.FormatBytes(totalSize))

		return nil
	},
}

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(dfCmd)
	}
}
