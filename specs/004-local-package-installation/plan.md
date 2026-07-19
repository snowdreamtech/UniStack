# Implementation Plan: Local Package Installation

**Branch**: `004-local-package-installation` | **Date**: 2026-07-19 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/004-local-package-installation/spec.md`

## Summary

This feature implements the `unistack install` command. It takes a package name or a local `.tar.gz` file, downloads it if necessary (via the existing `DownloadPackage` utility), extracts it to `~/.local/share/unistack/packages/<name>-<version>`, and creates a symlink from `~/.local/bin/<executable>` to the extracted binary.

## Technical Context

**Language/Version**: Go 1.21+

**Primary Dependencies**: `archive/tar`, `github.com/klauspost/compress/zstd` (for registry, not tarballs, but standard gzip/zlib is in stdlib for `.tar.gz`)

**Storage**: Local Filesystem (`~/.local/share/unistack/packages/` and `~/.local/bin/`)

**Testing**: Standard Go `testing`, integration tests via local files.

**Target Platform**: Linux, macOS (Windows support via wrapper scripts or native symlinks if possible, but symlinks work best on Unix).

**Project Type**: CLI Package Manager

**Performance Goals**: <5s installation from local cache.

**Constraints**: Pure Go (Zero CGO), atomic installation, clean error recovery.

**Scale/Scope**: Local filesystem operations, symlinking.

## Constitution Check

*GATE: Passed.* 
- Adheres to pure Go zero CGO requirement.
- Follows existing directory structure logic from `internal/env`.

## Project Structure

### Documentation (this feature)

```text
specs/004-local-package-installation/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
cmd/
├── 19.install.go        # Cobra command for unistack install

internal/
├── client/
│   ├── installer.go     # Core installation logic (extraction & symlinking)
│   ├── installer_test.go
│   ├── extractor.go     # tar.gz extraction utility
│   └── extractor_test.go
└── env/
    └── install_paths.go # Path resolution for ~/.local/share/unistack/packages and ~/.local/bin
```

**Structure Decision**: Extending the existing `internal/client` and `internal/env` packages. Creating a new `installer.go` specifically for extraction and linking, independent of `downloader.go`.
