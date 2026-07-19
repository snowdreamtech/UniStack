# Implementation Plan: Dependency Resolution Engine

**Branch**: `007-dependency-resolution` | **Date**: 2026-07-19 | **Spec**: [specs/007-dependency-resolution/spec.md](specs/007-dependency-resolution/spec.md)

**Input**: Feature specification from `/specs/007-dependency-resolution/spec.md`

## Summary

Implement a dependency graph resolution algorithm (DAG) for the UniStack Go client to ensure when installing a package, all of its dependencies are recursively downloaded and installed first. We will use topological sorting and circular dependency detection, querying the local SQLite Registry DB's `dependencies` table to construct the graph.

## Technical Context

**Language/Version**: Go 1.21+

**Primary Dependencies**: `github.com/snowdreamtech/unistack/internal/registry`, `github.com/snowdreamtech/unistack/internal/client`

**Storage**: Local SQLite database (`registry.db` / `packages.db`)

**Testing**: Go standard testing package (`testing`)

**Target Platform**: Linux, macOS, Windows

**Project Type**: CLI tool (Package Manager)

**Performance Goals**: Dependency graph resolution and cycle detection under 100ms.

**Constraints**: Must halt on circular dependencies gracefully. Skip already installed/satisfied dependencies.

**Scale/Scope**: Typically small dependency trees (under 100 nodes).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*
All principles (Go, CLI, Test-First, Error handling) are met.

## Project Structure

### Documentation (this feature)

```text
specs/007-dependency-resolution/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
└── quickstart.md        # Phase 1 output
```

### Source Code (repository root)

```text
internal/
├── registry/
│   └── query.go            # (Existing DB query logic, add method to fetch dependencies)
├── client/
│   ├── resolver.go         # NEW: DAG and dependency resolution logic
│   ├── resolver_test.go    # NEW: Tests for DAG, topo sort, cycle detection
│   └── installer.go        # MODIFIED: Hook resolver into install process
```

**Structure Decision**: Add a new `resolver.go` in `internal/client/` to handle graph traversal, keeping it decoupled from DB queries but integrated with installation logic.
