// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package registry

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	_ "modernc.org/sqlite"
)

// Builder defines the registry builder instance
type Builder struct {
	db *sql.DB
}

// PackageEntry wraps a Package with metadata about its physical file.
type PackageEntry struct {
	Pkg          *Package
	Hash         string
	RelativePath string
}

// NewBuilder initializes a new registry builder
func NewBuilder(dbPath string) (*Builder, error) {
	// Remove existing db if we are rebuilding
	os.Remove(dbPath)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	if err := initSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Builder{db: db}, nil
}

// Close closes the database connection
func (b *Builder) Close() error {
	return b.db.Close()
}

// Build scans the provided directory, parses package.yml from archives, and inserts into DB
func (b *Builder) Build(ctx context.Context, sourceDir string) error {
	entries, err := b.scanPackages(sourceDir)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return fmt.Errorf("no valid package archives found in %s", sourceDir)
	}

	return b.insertPackages(ctx, entries)
}

func (b *Builder) scanPackages(sourceDir string) ([]*PackageEntry, error) {
	var entries []*PackageEntry
	packagesDir := filepath.Join(sourceDir, "packages")

	// If packages/ directory doesn't exist, fallback to scanning sourceDir directly
	scanDir := packagesDir
	if _, err := os.Stat(packagesDir); os.IsNotExist(err) {
		scanDir = sourceDir
	}

	err := filepath.WalkDir(scanDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		isTarball := strings.HasSuffix(d.Name(), ".tar.gz") || strings.HasSuffix(d.Name(), ".uspkg")
		if !isTarball {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", path, err)
		}
		defer f.Close()

		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return fmt.Errorf("failed to hash %s: %w", path, err)
		}
		checksum := "sha256:" + hex.EncodeToString(h.Sum(nil))

		if _, err := f.Seek(0, 0); err != nil {
			return fmt.Errorf("failed to rewind %s: %w", path, err)
		}

		gzr, err := gzip.NewReader(f)
		if err != nil {
			fmt.Printf("Warning: failed to read gzip %s: %v\n", path, err)
			return nil
		}
		defer gzr.Close()

		tr := tar.NewReader(gzr)
		var pkgContent []byte
		for {
			header, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("Warning: failed to read tar %s: %v\n", path, err)
				return nil
			}

			if filepath.Base(header.Name) == "package.yml" {
				pkgContent, err = io.ReadAll(tr)
				if err != nil {
					return fmt.Errorf("failed to read package.yml in %s: %w", path, err)
				}
				break
			}
		}

		if pkgContent == nil {
			fmt.Printf("Warning: no package.yml found in %s\n", path)
			return nil
		}

		var pkg Package
		if err := yaml.Unmarshal(pkgContent, &pkg); err != nil {
			fmt.Printf("Warning: failed to parse package.yml in %s: %v\n", path, err)
			return nil
		}

		if err := Validate(&pkg); err != nil {
			fmt.Printf("Warning: validation failed for %s: %v\n", path, err)
			return nil
		}

		relPath, _ := filepath.Rel(sourceDir, path)
		// Fix windows paths for web URL compatibility
		relPath = filepath.ToSlash(relPath)

		entries = append(entries, &PackageEntry{
			Pkg:          &pkg,
			Hash:         checksum,
			RelativePath: relPath,
		})
		return nil
	})

	return entries, err
}

func (b *Builder) insertPackages(ctx context.Context, entries []*PackageEntry) error {
	tx, err := b.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	pkgStmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO packages (name, version, description, authors, homepage, license, hash, relative_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer pkgStmt.Close()

	depStmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO dependencies (package_name, package_version, dependency_name, version_constraint, is_recommended)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer depStmt.Close()

	for _, entry := range entries {
		pkg := entry.Pkg
		authors := strings.Join(pkg.Metadata.Authors, ",")

		_, err := pkgStmt.ExecContext(ctx,
			pkg.Metadata.Name,
			pkg.Metadata.Version,
			pkg.Metadata.Description,
			authors,
			pkg.Metadata.Homepage,
			pkg.Metadata.License,
			entry.Hash,
			entry.RelativePath,
		)
		if err != nil {
			return fmt.Errorf("failed to insert package %s: %w", pkg.Metadata.Name, err)
		}

		for depName, constraint := range pkg.Dependencies.Required {
			_, err := depStmt.ExecContext(ctx, pkg.Metadata.Name, pkg.Metadata.Version, depName, constraint, false)
			if err != nil {
				return fmt.Errorf("failed to insert dependency for %s: %w", pkg.Metadata.Name, err)
			}
		}

		for depName, constraint := range pkg.Dependencies.Recommended {
			_, err := depStmt.ExecContext(ctx, pkg.Metadata.Name, pkg.Metadata.Version, depName, constraint, true)
			if err != nil {
				return fmt.Errorf("failed to insert recommended dependency for %s: %w", pkg.Metadata.Name, err)
			}
		}
	}

	return tx.Commit()
}

func initSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS packages (
		name TEXT NOT NULL,
		version TEXT NOT NULL,
		description TEXT,
		authors TEXT,
		homepage TEXT,
		license TEXT,
		hash TEXT NOT NULL,
		relative_path TEXT NOT NULL,
		PRIMARY KEY (name, version)
	);

	CREATE TABLE IF NOT EXISTS dependencies (
		package_name TEXT NOT NULL,
		package_version TEXT NOT NULL,
		dependency_name TEXT NOT NULL,
		version_constraint TEXT NOT NULL,
		is_recommended BOOLEAN NOT NULL DEFAULT 0,
		PRIMARY KEY (package_name, package_version, dependency_name),
		FOREIGN KEY(package_name, package_version) REFERENCES packages(name, version)
	);
	`

	_, err := db.Exec(schema)
	return err
}
