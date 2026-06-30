// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/snowdreamtech/unigo/internal/pkg/env"
	"github.com/spf13/cobra"
)

var (
	implodeYes    bool
	implodeConfig bool
)

var implodeCmd = &cobra.Command{
	Use:   "implode",
	Short: "Completely remove all UniGo data and installations",
	Long: `Completely remove all UniGo data and installations.

This command will internal-combust and erase:
  • All data directories
  • All caches and temporary files
  • (Optional) Your configuration directory (~/.config/unigo)

WARNING: This action is permanent and IRREVERSIBLE.`,
	Args: cobra.NoArgs,
	RunE: runImplode,
}

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(implodeCmd)
	}
	implodeCmd.Flags().BoolVarP(&implodeYes, "yes", "y", false, "skip confirmation prompt")
	implodeCmd.Flags().BoolVar(&implodeConfig, "config", false, "also remove configuration directory (~/.config/unigo)")
}

func runImplode(cmd *cobra.Command, args []string) error {
	fmt.Println("!!! DANGER ZONE !!!")
	fmt.Println("IMPLODE SEQUENCE INITIATED")
	fmt.Println()

	dataDir := env.GetDataDir()
	configDir := env.GetConfigDir()

	targets := []struct {
		name string
		path string
	}{
		{"Data Directory", dataDir},
		{"Cache Directory", env.GetCacheDir()},
	}

	if implodeConfig {
		targets = append(targets, struct {
			name string
			path string
		}{"Configuration Directory", configDir})
	}

	if !implodeYes {
		fmt.Println("WARNING: This will permanently destroy ALL UniGo data.")
		fmt.Println("\nSelected Targets:")
		for _, t := range targets {
			fmt.Printf("  • %s (%s)\n", t.name, t.path)
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
