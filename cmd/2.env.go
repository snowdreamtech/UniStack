// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"strings"

	"github.com/snowdreamtech/unigo/internal/pkg/env"
	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(envCmd)
	}
}

var envCmd = &cobra.Command{
	Use:     "env",
	Short:   "Export generic shell environment variables",
	Long:    `Display or export the environment variables for the current context.`,
	Aliases: []string{"e"},
	Args:    cobra.NoArgs,
	RunE:    runEnv,
}

func runEnv(cmd *cobra.Command, args []string) error {
	vars := []string{
		fmt.Sprintf("%s_CONFIG_DIR=%s", strings.ToUpper(env.ProjectName), env.GetConfigDir()),
		fmt.Sprintf("%s_DATA_DIR=%s", strings.ToUpper(env.ProjectName), env.GetDataDir()),
		fmt.Sprintf("%s_CACHE_DIR=%s", strings.ToUpper(env.ProjectName), env.GetCacheDir()),
	}

	for _, v := range vars {
		fmt.Println(v)
	}

	return nil
}
