// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"

	"github.com/snowdreamtech/unigo/internal/cli/output"
	"github.com/snowdreamtech/unigo/internal/hello"
	"github.com/snowdreamtech/unigo/internal/pkg/env"
	"github.com/snowdreamtech/unigo/internal/pkg/errors"
	"github.com/snowdreamtech/unigo/internal/pkg/logger"
	"github.com/snowdreamtech/unigo/internal/updater"
	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	quiet       bool
	silent      bool
	noConfig    bool
	noEnv       bool
	noHooks     bool
	verbose     bool
	jsonOutput  bool
	dryRun      bool
	cdDir       string
	yes         bool
	jobs        int
	envName     string
	locked      bool
	showVersion bool
)

func getOutputFormat() output.OutputFormat {
	if jsonOutput {
		return output.FormatJSON
	}
	return output.FormatHuman
}

var rootCmd = &cobra.Command{
	Use:   "unigo",
	Short: "UniGo is a Golang template hello world application",
	Long:  `A fast and flexible Golang template referencing UniRTM and helloworld.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Change directory if --cd is provided
		if cdDir != "" {
			if err := os.Chdir(cdDir); err != nil {
				return errors.NewSystemError(fmt.Sprintf("failed to change directory to %s", cdDir), err)
			}
		}

		// Initialize the global logger before any command runs.
		// If --verbose is set, treat it as debug logging
		logger.Init(verbose, quiet, silent, jsonOutput)

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
		hello.PrintHello()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cdDir, "cd", "C", "", "change directory before running command")
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path (default: .unigo.toml or unigo.toml)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would happen without making changes")
	rootCmd.PersistentFlags().StringVarP(&envName, "env", "E", "", "set the environment for loading configuration")
	rootCmd.PersistentFlags().IntVar(&jobs, "jobs", 8, "how many jobs to run in parallel")
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "enable JSON output format")
	rootCmd.PersistentFlags().BoolVar(&locked, "locked", false, "require lockfile URLs to be present during installation")
	rootCmd.PersistentFlags().BoolVar(&noConfig, "no-config", false, "do not load any config files")
	rootCmd.PersistentFlags().BoolVar(&noEnv, "no-env", false, "do not load environment variables from config files")
	rootCmd.PersistentFlags().BoolVar(&noHooks, "no-hooks", false, "do not execute hooks from config files")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "enable quiet mode (minimal output)")
	rootCmd.PersistentFlags().BoolVar(&silent, "silent", false, "suppress all task output and non-error messages")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "V", false, "enable verbose output (debug logging)")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "display version information")
	rootCmd.PersistentFlags().BoolVarP(&yes, "yes", "y", false, "answer yes to all confirmation prompts")
	
	// Set DisableFlagsInUseLine to match typical Cobra help output
	rootCmd.DisableFlagsInUseLine = true
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(errors.ExitCode(err))
	}
}
