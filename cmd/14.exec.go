// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"log/slog"

	"github.com/snowdreamtech/unistack/internal/pkg/errors"
	"github.com/snowdreamtech/unistack/internal/pkg/orchestrator"
	"github.com/spf13/cobra"
)

var (
	// Exec-specific flags
	execModuleName  string
	execArgs        string
	execBackground  int
	execPoll        int
	execOneLine     bool
	execTree        string
	execPlaybookDir string
	execTaskTimeout int

	// Shared flags (similar to run.go)
	execInventory              string
	execPipIndexUrl            string
	execBecome                 bool
	execBecomeUser             string
	execBecomeMethod           string
	execAskBecomePass          bool
	execBecomePasswordFile     string
	execUser                   string
	execAskPass                bool
	execPrivateKey             string
	execConnection             string
	execTimeout                int
	execConnectionPasswordFile string
	execCheck                  bool
	execDiff                   bool
	execExtraVars              []string
	execFlushCache             bool
	execLimit                  string
	execListHosts              bool
	execForks                  int
	execModulePath             []string
	execVaultId                []string
	execAskVaultPass           bool
	execVaultPasswordFile      string
	execSshCommonArgs          string
	execSshExtraArgs           string
	execScpExtraArgs           string
	execSftpExtraArgs          string
)

