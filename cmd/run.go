// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/snowdreamtech/unistack/internal/pkg/orchestrator"
	"github.com/spf13/cobra"
)

var (
	playbookFile      string
	inventoryFile     string
	pipIndexUrl       string
	limit             string
	tags              []string
	skipTags          []string
	extraVars         []string
	become            bool
	askBecomePass     bool
	user              string
	askPass           bool
	privateKey        string
	check             bool
	diff              bool
	syntaxCheck       bool
	forks             int
	vaultPasswordFile string
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

		var dynamicArgs []string
		if limit != "" {
			dynamicArgs = append(dynamicArgs, "--limit", limit)
		}
		for _, tag := range tags {
			dynamicArgs = append(dynamicArgs, "--tags", tag)
		}
		for _, skipTag := range skipTags {
			dynamicArgs = append(dynamicArgs, "--skip-tags", skipTag)
		}
		for _, ev := range extraVars {
			dynamicArgs = append(dynamicArgs, "--extra-vars", ev)
		}
		if become {
			dynamicArgs = append(dynamicArgs, "--become")
		}
		if askBecomePass {
			dynamicArgs = append(dynamicArgs, "--ask-become-pass")
		}
		if user != "" {
			dynamicArgs = append(dynamicArgs, "--user", user)
		}
		if askPass {
			dynamicArgs = append(dynamicArgs, "--ask-pass")
		}
		if privateKey != "" {
			dynamicArgs = append(dynamicArgs, "--private-key", privateKey)
		}
		if check {
			dynamicArgs = append(dynamicArgs, "--check")
		}
		if diff {
			dynamicArgs = append(dynamicArgs, "--diff")
		}
		if syntaxCheck {
			dynamicArgs = append(dynamicArgs, "--syntax-check")
		}
		if forks != 5 {
			dynamicArgs = append(dynamicArgs, "--forks", fmt.Sprintf("%d", forks))
		}
		if vaultPasswordFile != "" {
			dynamicArgs = append(dynamicArgs, "--vault-password-file", vaultPasswordFile)
		}

		// Append the explicitly mapped flags first, then append any unmapped args passed after --
		dynamicArgs = append(dynamicArgs, args...)

		err = orchestrator.ExecutePlaybook(workDir, pb, inv, binary, venvEnv, dynamicArgs...)
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

	// Ansible-playbook flag mapping
	runCmd.Flags().StringVarP(&limit, "limit", "l", "", "further limit selected hosts to an additional pattern")
	runCmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "only run plays and tasks tagged with these values")
	runCmd.Flags().StringSliceVar(&skipTags, "skip-tags", []string{}, "only run plays and tasks whose tags do not match these values")
	runCmd.Flags().StringSliceVarP(&extraVars, "extra-vars", "e", []string{}, "set additional variables as key=value or YAML/JSON, if filename prepend with @")
	runCmd.Flags().BoolVarP(&become, "become", "b", false, "run operations with become (does not imply password prompting)")
	runCmd.Flags().BoolVarP(&askBecomePass, "ask-become-pass", "K", false, "ask for privilege escalation password")
	runCmd.Flags().StringVarP(&user, "user", "u", "", "connect as this user")
	runCmd.Flags().BoolVarP(&askPass, "ask-pass", "k", false, "ask for connection password")
	runCmd.Flags().StringVar(&privateKey, "private-key", "", "use this file to authenticate the connection")
	runCmd.Flags().BoolVar(&check, "check", false, "don't make any changes; instead, try to predict some of the changes that may occur")
	runCmd.Flags().BoolVarP(&diff, "diff", "D", false, "when changing (small) files and templates, show the differences in those files")
	runCmd.Flags().BoolVar(&syntaxCheck, "syntax-check", false, "perform a syntax check on the playbook, but do not execute it")
	runCmd.Flags().IntVarP(&forks, "forks", "f", 5, "specify number of parallel processes to use")
	runCmd.Flags().StringVar(&vaultPasswordFile, "vault-password-file", "", "vault password file")
}
