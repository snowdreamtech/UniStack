// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package registry

import (
	"fmt"
	"regexp"

	"github.com/Masterminds/semver/v3"
)

// Validate ensures that the package contains valid semantic versions
func Validate(pkg *Package) error {
	if pkg.Metadata.Name == "" {
		return fmt.Errorf("package name is required")
	}

	// Validate package name format: allow at most one level of namespace (e.g. "namespace/pkg" or "pkg")
	// and restrict characters to lowercase alphanumeric and dashes.
	nameRegex := regexp.MustCompile(`^[a-z0-9-]+(?:/[a-z0-9-]+)?$`)
	if !nameRegex.MatchString(pkg.Metadata.Name) {
		return fmt.Errorf("invalid package name '%s': must consist of lowercase alphanumeric characters or dashes, and contain at most one slash (e.g. 'pkg' or 'namespace/pkg')", pkg.Metadata.Name)
	}

	// Strict semver validation for the package version
	if _, err := semver.StrictNewVersion(pkg.Metadata.Version); err != nil {
		return fmt.Errorf("invalid package version '%s': %w", pkg.Metadata.Version, err)
	}

	// Validate appVersion ranges
	for _, appVer := range pkg.Metadata.AppVersion {
		if _, err := semver.NewConstraint(appVer); err != nil {
			return fmt.Errorf("invalid appVersion constraint '%s': %w", appVer, err)
		}
	}

	// Validate dependency constraints
	for depName, depConstraint := range pkg.Dependencies.Required {
		if _, err := semver.NewConstraint(depConstraint); err != nil {
			return fmt.Errorf("invalid required dependency constraint for '%s' ('%s'): %w", depName, depConstraint, err)
		}
	}

	for depName, depConstraint := range pkg.Dependencies.Recommended {
		if _, err := semver.NewConstraint(depConstraint); err != nil {
			return fmt.Errorf("invalid recommended dependency constraint for '%s' ('%s'): %w", depName, depConstraint, err)
		}
	}

	return nil
}
