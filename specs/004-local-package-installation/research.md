# Research: Local Package Installation

## Objective

Determine the best pure-Go approach to extract `.tar.gz` files and establish symlinks across operating systems, ensuring atomicity and clean error recovery.

## Findings

### Extractor Engine

- **Decision**: Use Go's standard library `archive/tar` and `compress/gzip`.
- **Rationale**: `.tar.gz` format is ubiquitous for software distribution, and Go's standard library has full support without external dependencies or CGO.
- **Alternatives considered**:
  - Using external commands like `tar xzf` via `os/exec` (Rejected due to non-portability on Windows without WSL/Git Bash).

### Symlink / Executable Placement

- **Decision**: Use `os.Symlink` on macOS/Linux.
- **Rationale**: This is the standard mechanism used by package managers like Homebrew or NPM to link installed binaries into `~/.local/bin` or equivalent PATH directories.
- **Alternatives considered**:
  - Copying the executable directly (Rejected: breaks relative paths to libraries/assets within the extracted package directory).
  - Wrapper scripts (Considered for Windows fallback).

### Atomicity

- **Decision**: Extract to a temporary directory `~/.local/share/unistack/packages/.tmp-<name>-<version>`, then `os.Rename` to final directory upon success.
- **Rationale**: `os.Rename` is generally atomic on POSIX filesystems. If the process is killed midway, the `.tmp-` directory can be safely deleted on next run or ignored.
