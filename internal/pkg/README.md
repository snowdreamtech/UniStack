# Internal Packages

This directory contains shared internal infrastructure packages used across the UniStack project template. These modules are strictly decoupled from business logic, making them universally applicable to any CLI tool.

## Packages

- **`download`**: A robust HTTP downloader with automatic checksum validation, retry mechanics, progress reporting, context cancellation, and concurrent proxy support.
- **`env`**: Environment and build metadata management (e.g., ProjectName, Version, GitTag, BuildTime).
- **`envpath`**: Safe cross-platform environment path resolution utilities.
- **`errors`**: Industrial-grade error categorization (User/System/External) with automated mapping to standard shell exit codes.
- **`gpg`**: GPG signature verification abstraction using both system `gpg` and a native Go implementation (`gopenpgp`).
- **`http`**: Robust HTTP client with smart proxy resolution and domestic mirror bypass.
- **`logger`**: Zero-dependency structured logger built on top of `log/slog` with seamless CLI flag integration.
- **`version`**: Semantic versioning (SemVer) parser supporting exact versions, ranges (`>=`, `^`, `~`), and aliases (`latest`, `lts`, `stable`).
