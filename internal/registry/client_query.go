// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package registry

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/snowdreamtech/unistack/internal/config"
	"github.com/snowdreamtech/unistack/internal/env"

	// SQLite driver
	_ "modernc.org/sqlite"
)

// PackageMetadata represents the metadata of a package needed for downloading.
type PackageMetadata struct {
	Name    string
	Version string
	Hash    string
	Source  string
}

// OpenRegistryDB opens an in-memory SQLite connection that attaches all configured registry sources
// and creates unified views (packages, dependencies) for transparent querying across all sources.
func OpenRegistryDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		return nil, fmt.Errorf("failed to open memory db: %w", err)
	}

	sources, err := config.LoadSources()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to load sources: %w", err)
	}

	var packagesSelects []string
	var dependenciesSelects []string

	for i, s := range sources {
		dbPath := env.GetSourceDatabasePath(s.Name)
		if _, err := os.Stat(dbPath); err != nil {
			continue // Source db not downloaded yet
		}

		dbAlias := fmt.Sprintf("db_%d", i)
		attachQuery := fmt.Sprintf("ATTACH DATABASE 'file:%s?mode=ro' AS %s;", dbPath, dbAlias)
		if _, err := db.Exec(attachQuery); err != nil {
			slog.Warn("Failed to attach source database", "source", s.Name, "error", err)
			continue
		}

		packagesSelects = append(packagesSelects, fmt.Sprintf("SELECT *, '%s' as source FROM %s.packages", s.Name, dbAlias))
		dependenciesSelects = append(dependenciesSelects, fmt.Sprintf("SELECT *, '%s' as source FROM %s.dependencies", s.Name, dbAlias))
	}

	if len(packagesSelects) > 0 {
		viewQuery := "CREATE TEMP VIEW packages AS " + strings.Join(packagesSelects, " UNION ALL ")
		if _, err := db.Exec(viewQuery); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to create packages view: %w", err)
		}
	} else {
		db.Exec("CREATE TEMP TABLE packages (id INTEGER, name TEXT, version TEXT, description TEXT, homepage TEXT, license TEXT, hash TEXT, source TEXT)")
	}

	if len(dependenciesSelects) > 0 {
		viewQuery := "CREATE TEMP VIEW dependencies AS " + strings.Join(dependenciesSelects, " UNION ALL ")
		if _, err := db.Exec(viewQuery); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to create dependencies view: %w", err)
		}
	} else {
		db.Exec("CREATE TEMP TABLE dependencies (id INTEGER, package_name TEXT, package_version TEXT, dependency_name TEXT, is_recommended INTEGER, source TEXT)")
	}

	return db, nil
}

// QueryPackage retrieves the metadata of the latest version of a package from the unified local registry DB.
func QueryPackage(ctx context.Context, name string) (*PackageMetadata, error) {
	db, err := OpenRegistryDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row := db.QueryRowContext(ctx, "SELECT name, version, COALESCE(hash, ''), source FROM packages WHERE name = ? ORDER BY version DESC LIMIT 1", name)

	var meta PackageMetadata
	err = row.Scan(&meta.Name, &meta.Version, &meta.Hash, &meta.Source)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("package '%s' not found in registry", name)
		}
		// Fallback: the hash or source column might not exist in the older schema.
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

// GetDependencies fetches the direct required dependencies for a given package.
func GetDependencies(ctx context.Context, db *sql.DB, pkgName string) ([]string, error) {
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
