// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"log/slog"
	
	"github.com/snowdreamtech/unistack/internal/cli/output"
	"github.com/snowdreamtech/unistack/internal/pkg/env"
	"github.com/snowdreamtech/unistack/internal/pkg/errors"
	"github.com/snowdreamtech/unistack/internal/pkg/logger"
	"github.com/snowdreamtech/unistack/internal/pkg/orchestrator"
	"github.com/snowdreamtech/unistack/internal/updater"
	"github.com/spf13/cobra"
)

var (
	quiet       bool
	silent      bool
	verbose     bool
	jsonOutput  bool
	cdDir       string
	yes         bool
	showVersion bool
	playbookFile  string
	inventoryFile string
	pipIndexUrl   string
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
		
		slog.Info("Starting UniStack Fat CLI MVP...")

		workDir, err := orchestrator.ExtractAnsibleFS()
		if err != nil {
			slog.Error("Failed to extract embedded Ansible files", "error", err)
			os.Exit(1)
		}
		slog.Info("Successfully extracted embedded files", "workDir", workDir)

		// Determine Playbook path
		pb := "playbooks/helloworld.yml"
		if playbookFile != "" {
			absPb, err := filepath.Abs(playbookFile)
			if err == nil {
				pb = absPb
			} else {
				pb = playbookFile
			}
		}

		// Determine Inventory path
		inv := "inventory"
		if inventoryFile != "" {
			// If it contains a comma, Ansible treats it as a host list (e.g. "localhost,")
			// Otherwise, treat it as a file/directory and get absolute path
			if !strings.Contains(inventoryFile, ",") {
				absInv, err := filepath.Abs(inventoryFile)
				if err == nil {
					inv = absInv
				} else {
					inv = inventoryFile
				}
			} else {
				inv = inventoryFile
			}
		}

		err = orchestrator.ExecutePlaybook(workDir, pb, inv, pipIndexUrl)
		if err != nil {
			slog.Error("Playbook execution failed", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cdDir, "cd", "C", "", "change directory before running command")
	rootCmd.PersistentFlags().StringVarP(&inventoryFile, "inventory", "i", "", "specify inventory host path or comma separated host list")
	rootCmd.PersistentFlags().StringVarP(&playbookFile, "playbook", "p", "", "specify playbook file path")
	rootCmd.PersistentFlags().StringVar(&pipIndexUrl, "pip-index-url", "https://pypi.org/simple", "specify pip index URL for bootstrapping virtual environment")
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "enable JSON output format")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "enable quiet mode (minimal output)")
	rootCmd.PersistentFlags().BoolVar(&silent, "silent", false, "suppress all task output and non-error messages")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "V", false, "enable verbose output (debug logging)")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "display version information")
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
