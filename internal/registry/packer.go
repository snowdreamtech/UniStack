// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package registry

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Pack scans the source directory for roles containing package.yml
// and packages them into the destination directory following the spec.
func Pack(ctx context.Context, sourceDir, destDir string) error {
	packagesDir := filepath.Join(destDir, "packages")

	// Create packages directory if not exists
	if err := os.MkdirAll(packagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create packages directory: %w", err)
	}

	// Walk through the source directory to find package.yml
	err := filepath.WalkDir(sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Only look for package.yml directly inside role directories
		if d.IsDir() || d.Name() != "package.yml" {
			return nil
		}

		roleDir := filepath.Dir(path)

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		var pkg Package
		if err := yaml.Unmarshal(content, &pkg); err != nil {
			slog.Warn("failed to parse package.yml, skipping", "path", path, "error", err)
			return nil
		}

		if err := Validate(&pkg); err != nil {
			slog.Warn("validation failed for package.yml, skipping", "path", path, "error", err)
			return nil
		}

		name := pkg.Metadata.Name
		version := pkg.Metadata.Version
		
		if len(name) == 0 {
			slog.Warn("package name is empty, skipping", "path", path)
			return nil
		}

		// First character of the package name
		firstChar := strings.ToLower(string(name[0]))
		
		// Target filename: packages/<first_char>/<name>-<version>.tar.gz
		targetDir := filepath.Join(packagesDir, firstChar)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create target directory %s: %w", targetDir, err)
		}

		targetFile := filepath.Join(targetDir, fmt.Sprintf("%s-%s.tar.gz", name, version))
		slog.Info("Packaging role", "name", name, "version", version, "source", roleDir, "target", targetFile)

		if err := tarGzDir(roleDir, targetFile); err != nil {
			return fmt.Errorf("failed to package %s: %w", name, err)
		}

		return nil
	})

	return err
}

func tarGzDir(sourceDir, targetFile string) error {
	f, err := os.Create(targetFile)
	if err != nil {
		return err
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return filepath.Walk(sourceDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Ignore root dir itself
		if file == sourceDir {
			return nil
		}

		// Get relative path for tar header
		relPath, err := filepath.Rel(sourceDir, file)
		if err != nil {
			return err
		}

		// Fix slashes for cross-platform compatibility inside tar
		relPath = filepath.ToSlash(relPath)

		// Create header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}
		
		// Use relative path
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// If it's a regular file, write contents
		if fi.Mode().IsRegular() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			defer data.Close()

			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		
		return nil
	})
}
