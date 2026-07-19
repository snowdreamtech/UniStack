# Data Model: Local Package Installation

## Entities

### `InstallationPaths` (Concept)
- **BasePackageDir**: `~/.local/share/unistack/packages`
- **PackageVersionDir**: `BasePackageDir/<name>-<version>`
- **BinDir**: `~/.local/bin`
- **TempExtractDir**: `BasePackageDir/.tmp-<name>-<version>`

### `PackageManifest` (Future / Implicit)
Currently, we assume the `.tar.gz` contains an executable that matches the package name inside a `bin/` directory or at the root. We do not have a robust manifest format yet, but the implicit contract is:
- **Expected Executable Location**: `<extract_root>/<package_name>` or `<extract_root>/bin/<package_name>`
