// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unistack/internal/pkg/env"
	"github.com/snowdreamtech/unistack/internal/sysinfo"
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
	vars := []struct {
		Name  string
		Value string
	}{
		{fmt.Sprintf("%s_CONFIG_DIR", strings.ToUpper(env.ProjectName)), env.GetConfigDir()},
		{fmt.Sprintf("%s_DATA_DIR", strings.ToUpper(env.ProjectName)), env.GetDataDir()},
		{fmt.Sprintf("%s_CACHE_DIR", strings.ToUpper(env.ProjectName)), env.GetCacheDir()},
		{fmt.Sprintf("%s_IS_MUSL", strings.ToUpper(env.ProjectName)), fmt.Sprintf("%t", sysinfo.IsMusl())},
	}

	isTerminal := false
	if stat, err := os.Stdout.Stat(); err == nil {
		isTerminal = (stat.Mode() & os.ModeCharDevice) != 0
	}

	if isTerminal {
		pterm.DefaultSection.Println("🔑 Environment Variables")
		var data [][]string
		data = append(data, []string{"Variable", "Value"})
		for _, v := range vars {
			data = append(data, []string{
				pterm.Bold.Sprint(v.Name),
				pterm.LightCyan(v.Value),
			})
		}
		pterm.DefaultTable.WithHasHeader().WithData(data).Render()
		fmt.Println()
		pterm.Info.Println("To apply this environment, run: " + pterm.LightMagenta("eval \"$(unistack env)\""))
	} else {
		for _, v := range vars {
			fmt.Printf("export %s=%q\n", v.Name, v.Value)
		}
	}

	return nil
}
