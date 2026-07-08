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

var applyCmd = &cobra.Command{
	Use:   "apply [playbook]",
	Short: "Apply an Ansible playbook",
	Long:  `Apply an Ansible playbook in the UniStack isolated environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("Starting UniStack Fat CLI MVP...")

		ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Minute)
		defer cancel()

		workDir, binary, venvEnv, err := orchestrator.PrepareEnvironment(ctx, pipIndexUrl)
		if err != nil {
			slog.Error("Failed to initialize UniStack environment", "error", err)
			os.Exit(1)
		}
		slog.Debug("Successfully initialized UniStack environment", "workDir", workDir)

		var pb string
		if playbookFile != "" {
			absPb, err := filepath.Abs(playbookFile)
			if err == nil {
				pb = absPb
			} else {
				pb = playbookFile
			}
		} else if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
			absPb, err := filepath.Abs(args[0])
			if err == nil {
				pb = absPb
			} else {
				pb = args[0]
			}
			args = args[1:]
		} else {
			// Try to find a default playbook in current directory
			defaults := []string{"site.yml", "main.yml", "playbook.yml", "healthcheck.yml", "ansible/playbooks/healthcheck.yml"}
			for _, def := range defaults {
				if _, err := os.Stat(def); err == nil {
					absPb, _ := filepath.Abs(def)
					pb = absPb
					break
				}
			}
		}

		if pb == "" {
			slog.Error("No playbook specified. Please provide a playbook file (e.g. unistack apply site.yml) or use -p/--playbook")
			os.Exit(1)
		}

		inv := ""
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
		
		// Map UniStack global verbose (-v) count to Ansible verbose
		if verbose > 0 {
			dynamicArgs = append(dynamicArgs, "-"+strings.Repeat("v", verbose))
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
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().StringVarP(&inventoryFile, "inventory", "i", "", "specify inventory host path or comma separated host list")
	applyCmd.Flags().StringVarP(&playbookFile, "playbook", "p", "", "specify playbook file path")
	applyCmd.Flags().StringVar(&pipIndexUrl, "pip-index-url", "https://pypi.org/simple", "specify pip index URL for bootstrapping virtual environment")

	// Ansible-playbook flag mapping
	applyCmd.Flags().StringVarP(&limit, "limit", "l", "", "further limit selected hosts to an additional pattern")
	applyCmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "only run plays and tasks tagged with these values")
	applyCmd.Flags().StringSliceVar(&skipTags, "skip-tags", []string{}, "only run plays and tasks whose tags do not match these values")
	applyCmd.Flags().StringSliceVarP(&extraVars, "extra-vars", "e", []string{}, "set additional variables as key=value or YAML/JSON, if filename prepend with @")
	applyCmd.Flags().BoolVarP(&become, "become", "b", false, "run operations with become (does not imply password prompting)")
	applyCmd.Flags().BoolVarP(&askBecomePass, "ask-become-pass", "K", false, "ask for privilege escalation password")
	applyCmd.Flags().StringVarP(&user, "user", "u", "", "connect as this user")
	applyCmd.Flags().BoolVarP(&askPass, "ask-pass", "k", false, "ask for connection password")
	applyCmd.Flags().StringVar(&privateKey, "private-key", "", "use this file to authenticate the connection")
	applyCmd.Flags().BoolVarP(&check, "check", "C", false, "don't make any changes; instead, try to predict some of the changes that may occur")
	applyCmd.Flags().BoolVarP(&diff, "diff", "D", false, "when changing (small) files and templates, show the differences in those files")
	applyCmd.Flags().BoolVar(&syntaxCheck, "syntax-check", false, "perform a syntax check on the playbook, but do not execute it")
	applyCmd.Flags().IntVarP(&forks, "forks", "f", 5, "specify number of parallel processes to use")
	applyCmd.Flags().StringVar(&vaultPasswordFile, "vault-password-file", "", "vault password file")
	applyCmd.Flags().StringVar(&becomePasswordFile, "become-password-file", "", "become password file")
	applyCmd.Flags().StringVar(&connectionPasswordFile, "connection-password-file", "", "connection password file")
	applyCmd.Flags().BoolVar(&flushCache, "flush-cache", false, "clear the fact cache for every host in inventory")
	applyCmd.Flags().BoolVar(&forceHandlers, "force-handlers", false, "run handlers even if a task fails")
	applyCmd.Flags().BoolVar(&listHosts, "list-hosts", false, "outputs a list of matching hosts; does not execute anything else")
	applyCmd.Flags().BoolVar(&listTags, "list-tags", false, "list all available tags")
	applyCmd.Flags().BoolVar(&listTasks, "list-tasks", false, "list all tasks that would be executed")
	applyCmd.Flags().StringVar(&startAtTask, "start-at-task", "", "start the playbook at the task matching this name")
	applyCmd.Flags().BoolVar(&step, "step", false, "one-step-at-a-time: confirm each task before running")
	applyCmd.Flags().StringSliceVar(&vaultId, "vault-id", []string{}, "the vault identity to use")
	applyCmd.Flags().BoolVarP(&askVaultPass, "ask-vault-pass", "J", false, "ask for vault password")
	applyCmd.Flags().StringSliceVarP(&modulePath, "module-path", "M", []string{}, "prepend colon-separated path(s) to module library")
	applyCmd.Flags().StringVarP(&connection, "connection", "c", "", "connection type to use (default=ssh)")
	applyCmd.Flags().IntVarP(&timeout, "timeout", "T", 0, "override the connection timeout in seconds")
	applyCmd.Flags().StringVar(&sshCommonArgs, "ssh-common-args", "", "specify common arguments to pass to sftp/scp/ssh")
	applyCmd.Flags().StringVar(&sftpExtraArgs, "sftp-extra-args", "", "specify extra arguments to pass to sftp only")
	applyCmd.Flags().StringVar(&scpExtraArgs, "scp-extra-args", "", "specify extra arguments to pass to scp only")
	applyCmd.Flags().StringVar(&sshExtraArgs, "ssh-extra-args", "", "specify extra arguments to pass to ssh only")
	applyCmd.Flags().StringVar(&becomeMethod, "become-method", "", "privilege escalation method to use (default=sudo)")
	applyCmd.Flags().StringVar(&becomeUser, "become-user", "", "run operations as this user (default=root)")
}
