# Implementation Plan: Registry Sources CRUD

**Branch**: `008-registry-sources` | **Date**: 2026-07-19 | **Spec**: [specs/008-registry-sources/spec.md](file:///Users/snowdream/Workspace/snowdreamtech/UniStack/specs/008-registry-sources/spec.md)

**Input**: Feature specification from `/specs/008-registry-sources/spec.md`

## Summary

This feature provides client-side "Source/Repo" management for UniStack. The user needs to add, remove, list, and update custom registry URLs via a new `unistack source` command. Multiple configured sources will be saved locally (e.g., in `sources.json`). When `unistack update` is invoked, it will fetch packages from all configured sources and store their DBs separately. The query mechanisms will attach all these DBs in-memory so `install` or `search` can resolve packages seamlessly.

## Technical Context

**Language/Version**: Go 1.22+

**Primary Dependencies**: 
- `github.com/spf13/cobra` (CLI)
- `modernc.org/sqlite` (Database)

**Storage**: Local JSON configuration (`sources.json`) and SQLite database files (`*.db`).

**Testing**: Standard `testing` package in Go.

**Target Platform**: Linux, macOS, Windows (Command Line Interface).

**Project Type**: CLI Application.

**Performance Goals**: <1s for configuration changes.

**Constraints**: Seamless backwards compatibility if no configuration exists (default to `https://registry.unistack.org`).

**Scale/Scope**: <10 package sources expected per user.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

No complex patterns are introduced. The approach relies purely on local JSON parsing and standard SQLite `ATTACH DATABASE` mechanisms which adhere to simplicity and minimal magic.

## Project Structure

### Documentation (this feature)

```text
specs/008-registry-sources/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
└── quickstart.md        # Phase 1 output
```

### Source Code

```text
cmd/
├── 17.update.go            # Modified: trigger multi-source update
└── 20.source.go            # New: add, remove, list, update commands

internal/
├── config/
│   └── sources.go          # New: sources.json configuration loader
├── client/
│   ├── updater.go          # Modified: sync all sources into separate .db files
│   └── updater_test.go     
└── registry/
    ├── client_query.go     # Modified: dynamic SQLite ATTACH DATABASE for all .db files
    └── client_query_test.go
```

**Structure Decision**: The CLI commands map logically to `cmd/20.source.go` following the numeric precedence of existing commands. The core logic fits elegantly into `config`, `client`, and `registry` internal packages.
