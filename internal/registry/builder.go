package registry

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
	"gopkg.in/yaml.v3"
)

// Builder defines the registry builder instance
type Builder struct {
	db *sql.DB
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

// Build scans the provided directory, parses package.yml, and inserts into DB
func (b *Builder) Build(ctx context.Context, sourceDir string) error {
	packages, err := b.scanPackages(sourceDir)
	if err != nil {
		return err
	}

	if len(packages) == 0 {
		return fmt.Errorf("no valid packages found in %s", sourceDir)
	}

	return b.insertPackages(ctx, packages)
}

func (b *Builder) scanPackages(sourceDir string) ([]*Package, error) {
	var pkgs []*Package

	err := filepath.WalkDir(sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || d.Name() != "package.yml" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		var pkg Package
		if err := yaml.Unmarshal(content, &pkg); err != nil {
			// Print warning and skip
			fmt.Printf("Warning: failed to parse %s: %v\n", path, err)
			return nil
		}

		if err := Validate(&pkg); err != nil {
			fmt.Printf("Warning: validation failed for %s: %v\n", path, err)
			return nil
		}

		pkgs = append(pkgs, &pkg)
		return nil
	})

	return pkgs, err
}

func (b *Builder) insertPackages(ctx context.Context, packages []*Package) error {
	tx, err := b.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	pkgStmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO packages (name, version, description, authors, homepage, license)
		VALUES (?, ?, ?, ?, ?, ?)
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

	for _, pkg := range packages {
		authors := strings.Join(pkg.Metadata.Authors, ",")
		
		_, err := pkgStmt.ExecContext(ctx,
			pkg.Metadata.Name,
			pkg.Metadata.Version,
			pkg.Metadata.Description,
			authors,
			pkg.Metadata.Homepage,
			pkg.Metadata.License,
		)
		if err != nil {
			return fmt.Errorf("failed to insert package %s: %w", pkg.Metadata.Name, err)
		}

		// Insert required dependencies
		for depName, constraint := range pkg.Dependencies.Required {
			_, err := depStmt.ExecContext(ctx, pkg.Metadata.Name, pkg.Metadata.Version, depName, constraint, false)
			if err != nil {
				return fmt.Errorf("failed to insert dependency for %s: %w", pkg.Metadata.Name, err)
			}
		}

		// Insert recommended dependencies
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
