// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unigo/internal/env"
	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(configCmd)
		configCmd.AddCommand(configGetCmd)
		configCmd.AddCommand(configSetCmd)
	}
}

func getConfigPath() string {
	return filepath.Join(env.GetConfigDir(), "config.json")
}

func loadConfig() map[string]string {
	path := getConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return make(map[string]string)
	}
	var conf map[string]string
	if err := json.Unmarshal(data, &conf); err != nil {
		return make(map[string]string)
	}
	return conf
}

func saveConfig(conf map[string]string) error {
	path := getConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  "Query or modify the configuration settings.",
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		slog.Debug("Getting config value", "key", key)

		conf := loadConfig()
		if val, ok := conf[key]; ok {
			fmt.Println(val)
		} else {
			pterm.Warning.Printf("Key '%s' not found in configuration\n", key)
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		val := args[1]
		slog.Debug("Setting config value", "key", key, "value", val)

		conf := loadConfig()
		conf[key] = val
		if err := saveConfig(conf); err != nil {
			pterm.Error.Printf("Failed to save configuration: %v\n", err)
			return err
		}

		pterm.Success.Printf("Set '%s' to '%s'\n", key, val)
		return nil
	},
}
