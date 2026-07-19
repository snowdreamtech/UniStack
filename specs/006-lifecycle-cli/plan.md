# Implementation Plan: Lifecycle CLI (Uninstall, Upgrade, List)

**Feature**: [006-lifecycle-cli]
**Status**: Approved

## 1. Technical Context

### Architecture & Boundaries
The feature extends the CLI interface in `cmd/unistack/` and the core client operations in `internal/client/`.
We will add:
- `cmd/unistack/14.list.go`: Implementation of `unistack list`.
- `cmd/unistack/15.uninstall.go`: Implementation of `unistack uninstall`.
- `cmd/unistack/16.upgrade.go`: Implementation of `unistack upgrade`.
Note: The file numbering in `cmd/unistack/` may vary depending on existing commands.

### Dependencies
- Standard library (`os`, `path/filepath`, `fmt`).
- Existing `internal/client` for operations.
- Existing `internal/repository` for SQLite access.
- Existing `internal/envpath` for UniStack directory paths (`~/.local/share/unistack`).

## 2. Constitution Check

- **Zero CGO**: Ensured. We only use pure Go standard library and SQLite driver (already pure Go/configured).
- **Security**: Removing files is dangerous. We must ensure paths are sanitized and restricted to `~/.local/share/unistack` before deletion.

## 3. Data Model

- SQLite `packages` table contains the remote registry state.
- Local state: `~/.local/share/unistack/packages/<pkg>-<version>` exists. Symlinks exist in `~/.local/share/unistack/bin`.

## 4. Phase 1: List Command

- Read directory `~/.local/share/unistack/packages/`.
- Parse directory names to extract `<pkg>` and `<version>`.
- Print to standard output in a tabular format.

## 5. Phase 2: Uninstall Command

- Accept package name.
- Locate package directory.
- Check if `app_loader.yml` exists. If so, execute `ansible-playbook -i localhost, -c local app_loader.yml -e app_source_path=<path> -e state=absent`.
- Remove symlinks from `~/.local/share/unistack/bin`.
- Remove package directory from `~/.local/share/unistack/packages`.

## 6. Phase 3: Upgrade Command

- Accept package name.
- Query local SQLite database for the highest version.
- Determine currently installed version.
- If current < highest: call Uninstall(pkg), then Install(pkg, highest_version).
  ├── models/
│   ├── services/
│   └── api/
└── tests/

frontend/
│   ├── components/
│   ├── pages/
│   └── services/
└── tests/

# [REMOVE IF UNUSED] Option 3: Mobile + API (when "iOS/Android" detected)
api/
└── [same as backend above]

ios/ or android/
└── [platform-specific structure: feature modules, UI flows, platform tests]
```

**Structure Decision**: [Document the selected structure and reference the real
directories captured above]

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
