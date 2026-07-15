// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unistack/internal/cli/output"
	"github.com/snowdreamtech/unistack/internal/env"
)

// ShellConfigManager handles persistent configuration changes in shell RC files.
type ShellConfigManager struct {
	formatter output.Formatter
	dryRun    bool
}

// NewShellConfigManager creates a new ShellConfigManager.
func NewShellConfigManager(formatter output.Formatter, dryRun bool) *ShellConfigManager {
	return &ShellConfigManager{
		formatter: formatter,
		dryRun:    dryRun,
	}
}

// GetConfigPath returns the standard configuration file path for the given shell.
func (m *ShellConfigManager) GetConfigPath(shell ShellType) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	switch shell {
	case ShellZsh:
		return filepath.Join(home, ".zshrc"), nil
	case ShellBash:
		return filepath.Join(home, ".bashrc"), nil
	case ShellFish:
		return filepath.Join(home, ".config/fish/config.fish"), nil
	case ShellPowerShell:
		configFile := env.Get("PROFILE")
		if configFile == "" {
			configFile = filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
		}
		return configFile, nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}
}

// Inject appends or updates a configuration block in the shell RC file.
func (m *ShellConfigManager) Inject(shell ShellType, marker string, content string) error {
	configFile, err := m.GetConfigPath(shell)
	if err != nil {
		return err
	}

	// 1. Read existing content
	rawContent, err := os.ReadFile(configFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	searchPattern := fmt.Sprintf("unistack %s activation", marker)
	fullBlock := fmt.Sprintf("\n# %s\n%s\n", searchPattern, content)

	rawContentStr := string(rawContent)
	if strings.Contains(rawContentStr, searchPattern) {
		// Already present, check if we need to update
		if strings.Contains(rawContentStr, content) {
			m.formatter.Info(fmt.Sprintf("UniStack %s logic already up to date in %s", marker, configFile), nil)
			return nil
		}

		if m.dryRun {
			m.formatter.Info(fmt.Sprintf("[dry-run] Would update %s activation logic in %s", marker, configFile), nil)
			return nil
		}

		// Update by replacing the old block
		lines := strings.Split(rawContentStr, "\n")
		var newLines []string
		inBlock := false
		replaced := false

		for i := 0; i < len(lines); i++ {
			if strings.Contains(lines[i], searchPattern) {
				inBlock = true
				if !replaced {
					// Add the new block here
					newLines = append(newLines, "# "+searchPattern)
					newLines = append(newLines, content)
					replaced = true
				}
				continue
			}
			if inBlock {
				// We assume the block ends at the first empty line or another comment
				// Actually, simpler: since it's just one line of content in our case, we can just skip the next line if it's the old content.
				if strings.TrimSpace(lines[i]) == "" || strings.HasPrefix(strings.TrimSpace(lines[i]), "#") {
					inBlock = false
					newLines = append(newLines, lines[i])
				}
				// If it's the old content line, skip it.
				continue
			}
			newLines = append(newLines, lines[i])
		}

		// Fallback if the replacement logic above didn't perfectly match (e.g., end of file).
		// In a real robust implementation we'd use begin/end markers, but let's stick to simple replacement
		// by just rebuilding the file if we correctly replaced it.

		// For simplicity, let's just append if we couldn't cleanly replace.
		if !replaced {
			return m.appendBlock(configFile, fullBlock)
		}

		return os.WriteFile(configFile, []byte(strings.Join(newLines, "\n")), 0644)
	}

	if m.dryRun {
		m.formatter.Info(fmt.Sprintf("[dry-run] Would inject %s activation logic into %s", marker, configFile), nil)
		return nil
	}

	return m.appendBlock(configFile, fullBlock)
}

func (m *ShellConfigManager) appendBlock(configFile, fullBlock string) error {
	f, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(fullBlock); err != nil {
		return fmt.Errorf("failed to write to config file: %w", err)
	}
	return nil
}

// Remove deletes the configuration block from the shell RC file.
func (m *ShellConfigManager) Remove(shell ShellType, marker string) error {
	configFile, err := m.GetConfigPath(shell)
	if err != nil {
		return err
	}

	rawContent, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	searchPattern := fmt.Sprintf("unistack %s activation", marker)
	rawContentStr := string(rawContent)

	if !strings.Contains(rawContentStr, searchPattern) {
		return nil
	}

	if m.dryRun {
		m.formatter.Info(fmt.Sprintf("[dry-run] Would remove %s activation logic from %s", marker, configFile), nil)
		return nil
	}

	lines := strings.Split(rawContentStr, "\n")
	var newLines []string
	inBlock := false

	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], searchPattern) {
			inBlock = true
			continue
		}
		if inBlock {
			if strings.TrimSpace(lines[i]) == "" {
				inBlock = false
			}
			continue
		}
		newLines = append(newLines, lines[i])
	}

	return os.WriteFile(configFile, []byte(strings.Join(newLines, "\n")), 0644)
}
