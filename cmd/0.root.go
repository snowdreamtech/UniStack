// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"

	"github.com/snowdreamtech/unigo/internal/hello"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	quiet   bool
	debug   bool
	jsonFmt bool
	dryRun  bool
)

var rootCmd = &cobra.Command{
	Use:   "unigo",
	Short: "UniGo is a Golang template hello world application",
	Long:  `A fast and flexible Golang template referencing UniRTM and helloworld.`,
	Run: func(cmd *cobra.Command, args []string) {
		hello.PrintHello()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "enable quiet mode (minimal output)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable verbose debug logging")
	rootCmd.PersistentFlags().BoolVarP(&jsonFmt, "json", "j", false, "enable JSON output format")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would happen without making changes")
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
