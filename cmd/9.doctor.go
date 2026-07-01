// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unigo/internal/cli/output"
	"github.com/snowdreamtech/unigo/internal/database"
	"github.com/snowdreamtech/unigo/internal/pkg/env"
	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(doctorCmd)
	}
}

var doctorCmd = &cobra.Command{
	Use:     "doctor",
	Aliases: []string{"dr"},
	Short:   "Check system health and diagnose issues",
	Long: `Check UniGo system health and diagnose potential issues.

This command partially aligns with UniRTM to ensure your environment is
properly configured, providing insights into directories, configurations, and cache.`,
	Args: cobra.NoArgs,
	RunE: runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// 1. Core Status
	pterm.DefaultSection.Println("🚀 Core Status")
	statusItems := []pterm.BulletListItem{
		{Level: 0, Text: fmt.Sprintf("%-12s: %s (%s)", "Version", pterm.LightCyan(env.GitTag), env.CommitHash)},
		{Level: 0, Text: fmt.Sprintf("%-12s: %s/%s", "Target", runtime.GOOS, runtime.GOARCH)},
		{Level: 0, Text: fmt.Sprintf("%-12s: %s", "Go Version", runtime.Version())},
		{Level: 0, Text: fmt.Sprintf("%-12s: %s", "Built", env.BuildTime)},
	}
	pterm.DefaultBulletList.WithItems(statusItems).Render()

	// 2. Context & Environment
	pterm.DefaultSection.Println("🐚 Context & Environment")
	cwd, _ := os.Getwd()

	pterm.DefaultBulletList.WithItems([]pterm.BulletListItem{
		{Level: 0, Text: fmt.Sprintf("%-12s: %s", "Work Dir", pterm.FgGray.Sprint(cwd))},
	}).Render()

	// 3. Directories & Usage
	pterm.DefaultSection.Println("📁 Directories & Usage")
	dirData := [][]string{
		{"cache", env.GetCacheDir(), getDirSize(env.GetCacheDir())},
		{"config", env.GetConfigDir(), "-"},
		{"data", env.GetDataDir(), getDirSize(env.GetDataDir())},
	}
	var dirTable [][]string
	dirTable = append(dirTable, []string{"Type", "Path", "Size"})
	for _, d := range dirData {
		dirTable = append(dirTable, []string{pterm.Bold.Sprint(d[0]), pterm.FgGray.Sprint(d[1]), pterm.LightCyan(d[2])})
	}
	pterm.DefaultTable.WithHasHeader().WithData(dirTable).Render()

	// 4. Configuration
	pterm.DefaultSection.Println("📝 Configuration")
	configs := []string{
		filepath.Join(cwd, ".unigo.toml"),
		filepath.Join(cwd, "unigo.toml"),
		env.GetGlobalConfigPath(),
	}
	var foundConfig bool
	for _, c := range configs {
		if _, err := os.Stat(c); err == nil {
			output.Successf("Loaded: %s", pterm.FgGray.Sprint(c))
			foundConfig = true
		}
	}
	if !foundConfig {
		output.Info("No configuration files found.")
	}

	// 5. Health Checks
	pterm.DefaultSection.Println("🌐 Health Checks")
	suggestions := 0
	
	// DB Check
	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true})
	if err != nil {
		output.Errorf("Database: %v", err)
		suggestions++
	} else {
		defer db.Close()
		output.Successf("Database: %s (Size: %s)", pterm.FgGray.Sprint(dbPath), getFileSize(dbPath))
	}

	// Check CWD Permissions
	if wd, err := os.Getwd(); err == nil {
		if f, err := os.CreateTemp(wd, ".unigo-doctor-*"); err == nil {
			f.Close()
			os.Remove(f.Name())
			output.Successf("Current dir: %s (Writable)", wd)
		} else {
			output.Warningf("Current dir: %s (Read-only or restricted: %v)", wd, err)
			suggestions++
		}
	}

	fmt.Println()
	if suggestions == 0 {
		pterm.DefaultBox.WithTitle(pterm.LightGreen("Diagnostics Complete")).Println("Your UniGo environment is perfectly configured and ready.")
	} else {
		pterm.DefaultBox.WithTitle(pterm.LightYellow("Diagnostics Complete")).Printf("Found %d potential issue(s).\n", suggestions)
	}

	return nil
}

func getDirSize(path string) string {
	var size int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return formatSize(size)
}

func getFileSize(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return "0 B"
	}
	return formatSize(info.Size())
}

func formatSize(size int64) string {
	if size == 0 {
		return "0 B"
	}
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
