# Implementation Plan: 客户端对软件仓库的集成

**Branch**: `003-client-registry-integration` | **Date**: 2026-07-19 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/003-client-registry-integration/spec.md`

## Summary

实现客户端通过 `unistack update` 拉取 zstd 压缩的 SQLite 数据库缓存到本地，并通过 `unistack download` 提取指定包的下载元数据、完成 `.tar.gz` 离线包下载以及 SHA-256 哈希校验。

## Technical Context

**Language/Version**: Go 1.22

**Primary Dependencies**: `modernc.org/sqlite` (DB queries), `github.com/klauspost/compress/zstd` (Decompression), `crypto/sha256` (Hash check).

**Storage**: Local cache SQLite DB (`~/.unistack/packages.db` or defined by config) and local file cache (`~/.unistack/cache/`).

**Testing**: Go unit tests (mock HTTP) & integration tests (local CLI runs).

**Target Platform**: Linux, macOS, Windows (cross-platform pure Go).

**Project Type**: CLI tool.

**Performance Goals**: `unistack update` finishes in <2s on fast networks. Decompression and query should be in milliseconds.

**Constraints**: Must strictly use zero CGO. No dependency on OS-level `curl`, `wget`, or `tar` for the Go portion.

**Scale/Scope**: Can handle SQLite databases of ~100MB with 10k+ packages effortlessly.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Project must remain pure Go.
- Must ensure cross-platform compatibility (mac, linux, windows).

## Project Structure

### Documentation (this feature)

```text
specs/003-client-registry-integration/
├── plan.md              # This file
├── research.md
├── data-model.md
├── quickstart.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/
├── 17.update.go            # `unistack update` CLI
└── 18.download.go          # `unistack download` CLI

internal/
├── client/                 # Client download and network operations
│   ├── downloader.go
│   └── hash.go
└── registry/
    └── client_query.go     # Local DB querying logic
```

**Structure Decision**:
The code is placed in `cmd/` for the new CLI commands, and `internal/client/` to handle network requests (HTTP GET, retry, download) and file operations. Queries to the synced local DB remain in `internal/registry/`.

## Complexity Tracking

N/A
