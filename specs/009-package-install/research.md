# Research: Package Installation & Dynamic Download Sourcing

## Decision: `Source` field in `PackageMetadata`
- **Rationale**: The core issue is that `unistack install` does not know which registry source to pull from because the `QueryPackage` function strips out the `source` column present in the SQLite view. Injecting this field directly into `PackageMetadata` allows downstream systems (like `installer.go`) to look up the exact base URL of the source repository (e.g., `file:///path/to/local` or `https://registry.unistack.org`).
- **Alternatives considered**: Passing the raw URL inside the view, or keeping an in-memory map. Putting the Source identifier into `PackageMetadata` is the cleanest and most loosely coupled method.

## Decision: Dynamic URL Mapping in Installer
- **Rationale**: In `internal/client/installer.go`, the hardcoded `http://localhost:8080` must be removed. By looking up the configured sources (using `config.LoadSources()`), the installer can match `meta.Source` to the source config and extract its true `URL`.
- **Alternatives considered**: Passing a list of URLs directly into `InstallPackage`. Dynamically fetching from the config ensures the single source of truth is always respected.
