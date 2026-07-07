// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadTOML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "unistack_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("UNISTACK_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("UNISTACK_CONFIG_DIR")

	tomlContent := []byte(`debug = true`)
	err = os.WriteFile(filepath.Join(tmpDir, "unistack.toml"), tomlContent, 0644)
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
	tmpDir, err := os.MkdirTemp("", "unistack_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("UNISTACK_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("UNISTACK_CONFIG_DIR")

	yamlContent := []byte(`debug: true`)
	err = os.WriteFile(filepath.Join(tmpDir, "unistack.yaml"), yamlContent, 0644)
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
	tmpDir, err := os.MkdirTemp("", "unistack_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("UNISTACK_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("UNISTACK_CONFIG_DIR")

	cfg := &Config{Debug: true}
	err = cfg.Save()
	if err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	// Verify it wrote TOML
	content, err := os.ReadFile(filepath.Join(tmpDir, "unistack.toml"))
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if string(content) != "debug = true\n" {
		t.Errorf("Unexpected toml content: %q", string(content))
	}
}

func TestLoadInvalidTOML(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "unistack_config_test")
	defer os.RemoveAll(tmpDir)
	os.Setenv("UNISTACK_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("UNISTACK_CONFIG_DIR")

	os.WriteFile(filepath.Join(tmpDir, "unistack.toml"), []byte(`[invalid`), 0644)

	_, err := Load()
	if err == nil {
		t.Errorf("Expected error when parsing invalid TOML")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "unistack_config_test")
	defer os.RemoveAll(tmpDir)
	os.Setenv("UNISTACK_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("UNISTACK_CONFIG_DIR")

	os.WriteFile(filepath.Join(tmpDir, "unistack.yaml"), []byte(`invalid: yaml: :`), 0644)

	_, err := Load()
	if err == nil {
		t.Errorf("Expected error when parsing invalid YAML")
	}
}

func TestLoadUnreadableConfig(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not support Unix-style file permissions for testing unreadable files")
	}

	tmpDir, _ := os.MkdirTemp("", "unistack_config_test")
	defer os.RemoveAll(tmpDir)
	os.Setenv("UNISTACK_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("UNISTACK_CONFIG_DIR")

	file := filepath.Join(tmpDir, "unistack.toml")
	os.WriteFile(file, []byte(`debug = true`), 0200) // write-only

	_, err := Load()
	if err == nil {
		t.Errorf("Expected error when reading unreadable file")
	}
}

func TestSaveMkdirError(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "unistack_config_test")
	defer os.RemoveAll(tmpDir)

	// Create a file where the config dir should be
	badDir := filepath.Join(tmpDir, "bad_dir")
	os.WriteFile(badDir, []byte("file"), 0644)

	os.Setenv("UNISTACK_CONFIG_DIR", badDir)
	defer os.Unsetenv("UNISTACK_CONFIG_DIR")

	cfg := &Config{Debug: true}
	err := cfg.Save()
	if err == nil {
		t.Errorf("Expected error when MkdirAll fails")
	}
}

func TestSaveWriteError(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "unistack_config_test")
	defer os.RemoveAll(tmpDir)
	os.Setenv("UNISTACK_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("UNISTACK_CONFIG_DIR")

	// Pre-create the config file as read-only
	configPath := filepath.Join(tmpDir, "unistack.toml")
	os.WriteFile(configPath, []byte(""), 0400)

	cfg := &Config{Debug: true}
	err := cfg.Save()
	if err == nil {
		t.Errorf("Expected error when WriteFile fails")
	}
}

func TestLoadDefault(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "unistack_config_test")
	defer os.RemoveAll(tmpDir)
	os.Setenv("UNISTACK_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("UNISTACK_CONFIG_DIR")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() expected no error when config is missing, got %v", err)
	}
	if cfg.Debug {
		t.Errorf("Expected default Debug to be false")
	}
}
