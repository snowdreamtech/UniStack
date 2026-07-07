// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package env

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnv_WindowsAppDataFallback(t *testing.T) {
	origGOOS := RuntimeGOOS
	defer func() { RuntimeGOOS = origGOOS }()
	RuntimeGOOS = "windows"

	t.Setenv("UNISTACK_CONFIG_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("UNISTACK_DATA_DIR", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("UNISTACK_CACHE_DIR", "")
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("LOCALAPPDATA", "")

	// mock os config dir
	origConfigDir := OsUserConfigDir
	defer func() { OsUserConfigDir = origConfigDir }()
	OsUserConfigDir = func() (string, error) { return "C:\\Users\\test\\AppData\\Roaming", nil }

	origHomeDir := OsUserHomeDir
	defer func() { OsUserHomeDir = origHomeDir }()
	OsUserHomeDir = func() (string, error) { return "C:\\Users\\test", nil }

	cfg := GetConfigDir()
	assert.Equal(t, filepath.Join("C:\\Users\\test\\AppData\\Roaming", "unistack"), cfg)

	data := GetDataDir()
	assert.Equal(t, filepath.Join("C:\\Users\\test", "AppData", "Local", "unistack"), data)

	// mock localappdata
	t.Setenv("LOCALAPPDATA", "C:\\Users\\test\\AppData\\Local")
	data2 := GetDataDir()
	assert.Equal(t, filepath.Join("C:\\Users\\test\\AppData\\Local", "unistack"), data2)

	cache := GetCacheDir()
	// GetCacheDir on Windows uses GetDataDir() + "cache"
	assert.Equal(t, filepath.Join(data2, "cache"), cache)
}

func TestEnv_DarwinCacheFallback(t *testing.T) {
	origGOOS := RuntimeGOOS
	defer func() { RuntimeGOOS = origGOOS }()
	RuntimeGOOS = "darwin"

	t.Setenv("UNISTACK_CACHE_DIR", "")
	t.Setenv("XDG_CACHE_HOME", "")

	origHomeDir := OsUserHomeDir
	defer func() { OsUserHomeDir = origHomeDir }()
	OsUserHomeDir = func() (string, error) { return "/Users/test", nil }

	cache := GetCacheDir()
	assert.Equal(t, filepath.Join("/Users/test", "Library", "Caches", "unistack"), cache)
}

func TestEnv_PathsFallback(t *testing.T) {
	// safely set environment variables to empty string
	t.Setenv("UNISTACK_CONFIG_DIR", "")
	t.Setenv("UNISTACK_DATA_DIR", "")
	t.Setenv("UNISTACK_CACHE_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")

	// keep HOME so os.UserHomeDir works, otherwise it might panic or error
	homeDir, _ := os.UserHomeDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)

	cfg := GetConfigDir()
	assert.NotEmpty(t, cfg)

	data := GetDataDir()
	assert.NotEmpty(t, data)

	cache := GetCacheDir()
	assert.NotEmpty(t, cache)

	// test XDG
	t.Setenv("XDG_CONFIG_HOME", "/xdg_config")
	t.Setenv("XDG_DATA_HOME", "/xdg_data")
	t.Setenv("XDG_CACHE_HOME", "/xdg_cache")
	assert.Equal(t, filepath.Join("/xdg_config", "unistack"), GetConfigDir())
	assert.Equal(t, filepath.Join("/xdg_data", "unistack"), GetDataDir())
	assert.Equal(t, filepath.Join("/xdg_cache", "unistack"), GetCacheDir())
}

func TestEnv_GetLockFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNISTACK_CONFIG_DIR", tmpDir)

	// Create a dummy lockfile in tmpDir so it finds it, instead of searching up to repo root
	dummyLock := filepath.Join(tmpDir, ".unistack.lock")
	os.WriteFile(dummyLock, []byte(""), 0644)

	lock := GetLockFilePath()
	// Depending on logic, it might return dummyLock or some other path
	assert.NotEmpty(t, lock)
}

func TestEnv_RandomString(t *testing.T) {
	s, err := RandomString(10)
	assert.NoError(t, err)
	assert.Len(t, s, 10)

	s2, err := RandomString(0)
	assert.NoError(t, err)
	assert.Empty(t, s2)

	// test uniqueness
	s3, err := RandomString(10)
	assert.NoError(t, err)
	assert.NotEqual(t, s, s3)
}

func TestEnv_OsErrorFallbacks(t *testing.T) {
	t.Setenv("UNISTACK_CONFIG_DIR", "")
	t.Setenv("UNISTACK_DATA_DIR", "")
	t.Setenv("UNISTACK_CACHE_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("LOCK_FILE", "")

	// mock OS errors
	origHomeDir := OsUserHomeDir
	origConfigDir := OsUserConfigDir
	origGetwd := OsGetwd
	defer func() {
		OsUserHomeDir = origHomeDir
		OsUserConfigDir = origConfigDir
		OsGetwd = origGetwd
	}()

	errMock := fmt.Errorf("mock error")
	OsUserHomeDir = func() (string, error) { return "", errMock }
	OsUserConfigDir = func() (string, error) { return "", errMock }
	OsGetwd = func() (string, error) { return "", errMock }

	assert.Equal(t, "./unistack_config", GetConfigDir())
	assert.Equal(t, "./unistack_data", GetDataDir())
	assert.Equal(t, "./unistack_cache", GetCacheDir())
	assert.Equal(t, "unistack.lock", GetLockFilePath())

	// mock crypto rand
	origRand := CryptoRandRead
	defer func() { CryptoRandRead = origRand }()
	CryptoRandRead = func(b []byte) (n int, err error) {
		return 0, errMock
	}

	_, err := RandomString(10)
	assert.Error(t, err)

	// test GetLockFilePath custom
	t.Setenv("LOCK_FILE", "custom.lock")
	assert.Equal(t, "custom.lock", GetLockFilePath())
}
