// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/snowdreamtech/unigo/internal/pkg/env"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open the config file in $EDITOR",
	Long: `Open the UniGo config file in your preferred editor.

Priority for finding an editor:
1.  UNIGO_EDITOR environment variable
2.  VISUAL environment variable
3.  EDITOR environment variable
4.  Standard system defaults (vim, nano, notepad)`,
	Args: cobra.NoArgs,
	RunE: runEdit,
}

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(editCmd)
	}
}

func runEdit(cmd *cobra.Command, args []string) error {
	editor, source := getBestEditorWithSource()
	fmt.Printf("Opening configuration editor (using %s via %s)...\n", editor, source)

	if source == "system default" {
		fmt.Printf("Tip: Set $EDITOR to change your preference.\n\n")
	}

	targetFile := filepath.Join(env.GetConfigDir(), "unigo.yaml")

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(targetFile), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create a dummy config if it doesn't exist
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		fmt.Printf("Creating new config file: %s\n", targetFile)
		if err := os.WriteFile(targetFile, []byte("# UniGo Configuration\n\n"), 0644); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
	}

	c := exec.Command(editor, targetFile)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	fmt.Printf("Configuration saved: %s\n", targetFile)
	return nil
}

func getBestEditorWithSource() (string, string) {
	if e := os.Getenv("UNIGO_EDITOR"); e != "" {
		return e, "UNIGO_EDITOR"
	}
	if e := os.Getenv("VISUAL"); e != "" {
		return e, "VISUAL"
	}
	if e := os.Getenv("EDITOR"); e != "" {
		return e, "EDITOR"
	}

	// Fallbacks
	if _, err := exec.LookPath("vim"); err == nil {
		return "vim", "system default"
	}
	if _, err := exec.LookPath("nano"); err == nil {
		return "nano", "system default"
	}
	if _, err := exec.LookPath("notepad"); err == nil {
		return "notepad", "system default"
	}
	return "vi", "system default"
}
