// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/snowdreamtech/unistack/internal/client"
	"github.com/snowdreamtech/unistack/internal/config"
	"github.com/spf13/cobra"
)

// sourceCmd represents the source management command
var sourceCmd = &cobra.Command{
	Use:   "source",
	Short: "Manage package registry sources",
	Long:  `Add, remove, list, or update package registry sources (mirrors/private registries).`,
}

var sourceAddCmd = &cobra.Command{
	Use:   "add [name] [url]",
	Short: "Add a new package registry source",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		url := args[1]
		err := config.AddSource(name, url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully added source '%s' (%s)\n", name, url)

		fmt.Printf("Syncing source '%s' database...\n", name)
		if err := client.UpdateSource(cmd.Context(), name, url); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to sync database for '%s': %v\n", name, err)
		} else {
			fmt.Printf("Successfully synced source '%s'\n", name)
		}
	},
}

var sourceUpdateCmd = &cobra.Command{
	Use:   "update [name] [new_url]",
	Short: "Update the URL of an existing registry source",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		url := args[1]
		err := config.UpdateSource(name, url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully updated source '%s' to %s\n", name, url)

		fmt.Printf("Syncing source '%s' database...\n", name)
		if err := client.UpdateSource(cmd.Context(), name, url); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to sync database for '%s': %v\n", name, err)
		} else {
			fmt.Printf("Successfully synced source '%s'\n", name)
		}
	},
}

var sourceRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a package registry source",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		err := config.RemoveSource(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully removed source '%s'\n", name)
	},
}

var sourceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured package registry sources",
	Run: func(cmd *cobra.Command, args []string) {
		sources, err := config.LoadSources()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(sources) == 0 {
			fmt.Println("No sources configured.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
		fmt.Fprintln(w, "NAME\tURL")
		for _, s := range sources {
			fmt.Fprintf(w, "%s\t%s\n", s.Name, s.URL)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(sourceCmd)
	sourceCmd.AddCommand(sourceAddCmd)
	sourceCmd.AddCommand(sourceUpdateCmd)
	sourceCmd.AddCommand(sourceRemoveCmd)
	sourceCmd.AddCommand(sourceListCmd)
}