var execCmd = &cobra.Command{
	Use:   "exec [pattern] [flags]",
	Short: "Run an Ansible ad-hoc command",
	Long:  `Run an Ansible ad-hoc command in the UniStack isolated environment.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("Initializing UniStack isolated environment for Ad-Hoc command...")
		ctx := context.Background()

		workDir, binary, venvEnv, err := orchestrator.PrepareEnvironment(ctx, execPipIndexUrl)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to initialize environment: %v", err))
			os.Exit(errors.ExitCode(err))
		}

		pattern := args[0]

		// The remaining args (after --) might be caught in args[1:]
		var unmappedArgs []string
		if len(args) > 1 {
			unmappedArgs = args[1:]
		}

		// Assemble explicitly mapped flags
		dynamicArgs := []string{}

		if execInventory != "" {
			dynamicArgs = append(dynamicArgs, "-i", execInventory)
		}

		if execModuleName != "" {
			dynamicArgs = append(dynamicArgs, "-m", execModuleName)
		}
		if execArgs != "" {
			dynamicArgs = append(dynamicArgs, "-a", execArgs)
		}
		if execBackground > 0 {
			dynamicArgs = append(dynamicArgs, "-B", fmt.Sprintf("%d", execBackground))
		}
		if execPoll != 15 { // 15 is default
			dynamicArgs = append(dynamicArgs, "-P", fmt.Sprintf("%d", execPoll))
		}
		if execOneLine {
			dynamicArgs = append(dynamicArgs, "-o")
		}
		if execTree != "" {
			dynamicArgs = append(dynamicArgs, "-t", execTree)
		}
		if execPlaybookDir != "" {
			dynamicArgs = append(dynamicArgs, "--playbook-dir", execPlaybookDir)
		}
		if execTaskTimeout > 0 {
			dynamicArgs = append(dynamicArgs, "--task-timeout", fmt.Sprintf("%d", execTaskTimeout))
		}

		// Map Shared Native Flags
		if execBecome {
			dynamicArgs = append(dynamicArgs, "-b")
		}
		if execBecomeMethod != "" {
			dynamicArgs = append(dynamicArgs, "--become-method", execBecomeMethod)
		}
		if execBecomeUser != "" {
			dynamicArgs = append(dynamicArgs, "--become-user", execBecomeUser)
		}
		if execAskBecomePass {
			dynamicArgs = append(dynamicArgs, "-K")
		}
		if execBecomePasswordFile != "" {
			dynamicArgs = append(dynamicArgs, "--become-password-file", execBecomePasswordFile)
		}
		if execUser != "" {
			dynamicArgs = append(dynamicArgs, "-u", execUser)
		}
		if execAskPass {
			dynamicArgs = append(dynamicArgs, "-k")
		}
		if execPrivateKey != "" {
			dynamicArgs = append(dynamicArgs, "--private-key", execPrivateKey)
		}
		if execConnection != "" {
			dynamicArgs = append(dynamicArgs, "-c", execConnection)
		}
		if execTimeout > 0 {
			dynamicArgs = append(dynamicArgs, "-T", fmt.Sprintf("%d", execTimeout))
		}
		if execConnectionPasswordFile != "" {
			dynamicArgs = append(dynamicArgs, "--connection-password-file", execConnectionPasswordFile)
		}
		if execCheck {
			dynamicArgs = append(dynamicArgs, "-C")
		}
		if execDiff {
			dynamicArgs = append(dynamicArgs, "-D")
		}
		for _, v := range execExtraVars {
			dynamicArgs = append(dynamicArgs, "-e", v)
		}
		if execFlushCache {
			dynamicArgs = append(dynamicArgs, "--flush-cache")
		}
		if execLimit != "" {
			dynamicArgs = append(dynamicArgs, "-l", execLimit)
		}
		if execListHosts {
			dynamicArgs = append(dynamicArgs, "--list-hosts")
		}
		if execForks != 5 { // 5 is ansible default
			dynamicArgs = append(dynamicArgs, "-f", fmt.Sprintf("%d", execForks))
		}
		for _, v := range execModulePath {
			dynamicArgs = append(dynamicArgs, "-M", v)
		}
		for _, v := range execVaultId {
			dynamicArgs = append(dynamicArgs, "--vault-id", v)
		}
		if execAskVaultPass {
			dynamicArgs = append(dynamicArgs, "-J")
		}
		if execVaultPasswordFile != "" {
			dynamicArgs = append(dynamicArgs, "--vault-password-file", execVaultPasswordFile)
		}
		if execSshCommonArgs != "" {
			dynamicArgs = append(dynamicArgs, "--ssh-common-args", execSshCommonArgs)
		}
		if execSshExtraArgs != "" {
			dynamicArgs = append(dynamicArgs, "--ssh-extra-args", execSshExtraArgs)
		}
		if execScpExtraArgs != "" {
			dynamicArgs = append(dynamicArgs, "--scp-extra-args", execScpExtraArgs)
		}
		if execSftpExtraArgs != "" {
			dynamicArgs = append(dynamicArgs, "--sftp-extra-args", execSftpExtraArgs)
		}

		// Map UniStack global verbose (-v) count to Ansible verbose
		if verbose > 0 {
			dynamicArgs = append(dynamicArgs, "-"+strings.Repeat("v", verbose))
		}

		// Append the unmapped args passed after --
		dynamicArgs = append(dynamicArgs, unmappedArgs...)

		if err := orchestrator.ExecuteAdHoc(workDir, pattern, binary, venvEnv, dynamicArgs...); err != nil {
			slog.Error(fmt.Sprintf("Exec command failed: %v", err))
			os.Exit(errors.ExitCode(err))
		}
	},
}

func init() {
	rootCmd.AddCommand(execCmd)

	// Ad-Hoc exclusive flags
	execCmd.Flags().StringVarP(&execModuleName, "module-name", "m", "command", "Name of the action to execute")
	execCmd.Flags().StringVarP(&execArgs, "args", "a", "", "The action's options in space separated k=v format")
	execCmd.Flags().IntVarP(&execBackground, "background", "B", 0, "run asynchronously, failing after X seconds (default=N/A)")
	execCmd.Flags().IntVarP(&execPoll, "poll", "P", 15, "set the poll interval if using -B")
	execCmd.Flags().BoolVarP(&execOneLine, "one-line", "o", false, "condense output")
	execCmd.Flags().StringVarP(&execTree, "tree", "t", "", "log output to this directory")
	execCmd.Flags().StringVar(&execPlaybookDir, "playbook-dir", "", "Since this tool does not use playbooks, use this as a substitute playbook directory")
	execCmd.Flags().IntVar(&execTaskTimeout, "task-timeout", 0, "set task timeout limit in seconds, must be positive integer")

	// Shared native flags
	execCmd.Flags().StringVarP(&execInventory, "inventory", "i", "", "specify inventory host path or comma separated host list")
	execCmd.Flags().StringVar(&execPipIndexUrl, "pip-index-url", "https://pypi.org/simple", "specify pip index URL for bootstrapping virtual environment")
	execCmd.Flags().BoolVarP(&execBecome, "become", "b", false, "run operations with become")
	execCmd.Flags().StringVar(&execBecomeMethod, "become-method", "sudo", "privilege escalation method to use")
	execCmd.Flags().StringVar(&execBecomeUser, "become-user", "root", "run operations as this user")
	execCmd.Flags().BoolVarP(&execAskBecomePass, "ask-become-pass", "K", false, "ask for privilege escalation password")
	execCmd.Flags().StringVar(&execBecomePasswordFile, "become-password-file", "", "become password file")
	execCmd.Flags().StringVarP(&execUser, "user", "u", "", "connect as this user")
	execCmd.Flags().BoolVarP(&execAskPass, "ask-pass", "k", false, "ask for connection password")
	execCmd.Flags().StringVar(&execPrivateKey, "private-key", "", "use this file to authenticate the connection")
	execCmd.Flags().StringVarP(&execConnection, "connection", "c", "ssh", "connection type to use")
	execCmd.Flags().IntVarP(&execTimeout, "timeout", "T", 0, "override the connection timeout in seconds")
	execCmd.Flags().StringVar(&execConnectionPasswordFile, "connection-password-file", "", "connection password file")
	execCmd.Flags().BoolVarP(&execCheck, "check", "C", false, "don't make any changes; instead, try to predict some of the changes that may occur")
	execCmd.Flags().BoolVarP(&execDiff, "diff", "D", false, "when changing (small) files and templates, show the differences in those files")
	execCmd.Flags().StringSliceVarP(&execExtraVars, "extra-vars", "e", []string{}, "set additional variables as key=value or YAML/JSON")
	execCmd.Flags().BoolVar(&execFlushCache, "flush-cache", false, "clear the fact cache for every host in inventory")
	execCmd.Flags().StringVarP(&execLimit, "limit", "l", "", "further limit selected hosts to an additional pattern")
	execCmd.Flags().BoolVar(&execListHosts, "list-hosts", false, "outputs a list of matching hosts; does not execute anything else")
	execCmd.Flags().IntVarP(&execForks, "forks", "f", 5, "specify number of parallel processes to use")
	execCmd.Flags().StringSliceVarP(&execModulePath, "module-path", "M", []string{}, "prepend colon-separated path(s) to module library")
	execCmd.Flags().StringSliceVar(&execVaultId, "vault-id", []string{}, "the vault identity to use")
	execCmd.Flags().BoolVarP(&execAskVaultPass, "ask-vault-pass", "J", false, "ask for vault password")
	execCmd.Flags().StringVar(&execVaultPasswordFile, "vault-password-file", "", "vault password file")
	execCmd.Flags().StringVar(&execSshCommonArgs, "ssh-common-args", "", "specify common arguments to pass to sftp/scp/ssh")
	execCmd.Flags().StringVar(&execSshExtraArgs, "ssh-extra-args", "", "specify extra arguments to pass to ssh only")
	execCmd.Flags().StringVar(&execScpExtraArgs, "scp-extra-args", "", "specify extra arguments to pass to scp only")
	execCmd.Flags().StringVar(&execSftpExtraArgs, "sftp-extra-args", "", "specify extra arguments to pass to sftp only")
}
