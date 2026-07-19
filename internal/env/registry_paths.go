// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package env

import "path/filepath"

// GetRegistryCacheDir returns the directory where registry files are stored.
func GetRegistryCacheDir() string {
	return filepath.Join(GetCacheDir(), "registry")
}

// GetSourceDatabasePath returns the path to the registry database for a specific source.
func GetSourceDatabasePath(name string) string {
	return filepath.Join(GetRegistryCacheDir(), name+".db")
}
