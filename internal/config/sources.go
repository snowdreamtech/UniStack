// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/snowdreamtech/unistack/internal/env"
)

// Source represents a remote package registry source.
type Source struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// GetSourcesFile returns the absolute path to the sources.json configuration file.
func GetSourcesFile() string {
	return filepath.Join(env.GetConfigDir(), "sources.json")
}

// GetDefaultSources returns the default out-of-the-box sources.
func GetDefaultSources() []Source {
	return []Source{
		{
			Name: "core",
			URL:  "https://registry.unistack.org/core",
		},
		{
			Name: "community",
			URL:  "https://registry.unistack.org/community",
		},
	}
}

// LoadSources reads the configured sources from disk.
// If the file doesn't exist, it returns the default sources.
func LoadSources() ([]Source, error) {
	file := GetSourcesFile()
	data, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return GetDefaultSources(), nil
		}
		return nil, fmt.Errorf("failed to read sources config: %w", err)
	}

	var sources []Source
	if err := json.Unmarshal(data, &sources); err != nil {
		// If JSON is invalid, return default to avoid crashing the whole CLI.
		return GetDefaultSources(), nil
	}

	if len(sources) == 0 {
		return GetDefaultSources(), nil
	}

	return sources, nil
}

// saveSources writes the given sources back to the JSON configuration file.
func saveSources(sources []Source) error {
	file := GetSourcesFile()
	
	// Ensure config directory exists
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(sources, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode sources: %w", err)
	}

	if err := os.WriteFile(file, data, 0644); err != nil {
		return fmt.Errorf("failed to write sources config: %w", err)
	}
	
	return nil
}

func normalizeURL(u string) string {
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "file://") {
		return u
	}
	abs, err := filepath.Abs(u)
	if err == nil {
		abs = filepath.ToSlash(abs)
		if !strings.HasPrefix(abs, "/") {
			abs = "/" + abs
		}
		return "file://" + abs
	}
	return u
}

var sourceNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

func validateSourceName(name string) error {
	if name == "." || name == ".." {
		return fmt.Errorf("invalid source name '%s': cannot be '.' or '..'", name)
	}
	if !sourceNamePattern.MatchString(name) {
		return fmt.Errorf("invalid source name '%s': only alphanumeric, dashes, underscores, and dots are allowed", name)
	}
	return nil
}

// AddSource adds a new registry source.
func AddSource(name, url string) error {
	if err := validateSourceName(name); err != nil {
		return err
	}
	url = normalizeURL(url)
	sources, err := LoadSources()
	if err != nil {
		return err
	}

	for _, s := range sources {
		if s.Name == name {
			return fmt.Errorf("source '%s' already exists", name)
		}
	}

	sources = append(sources, Source{Name: name, URL: url})
	return saveSources(sources)
}

// UpdateSource updates the URL of an existing registry source.
func UpdateSource(name, newURL string) error {
	newURL = normalizeURL(newURL)
	sources, err := LoadSources()
	if err != nil {
		return err
	}

	found := false
	for i, s := range sources {
		if s.Name == name {
			sources[i].URL = newURL
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("source '%s' not found", name)
	}

	return saveSources(sources)
}

// RemoveSource deletes a registry source from the configuration.
func RemoveSource(name string) error {
	sources, err := LoadSources()
	if err != nil {
		return err
	}

	var updated []Source
	found := false
	for _, s := range sources {
		if s.Name == name {
			found = true
			continue
		}
		updated = append(updated, s)
	}

	if !found {
		return fmt.Errorf("source '%s' not found", name)
	}

	// Remove associated cache db file
	dbPath := filepath.Join(env.GetRegistryCacheDir(), name+".db")
	_ = os.Remove(dbPath)

	return saveSources(updated)
}
