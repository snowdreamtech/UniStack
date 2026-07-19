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
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
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

type scannedFile struct {
	Path     string
	Checksum string
	Pkg      Package
	Version  *semver.Version
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

// Build scans the provided directory, auto-arranges packages, parses package.yml from archives, and inserts into DB
func (b *Builder) Build(ctx context.Context, sourceDir, destDir string) error {
	entries, err := b.scanAndArrangePackages(sourceDir, destDir)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return fmt.Errorf("no valid package archives found in %s", sourceDir)
	}

	return b.insertPackages(ctx, entries)
}

func (b *Builder) scanAndArrangePackages(sourceDir, destDir string) ([]*PackageEntry, error) {
	var entries []*PackageEntry
	packagesDir := filepath.Join(destDir, "packages")

	if err := os.MkdirAll(packagesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create packages directory: %w", err)
	}

	// 1. Collect all tarballs
	scanned := make(map[string][]*scannedFile)

	err := filepath.WalkDir(sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(d.Name(), ".tar.gz") && !strings.HasSuffix(d.Name(), ".uspkg") {
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

		f.Seek(0, 0)
		gzr, err := gzip.NewReader(f)
		if err != nil {
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
			return nil
		}

		var pkg Package
		if err := yaml.Unmarshal(pkgContent, &pkg); err != nil {
			return nil
		}
		if err := Validate(&pkg); err != nil {
			return nil
		}

		name := pkg.Metadata.Name
		verStr := pkg.Metadata.Version
		if len(name) == 0 {
			return nil
		}

		ver, err := semver.NewVersion(verStr)
		if err != nil {
			slog.Warn("invalid semver, skipping", "path", path, "version", verStr)
			return nil
		}

		scanned[name] = append(scanned[name], &scannedFile{
			Path:     path,
			Checksum: checksum,
			Pkg:      pkg,
			Version:  ver,
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 2. Determine max version and arrange
	for name, files := range scanned {
		var maxFile *scannedFile
		for _, sf := range files {
			if maxFile == nil || sf.Version.GreaterThan(maxFile.Version) {
				maxFile = sf
			}
		}

		// Delete losers from sourceDir if sourceDir == destDir
		if sourceDir == destDir {
			for _, sf := range files {
				if sf != maxFile {
					slog.Info("Deleting old version", "path", sf.Path)
					os.Remove(sf.Path)
				}
			}
		}

		// Calculate target path for maxFile
		firstChar := strings.ToLower(string(name[0]))
		expectedRelPath := filepath.Join("packages", firstChar, fmt.Sprintf("%s-%s.tar.gz", name, maxFile.Pkg.Metadata.Version))
		expectedAbsPath := filepath.Join(destDir, expectedRelPath)

		// Before moving/copying, glob the destDir for old versions of THIS package and delete them
		globPattern := filepath.Join(packagesDir, firstChar, name+"-*.tar.gz")
		matches, _ := filepath.Glob(globPattern)
		for _, match := range matches {
			if filepath.Clean(match) == filepath.Clean(expectedAbsPath) {
				continue // Don't delete the file if it's already exactly the max file
			}
			slog.Info("Deleting existing older version in registry", "path", match)
			os.Remove(match)
		}

		// Now move/copy the maxFile to expectedAbsPath if it's not already there
		if filepath.Clean(maxFile.Path) != filepath.Clean(expectedAbsPath) {
			slog.Info("Arranging package", "name", name, "version", maxFile.Pkg.Metadata.Version, "from", maxFile.Path, "to", expectedAbsPath)
			if err := os.MkdirAll(filepath.Dir(expectedAbsPath), 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory for arranged package: %w", err)
			}

			if sourceDir == destDir {
				if err := os.Rename(maxFile.Path, expectedAbsPath); err != nil {
					return nil, fmt.Errorf("failed to move package %s: %w", maxFile.Path, err)
				}
			} else {
				if err := copyFile(maxFile.Path, expectedAbsPath); err != nil {
					return nil, fmt.Errorf("failed to copy package %s: %w", maxFile.Path, err)
				}
			}
		}

		entries = append(entries, &PackageEntry{
			Pkg:          &maxFile.Pkg,
			Hash:         maxFile.Checksum,
			RelativePath: filepath.ToSlash(expectedRelPath),
		})
	}

	return entries, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
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
