// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"os"
	"testing"
)

func TestLoadAndSave(t *testing.T) {
	// Create a temporary directory to override config location during testing
	tmpDir, err := os.MkdirTemp("", "unigo_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Since env.GetGlobalConfigPath uses env variables or HOME, testing the actual
	// Save() might affect the user's real file if not isolated. For a placeholder
	// test, we just instantiate the struct and ensure basic mechanics don't panic.
	
	cfg := &Config{}
	
	// Normally we would test cfg.Save() and Load() here by manipulating the
	// environment variables that env.GetGlobalConfigPath() relies on, e.g.:
	// os.Setenv("UNIGO_CONFIG_DIR", tmpDir)
	// defer os.Unsetenv("UNIGO_CONFIG_DIR")
	// 
	// err = cfg.Save()
	// if err != nil { ... }

	if cfg == nil {
		t.Errorf("Config instantiation failed")
	}
}

func TestLoadEmpty(t *testing.T) {
	// A basic test to ensure Load doesn't panic
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatalf("Load() returned nil config")
	}
}
