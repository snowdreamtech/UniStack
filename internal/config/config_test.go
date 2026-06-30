// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTOML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "unigo_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("UNIGO_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("UNIGO_CONFIG_DIR")

	tomlContent := []byte(`debug = true`)
	err = os.WriteFile(filepath.Join(tmpDir, "unigo.toml"), tomlContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write toml: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if !cfg.Debug {
		t.Errorf("Expected Debug to be true, got false")
	}
}

func TestLoadYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "unigo_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("UNIGO_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("UNIGO_CONFIG_DIR")

	yamlContent := []byte(`debug: true`)
	err = os.WriteFile(filepath.Join(tmpDir, "unigo.yaml"), yamlContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write yaml: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if !cfg.Debug {
		t.Errorf("Expected Debug to be true, got false")
	}
}

func TestSave(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "unigo_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("UNIGO_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("UNIGO_CONFIG_DIR")

	cfg := &Config{Debug: true}
	err = cfg.Save()
	if err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	// Verify it wrote TOML
	content, err := os.ReadFile(filepath.Join(tmpDir, "unigo.toml"))
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if string(content) != "debug = true\n" {
		t.Errorf("Unexpected toml content: %q", string(content))
	}
}
