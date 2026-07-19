package config

import (
	"os"
	"testing"
)

func TestSourcesCRUD(t *testing.T) {
	// Setup isolated config dir for tests
	tempDir, err := os.MkdirTemp("", "unistack_test_config")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Override HOME for the test to ensure env.GetConfigDir() resolves here
	os.Setenv("HOME", tempDir)
	os.Setenv("USERPROFILE", tempDir) // for Windows

	// 1. Test Load Default
	sources, err := LoadSources()
	if err != nil {
		t.Fatalf("LoadSources failed: %v", err)
	}
	if len(sources) != 1 || sources[0].Name != "default" {
		t.Fatalf("Expected default source, got %v", sources)
	}

	// 2. Test Add Source
	if err := AddSource("custom", "http://test.local"); err != nil {
		t.Fatalf("AddSource failed: %v", err)
	}
	
	sources, _ = LoadSources()
	if len(sources) != 2 {
		t.Fatalf("Expected 2 sources, got %d", len(sources))
	}
	if sources[1].Name != "custom" || sources[1].URL != "http://test.local" {
		t.Fatalf("Custom source incorrect: %v", sources[1])
	}

	// 3. Test Update Source
	if err := UpdateSource("custom", "http://test.updated"); err != nil {
		t.Fatalf("UpdateSource failed: %v", err)
	}
	
	sources, _ = LoadSources()
	if sources[1].URL != "http://test.updated" {
		t.Fatalf("UpdateSource didn't change URL, got: %s", sources[1].URL)
	}

	// 4. Test Remove Source
	if err := RemoveSource("custom"); err != nil {
		t.Fatalf("RemoveSource failed: %v", err)
	}
	
	sources, _ = LoadSources()
	if len(sources) != 1 {
		t.Fatalf("Expected 1 source after removal, got %d", len(sources))
	}
}
