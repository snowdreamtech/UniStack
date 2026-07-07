// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"github.com/snowdreamtech/unistack/internal/pkg/env"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	Debug bool `json:"debug" yaml:"debug" toml:"debug"`
}

// Load reads the configuration from the global config directory.
// It prioritizes TOML (unistack.toml) over YAML (unistack.yaml/unistack.yml).
func Load() (*Config, error) {
	configDir := env.GetConfigDir()
	cfg := &Config{}

	// Check TOML
	tomlPath := filepath.Join(configDir, "unistack.toml")
	if data, err := os.ReadFile(tomlPath); err == nil {
		if err := toml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse TOML config: %w", err)
		}
		return cfg, nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("error reading TOML config: %w", err)
	}

	// Check YAML
	yamlPaths := []string{"unistack.yaml", "unistack.yml"}
	for _, yp := range yamlPaths {
		yamlPath := filepath.Join(configDir, yp)
		if data, err := os.ReadFile(yamlPath); err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse YAML config: %w", err)
			}
			return cfg, nil
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error reading YAML config: %w", err)
		}
	}

	// Default config
	return cfg, nil
}

// Save writes the current configuration to unistack.toml in the global config directory.
func (c *Config) Save() error {
	configPath := env.GetGlobalConfigPath() // Defaults to unistack.toml

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := toml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
