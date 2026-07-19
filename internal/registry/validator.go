package registry

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// Validate ensures that the package contains valid semantic versions
func Validate(pkg *Package) error {
	if pkg.Metadata.Name == "" {
		return fmt.Errorf("package name is required")
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
