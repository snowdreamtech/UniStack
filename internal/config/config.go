// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/snowdreamtech/unigo/internal/pkg/env"
)

// Config represents the application configuration.
// This is a placeholder struct that can be expanded with real configuration fields.
type Config struct {
	// Add configuration fields here, for example:
	// Debug bool `json:"debug" yaml:"debug" toml:"debug"`
}

// Load reads the configuration from the global config file.
// Currently acts as a placeholder that returns an empty config.
func Load() (*Config, error) {
	configPath := env.GetGlobalConfigPath()

	// If the file doesn't exist, we can just return a default configuration.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil
	}

	// TODO: Implement actual configuration parsing (e.g., using Viper, go-yaml, go-toml)
	// For now, return a placeholder
	return &Config{}, nil
}

// Save writes the current configuration to the global config file.
// Currently acts as a placeholder that just ensures the file exists.
func (c *Config) Save() error {
	configPath := env.GetGlobalConfigPath()

	// Ensure the parent directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// TODO: Implement actual configuration serialization.
	// For now, write a dummy file if it doesn't exist.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return os.WriteFile(configPath, []byte("# UniGo Configuration\n\n"), 0644)
	}

	return nil
}
