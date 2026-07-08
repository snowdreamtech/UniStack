// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"

	"github.com/snowdreamtech/unistack/internal/cli/output"
	"github.com/snowdreamtech/unistack/internal/pkg/env"
	"github.com/snowdreamtech/unistack/internal/pkg/errors"
	"github.com/snowdreamtech/unistack/internal/pkg/logger"
	"github.com/snowdreamtech/unistack/internal/updater"
	"github.com/spf13/cobra"
)

var (
	quiet       bool
	silent      bool
	verbose     int
	jsonOutput  bool
	cdDir       string
	yes         bool
	showVersion bool
)

func getOutputFormat() output.OutputFormat {
	if jsonOutput {
		return output.FormatJSON
	}
	return output.FormatHuman
}

var rootCmd = &cobra.Command{
	Use:   "unistack",
	Short: "UniStack is a Golang template hello world application",
	Long:  `A fast and flexible Golang template referencing UniRTM and helloworld.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Change directory if --cd is provided
		if cdDir != "" {
			if err := os.Chdir(cdDir); err != nil {
				return errors.NewSystemError(fmt.Sprintf("failed to change directory to %s", cdDir), err)
			}
		}

		// Initialize the global logger before any command runs.
		// If --verbose (-v) is used at least once, treat it as debug logging for UniStack
		logger.Init(verbose > 0, quiet, silent, jsonOutput)

		// Asynchronously check for a newer version (non-blocking).
		updater.CheckUpdateAsync(env.GitTag)
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		// Prompt if a newer version is available (once per day, TTY only).
		updater.PromptIfAvailable(env.GitTag, cmd.Name())
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if showVersion {
			runVersion(cmd, args)
			return
		}

		cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cdDir, "cd", "", "change directory before running command")
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "enable JSON output format")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "enable quiet mode (minimal output)")
	rootCmd.PersistentFlags().BoolVar(&silent, "silent", false, "suppress all task output and non-error messages")
	rootCmd.PersistentFlags().CountVarP(&verbose, "verbose", "v", "enable verbose output (use -v, -vv, -vvv, etc. for more detail)")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "V", false, "display version information")
	rootCmd.PersistentFlags().BoolVarP(&yes, "yes", "y", false, "answer yes to all confirmation prompts")
	rootCmd.PersistentFlags().Bool("help", false, "help for this command")

	// Set DisableFlagsInUseLine to match typical Cobra help output
	rootCmd.DisableFlagsInUseLine = true
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(errors.ExitCode(err))
	}
}
