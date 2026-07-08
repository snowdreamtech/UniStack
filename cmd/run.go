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
	limit                   string
	tags                    []string
	skipTags                []string
	extraVars               []string
	become                  bool
	askBecomePass           bool
	user                    string
	askPass                 bool
	privateKey              string
	check                   bool
	diff                    bool
	syntaxCheck             bool
	forks                   int
	vaultPasswordFile       string
	becomePasswordFile      string
	connectionPasswordFile  string
	flushCache              bool
	forceHandlers           bool
	listHosts               bool
	listTags                bool
	listTasks               bool
	startAtTask             string
	step                    bool
	vaultId                 []string
	askVaultPass            bool
	modulePath              []string
	connection              string
	timeout                 int
	sshCommonArgs           string
	sftpExtraArgs           string
	scpExtraArgs            string
	sshExtraArgs            string
	becomeMethod            string
	becomeUser              string
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
		if becomePasswordFile != "" {
			dynamicArgs = append(dynamicArgs, "--become-password-file", becomePasswordFile)
		}
		if connectionPasswordFile != "" {
			dynamicArgs = append(dynamicArgs, "--connection-password-file", connectionPasswordFile)
		}
		if flushCache {
			dynamicArgs = append(dynamicArgs, "--flush-cache")
		}
		if forceHandlers {
			dynamicArgs = append(dynamicArgs, "--force-handlers")
		}
		if listHosts {
			dynamicArgs = append(dynamicArgs, "--list-hosts")
		}
		if listTags {
			dynamicArgs = append(dynamicArgs, "--list-tags")
		}
		if listTasks {
			dynamicArgs = append(dynamicArgs, "--list-tasks")
		}
		if startAtTask != "" {
			dynamicArgs = append(dynamicArgs, "--start-at-task", startAtTask)
		}
		if step {
			dynamicArgs = append(dynamicArgs, "--step")
		}
		for _, vid := range vaultId {
			dynamicArgs = append(dynamicArgs, "--vault-id", vid)
		}
		if askVaultPass {
			dynamicArgs = append(dynamicArgs, "--ask-vault-pass")
		}
		for _, mp := range modulePath {
			dynamicArgs = append(dynamicArgs, "--module-path", mp)
		}
		if connection != "" {
			dynamicArgs = append(dynamicArgs, "--connection", connection)
		}
		if timeout != 0 {
			dynamicArgs = append(dynamicArgs, "--timeout", fmt.Sprintf("%d", timeout))
		}
		if sshCommonArgs != "" {
			dynamicArgs = append(dynamicArgs, "--ssh-common-args", sshCommonArgs)
		}
		if sftpExtraArgs != "" {
			dynamicArgs = append(dynamicArgs, "--sftp-extra-args", sftpExtraArgs)
		}
		if scpExtraArgs != "" {
			dynamicArgs = append(dynamicArgs, "--scp-extra-args", scpExtraArgs)
		}
		if sshExtraArgs != "" {
			dynamicArgs = append(dynamicArgs, "--ssh-extra-args", sshExtraArgs)
		}
		if becomeMethod != "" {
			dynamicArgs = append(dynamicArgs, "--become-method", becomeMethod)
		}
		if becomeUser != "" {
			dynamicArgs = append(dynamicArgs, "--become-user", becomeUser)
		}
		
		// Map UniStack global verbose (-V) to Ansible verbose (-vvv)
		if verbose {
			dynamicArgs = append(dynamicArgs, "-vvv")
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
	runCmd.Flags().StringVar(&becomePasswordFile, "become-password-file", "", "become password file")
	runCmd.Flags().StringVar(&connectionPasswordFile, "connection-password-file", "", "connection password file")
	runCmd.Flags().BoolVar(&flushCache, "flush-cache", false, "clear the fact cache for every host in inventory")
	runCmd.Flags().BoolVar(&forceHandlers, "force-handlers", false, "run handlers even if a task fails")
	runCmd.Flags().BoolVar(&listHosts, "list-hosts", false, "outputs a list of matching hosts; does not execute anything else")
	runCmd.Flags().BoolVar(&listTags, "list-tags", false, "list all available tags")
	runCmd.Flags().BoolVar(&listTasks, "list-tasks", false, "list all tasks that would be executed")
	runCmd.Flags().StringVar(&startAtTask, "start-at-task", "", "start the playbook at the task matching this name")
	runCmd.Flags().BoolVar(&step, "step", false, "one-step-at-a-time: confirm each task before running")
	runCmd.Flags().StringSliceVar(&vaultId, "vault-id", []string{}, "the vault identity to use")
	runCmd.Flags().BoolVarP(&askVaultPass, "ask-vault-pass", "J", false, "ask for vault password")
	runCmd.Flags().StringSliceVarP(&modulePath, "module-path", "M", []string{}, "prepend colon-separated path(s) to module library")
	runCmd.Flags().StringVarP(&connection, "connection", "c", "", "connection type to use (default=ssh)")
	runCmd.Flags().IntVarP(&timeout, "timeout", "T", 0, "override the connection timeout in seconds")
	runCmd.Flags().StringVar(&sshCommonArgs, "ssh-common-args", "", "specify common arguments to pass to sftp/scp/ssh")
	runCmd.Flags().StringVar(&sftpExtraArgs, "sftp-extra-args", "", "specify extra arguments to pass to sftp only")
	runCmd.Flags().StringVar(&scpExtraArgs, "scp-extra-args", "", "specify extra arguments to pass to scp only")
	runCmd.Flags().StringVar(&sshExtraArgs, "ssh-extra-args", "", "specify extra arguments to pass to ssh only")
	runCmd.Flags().StringVar(&becomeMethod, "become-method", "", "privilege escalation method to use (default=sudo)")
	runCmd.Flags().StringVar(&becomeUser, "become-user", "", "run operations as this user (default=root)")
}
