// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package env

import (
	"errors"
	"os"
	"testing"
)

func TestEnvPaths_CoverageMore(t *testing.T) {
	// 1. Test GetConfigDir windows path
	oldOS := RuntimeGOOS
	RuntimeGOOS = "windows"

	os.Setenv("UNISTACK_CONFIG_DIR", "")
	os.Setenv("UNISTACK_XDG_CONFIG_HOME", "")

	oldOsUserConfigDir := OsUserConfigDir
	OsUserConfigDir = func() (string, error) { return "C:\\Users\\test\\AppData\\Roaming", nil }
	GetConfigDir()

	OsUserConfigDir = func() (string, error) { return "", errors.New("err") }
	GetConfigDir()

	// 2. Test GetDataDir windows
	os.Setenv("UNISTACK_DATA_DIR", "")
	os.Setenv("UNISTACK_XDG_DATA_HOME", "")
	os.Setenv("UNISTACK_LOCALAPPDATA", "C:\\AppData\\Local")
	GetDataDir()
	os.Unsetenv("UNISTACK_LOCALAPPDATA")
	GetDataDir()

	// 3. Test GetCacheDir darwin
	os.Setenv("UNISTACK_CACHE_DIR", "")
	os.Setenv("UNISTACK_XDG_CACHE_HOME", "")
	RuntimeGOOS = "darwin"
	GetCacheDir()

	// 4. Test GetCacheDir windows
	RuntimeGOOS = "windows"
	GetCacheDir()

	// 5. Test GetLockFilePath error
	os.Setenv("UNISTACK_LOCK_FILE", "")
	oldOsGetwd := OsGetwd
	OsGetwd = func() (string, error) { return "", errors.New("err") }
	GetLockFilePath()

	// Reset
	RuntimeGOOS = oldOS
	OsUserConfigDir = oldOsUserConfigDir
	OsGetwd = oldOsGetwd

	// 6. Test OsUserHomeDir error paths
	oldOsUserHomeDir := OsUserHomeDir
	OsUserHomeDir = func() (string, error) { return "", errors.New("err") }
	os.Setenv("UNISTACK_CONFIG_DIR", "")
	os.Setenv("UNISTACK_XDG_CONFIG_HOME", "")
	GetConfigDir()

	os.Setenv("UNISTACK_DATA_DIR", "")
	os.Setenv("UNISTACK_XDG_DATA_HOME", "")
	GetDataDir()

	os.Setenv("UNISTACK_CACHE_DIR", "")
	os.Setenv("UNISTACK_XDG_CACHE_HOME", "")
	GetCacheDir()

	OsUserHomeDir = oldOsUserHomeDir

	// 7. Test direct UNISTACK_DIR env vars
	os.Setenv("UNISTACK_CONFIG_DIR", "/custom/config")
	GetConfigDir()
	os.Unsetenv("UNISTACK_CONFIG_DIR")

	os.Setenv("UNISTACK_DATA_DIR", "/custom/data")
	GetDataDir()
	os.Unsetenv("UNISTACK_DATA_DIR")

	os.Setenv("UNISTACK_CACHE_DIR", "/custom/cache")
	GetCacheDir()
	os.Unsetenv("UNISTACK_CACHE_DIR")

	// 8. Test GetCacheDir linux fallback
	RuntimeGOOS = "linux"
	GetCacheDir()
	RuntimeGOOS = oldOS
}
