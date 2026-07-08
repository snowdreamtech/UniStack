// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/snowdreamtech/unistack/internal/pkg/orchestrator"
	"github.com/spf13/cobra"
)

var (
	playbookFile  string
	inventoryFile string
	pipIndexUrl   string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run an Ansible playbook",
	Long:  `Run an Ansible playbook in the UniStack isolated environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("Starting UniStack Fat CLI MVP...")

		ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Minute)
		defer cancel()

		workDir, binary, venvEnv, err := orchestrator.PrepareEnvironment(ctx, pipIndexUrl)
		if err != nil {
			slog.Error("Failed to initialize UniStack environment", "error", err)
			os.Exit(1)
		}
		slog.Info("Successfully initialized UniStack environment", "workDir", workDir)

		pb := "playbooks/helloworld.yml"
		if playbookFile != "" {
			absPb, err := filepath.Abs(playbookFile)
			if err == nil {
				pb = absPb
			} else {
				pb = playbookFile
			}
		}

		inv := "inventory"
		if inventoryFile != "" {
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

		err = orchestrator.ExecutePlaybook(workDir, pb, inv, binary, venvEnv)
		if err != nil {
			slog.Error("Playbook execution failed", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&inventoryFile, "inventory", "i", "", "specify inventory host path or comma separated host list")
	runCmd.Flags().StringVarP(&playbookFile, "playbook", "p", "", "specify playbook file path")
	runCmd.Flags().StringVar(&pipIndexUrl, "pip-index-url", "https://pypi.org/simple", "specify pip index URL for bootstrapping virtual environment")
}
