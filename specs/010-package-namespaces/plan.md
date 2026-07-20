# Implementation Plan: Package Namespace Support

**Branch**: `010-package-namespaces` | **Date**: 2026-07-20 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/010-package-namespaces/spec.md`

## Summary

Add support for package namespaces (e.g. `snowdreamtech/hello`) to prevent naming collisions in the community registry. The technical approach involves URL-safe flat mapping (replacing `/` with `_`) during local filesystem extraction and registry archive generation to maintain compatibility with 1-level directory scanning, while preserving the original namespace structure in logical IDs, memory, and the SQLite registry.

## Technical Context

**Language/Version**: Go 1.21+

**Primary Dependencies**: standard library (`path/filepath`, `strings`, `os`)

**Storage**: SQLite (`local_repo/repodata/packages.db`), Local Filesystem (`~/.local/share/unistack/`)

**Testing**: Bash script lifecycle testing (`test_full_lifecycle.sh`)

**Target Platform**: Mac/Linux

**Project Type**: CLI Package Manager

**Performance Goals**: N/A

**Constraints**: Must maintain backward compatibility with 1-level deep package scanning in `ListInstalledPackages`.

**Scale/Scope**: Impacts `registry build` and `install`/`upgrade` logic.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

No violations. The implementation simplifies directory structures by mapping slashes to flat strings, avoiding deeply nested directory overhead and complex scanner logic.

## Project Structure

### Documentation (this feature)

```text
specs/010-package-namespaces/
├── plan.md              # This file
├── research.md          # Phase 0
├── data-model.md        # Phase 1
├── quickstart.md        # Phase 1
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
src/
├── internal/registry/builder.go  # Registry package arrangement
├── internal/client/installer.go  # Package ID and extraction path logic
```

**Structure Decision**: The changes are localized to internal package resolution and builder logic.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None      | N/A        | N/A                                 |
