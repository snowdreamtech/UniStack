// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package env

import (
	"path/filepath"
)

// GetInstallPackagesDir returns the directory where packages are extracted.
// It maps to ~/.local/share/unistack/packages on Unix-like systems.
func GetInstallPackagesDir() string {
	return filepath.Join(GetDataDir(), "packages")
}

// GetInstallBinDir returns the directory where executable symlinks are placed.
// It maps to ~/.local/bin on Unix-like systems.
func GetInstallBinDir() string {
	homeDir, err := OsUserHomeDir()
	if err != nil {
		return "./.local/bin"
	}

	if RuntimeGOOS == "windows" {
		if localAppData := Get("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, "unistack", "bin")
		}
		return filepath.Join(homeDir, "AppData", "Local", "unistack", "bin")
	}

	return filepath.Join(homeDir, ".local", "bin")
}
