// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package updater

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateCache_ReadWrite(t *testing.T) {
	// Setup temporary data directory
	tempDir := t.TempDir()
	t.Setenv("UNIGO_DATA_DIR", tempDir)

	// Initially cache should be empty
	cache, err := readCache()
	require.NoError(t, err)
	assert.Equal(t, "", cache.LatestVersion)
	assert.True(t, cache.LastChecked.IsZero())
	assert.True(t, cache.LastPrompted.IsZero())

	// Write cache
	now := time.Now()
	expected := &UpdateCache{
		LatestVersion: "1.0.0",
		LastChecked:   now,
		LastPrompted:  now.Add(-24 * time.Hour),
	}
	err = writeCache(expected)
	require.NoError(t, err)

	// Read cache back
	actual, err := readCache()
	require.NoError(t, err)
	assert.Equal(t, expected.LatestVersion, actual.LatestVersion)

	// time.Time comparison over JSON serialization may lose some precision
	assert.Equal(t, expected.LastChecked.Unix(), actual.LastChecked.Unix())
	assert.Equal(t, expected.LastPrompted.Unix(), actual.LastPrompted.Unix())
}

func TestUpdateCache_Corrupt(t *testing.T) {
	// Setup temporary data directory
	tempDir := t.TempDir()
	t.Setenv("UNIGO_DATA_DIR", tempDir)

	// Write corrupt JSON
	cachePath := getCachePath()
	err := os.MkdirAll(filepath.Dir(cachePath), 0755)
	require.NoError(t, err)
	err = os.WriteFile(cachePath, []byte("invalid json"), 0644)
	require.NoError(t, err)

	// Reading corrupt cache should return empty struct instead of error
	cache, err := readCache()
	require.NoError(t, err)
	assert.Equal(t, "", cache.LatestVersion)
}

func TestFetchLatestRelease(t *testing.T) {
	// Only run this test if explicitly requested, as it makes network calls
	if os.Getenv("TEST_NETWORK") == "" {
		t.Skip("Skipping network test. Set TEST_NETWORK=1 to enable.")
	}

	ver, err := fetchLatestRelease()
	require.NoError(t, err)
	assert.NotEmpty(t, ver)
}

func TestPromptIfAvailable_Blacklist(t *testing.T) {
	// Tests that blacklisted commands do not output the prompt.
	// Ensure it doesn't panic and returns early for all blacklisted commands.
	PromptIfAvailable("0.5.0", "env")
	PromptIfAvailable("0.5.0", "version")
	PromptIfAvailable("0.5.0", "completion")
	PromptIfAvailable("0.5.0", "self-update")
	PromptIfAvailable("0.5.0", "__complete")
}

func TestCheckUpdateAsync_SkipDevVersions(t *testing.T) {
	// Dev/N/A versions should be silently skipped
	CheckUpdateAsync("N/A")
	CheckUpdateAsync("dev")
	CheckUpdateAsync("")
}

func TestCheckUpdateAsync_RecentCheck(t *testing.T) {
	// If last checked < 24h ago, should not re-fetch
	tempDir := t.TempDir()
	t.Setenv("UNIGO_DATA_DIR", tempDir)

	recent := &UpdateCache{
		LatestVersion: "1.0.0",
		LastChecked:   time.Now(), // just now → within checkPeriod
	}
	err := writeCache(recent)
	require.NoError(t, err)

	// Should return early without making a network call
	CheckUpdateAsync("0.9.0")
}

func TestPromptIfAvailable_DevVersion(t *testing.T) {
	// Should skip silently for dev/N/A
	PromptIfAvailable("N/A", "install")
	PromptIfAvailable("dev", "install")
	PromptIfAvailable("", "install")
}

func TestGetCachePath(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("UNIGO_DATA_DIR", tempDir)

	path := getCachePath()
	assert.Contains(t, path, "update-cache.json")
	assert.Contains(t, path, tempDir)
}

func TestWriteCache_CreatesDir(t *testing.T) {
	// Write to a nested path that doesn't exist yet
	tempDir := t.TempDir()
	t.Setenv("UNIGO_DATA_DIR", filepath.Join(tempDir, "nested", "path"))

	cache := &UpdateCache{LatestVersion: "2.0.0", LastChecked: time.Now()}
	err := writeCache(cache)
	require.NoError(t, err)

	// Verify file was created
	assert.FileExists(t, getCachePath())
}
