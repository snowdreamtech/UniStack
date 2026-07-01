// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package shell

import (
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unigo/internal/pkg/env"
)

// ShellType represents the type of shell being used.
type ShellType string

const (
	ShellBash       ShellType = "bash"
	ShellZsh        ShellType = "zsh"
	ShellFish       ShellType = "fish"
	ShellPowerShell ShellType = "powershell"
)

// DetectShell attempts to determine the user's current shell based on environment variables.
func DetectShell() (ShellType, error) {
	// On Unix-like systems, check SHELL environment variable first
	if shellPath := env.Get("SHELL"); shellPath != "" {
		shell := filepath.Base(shellPath)
		switch {
		case strings.Contains(shell, "bash"):
			return ShellBash, nil
		case strings.Contains(shell, "zsh"):
			return ShellZsh, nil
		case strings.Contains(shell, "fish"):
			return ShellFish, nil
		}
	}

	// On Windows, default to PowerShell if SHELL is not set or not recognized
	if env.RuntimeGOOS == "windows" {
		return ShellPowerShell, nil
	}

	// Default to bash on Unix-like systems
	return ShellBash, nil
}
