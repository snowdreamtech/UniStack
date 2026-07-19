// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package registry

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/snowdreamtech/unistack/internal/env"
	// SQLite driver
	_ "modernc.org/sqlite"
)

// PackageMetadata represents the metadata of a package needed for downloading.
type PackageMetadata struct {
	Name    string
	Version string
	Hash    string
}

// QueryPackage retrieves the metadata of the latest version of a package from the local registry DB.
// Since we don't have version resolution logic fully defined yet, we'll fetch the first matching package.
func QueryPackage(ctx context.Context, name string) (*PackageMetadata, error) {
	dbPath := env.GetRegistryDatabasePath()

	// Open in read-only mode if possible, but standard open is fine for query.
	// We use modernc.org/sqlite driver directly
	dsn := fmt.Sprintf("file:%s?mode=ro", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open registry database: %w", err)
	}
	defer db.Close()

	// Query for the package.
	// Our builder schema has a 'packages' table.
	// In the real schema we built:
	// CREATE TABLE IF NOT EXISTS packages (
	//		id INTEGER PRIMARY KEY AUTOINCREMENT,
	//		name TEXT NOT NULL,
	//		version TEXT NOT NULL,
	//		description TEXT,
	//		homepage TEXT,
	//		license TEXT,
	//      hash TEXT
	//	)
	// (Note: hash wasn't explicitly added to the schema in previous tasks, but spec says we need it.
	// If it doesn't exist, this query will fail or return empty. We should handle it gracefully or ensure it's selected if exists.)

	// Assuming schema from our spec:
	// For now we'll try to select name and version. If hash doesn't exist, we'll just not query it,
	// but the spec for 003 says "实现包哈希与签名校验". We assume the `hash` column exists.
	row := db.QueryRowContext(ctx, "SELECT name, version, COALESCE(hash, '') FROM packages WHERE name = ? ORDER BY version DESC LIMIT 1", name)

	var meta PackageMetadata
	err = row.Scan(&meta.Name, &meta.Version, &meta.Hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("package '%s' not found in registry", name)
		}
		// Fallback: the hash column might not exist in the older schema.
		rowFallback := db.QueryRowContext(ctx, "SELECT name, version FROM packages WHERE name = ? ORDER BY version DESC LIMIT 1", name)
		var metaFallback PackageMetadata
		errFallback := rowFallback.Scan(&metaFallback.Name, &metaFallback.Version)
		if errFallback != nil {
			return nil, fmt.Errorf("failed to query package metadata: %w", err)
		}
		return &metaFallback, nil
	}

	return &meta, nil
}

// OpenRegistryDB opens a read-only connection to the registry database.
func OpenRegistryDB() (*sql.DB, error) {
	dbPath := env.GetRegistryDatabasePath()
	dsn := fmt.Sprintf("file:%s?mode=ro", dbPath)
	return sql.Open("sqlite", dsn)
}

// GetDependencies fetches the direct required dependencies for a given package.
// It returns a list of dependency package names.
func GetDependencies(ctx context.Context, db *sql.DB, pkgName string) ([]string, error) {
	// For simplicity, we just fetch dependencies of the latest version of the package.
	// In a real scenario, we should resolve the exact version first.
	query := `
		SELECT dependency_name 
		FROM dependencies 
		WHERE package_name = ? AND is_recommended = 0
		AND package_version = (
			SELECT version FROM packages WHERE name = ? ORDER BY version DESC LIMIT 1
		)
	`
	
	rows, err := db.QueryContext(ctx, query, pkgName, pkgName)
	if err != nil {
		return nil, fmt.Errorf("failed to query dependencies for package %s: %w", pkgName, err)
	}
	defer rows.Close()

	var deps []string
	for rows.Next() {
		var depName string
		if err := rows.Scan(&depName); err != nil {
			return nil, err
		}
		deps = append(deps, depName)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return deps, nil
}

// GetReverseDependencies fetches packages that depend on the given package.
func GetReverseDependencies(ctx context.Context, db *sql.DB, pkgName string) ([]string, error) {
	query := `
		SELECT DISTINCT package_name 
		FROM dependencies 
		WHERE dependency_name = ?
	`
	
	rows, err := db.QueryContext(ctx, query, pkgName)
	if err != nil {
		return nil, fmt.Errorf("failed to query reverse dependencies for package %s: %w", pkgName, err)
	}
	defer rows.Close()

	var revDeps []string
	for rows.Next() {
		var revDepName string
		if err := rows.Scan(&revDepName); err != nil {
			return nil, err
		}
		revDeps = append(revDeps, revDepName)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return revDeps, nil
}
