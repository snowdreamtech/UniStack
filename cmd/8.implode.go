// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unigo/internal/pkg/env"
	"github.com/spf13/cobra"
)

var (
	implodeYes    bool
	implodeConfig bool
)

func init() {
	implodeCmd.Flags().BoolVarP(&implodeYes, "yes", "y", false, "skip confirmation prompt")
	implodeCmd.Flags().BoolVar(&implodeConfig, "config", false, "also remove configuration directory (~/.config/unigo)")

	if rootCmd != nil {
		rootCmd.AddCommand(implodeCmd)
	}
}

// implodeCmd removes all UniGo data, cache, and configuration.
var implodeCmd = &cobra.Command{
	Use:   "implode",
	Short: "Completely remove all UniGo data and configurations",
	Long: `Completely remove all UniGo data and configurations.

This command will internal-combust and erase:
  • All download caches and temporary files
  • (Optional) Your configuration directory (~/.config/unigo)

WARNING: This action is permanent and IRREVERSIBLE.`,
	Args: cobra.NoArgs,
	RunE: runImplode,
}

func runImplode(cmd *cobra.Command, args []string) error {
	// 1. Visual Banner
	pterm.DefaultCenter.Println(pterm.Red("!!! DANGER ZONE !!!"))
	pterm.DefaultBigText.WithLetters(
		pterm.NewLettersFromStringWithStyle("IM", pterm.NewStyle(pterm.FgRed)),
		pterm.NewLettersFromStringWithStyle("PLODE", pterm.NewStyle(pterm.FgWhite)),
	).Render()

	configDir := env.GetConfigDir()
	cacheDir := env.GetCacheDir()

	targets := []struct {
		name string
		path string
	}{
		{"Cache Directory", cacheDir},
	}

	if implodeConfig {
		targets = append(targets, struct {
			name string
			path string
		}{"Configuration Directory", configDir})
	}

	// 2. Confirmation
	if !implodeYes {
		pterm.Warning.Prefix = pterm.Prefix{Text: "WARNING", Style: pterm.NewStyle(pterm.BgRed, pterm.FgWhite)}
		pterm.Warning.Println("This will permanently destroy ALL UniGo data.")
		fmt.Printf("\nSelected Targets:\n")
		for _, t := range targets {
			pterm.BulletListPrinter{}.WithItems([]pterm.BulletListItem{
				{Level: 0, Text: fmt.Sprintf("%s (%s)", pterm.Bold.Sprint(t.name), t.path), Bullet: "•", BulletStyle: pterm.NewStyle(pterm.FgRed)},
			}).Render()
		}

		fmt.Print("\nType 'yes' to proceed with self-destruction: ")

		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "yes" {
			fmt.Println("Implode aborted. You live to fight another day.")
			return nil
		}

		fmt.Println()
		for i := 3; i > 0; i-- {
			fmt.Printf("Initiating sequence in %d...\n", i)
			time.Sleep(1 * time.Second)
		}
		fmt.Println()
	}

	fmt.Println("Self-destruct sequence active...\n")

	for _, t := range targets {
		fmt.Printf("Destroying %s... ", t.name)

		if _, err := os.Stat(t.path); os.IsNotExist(err) {
			fmt.Println("Skipped (Already gone)")
			continue
		}

		if err := os.RemoveAll(t.path); err != nil {
			fmt.Printf("FAILED (%v)\n", err)
		} else {
			fmt.Println("ERASED")
		}
	}

	fmt.Println("\nImplode complete.")
	return nil
}
