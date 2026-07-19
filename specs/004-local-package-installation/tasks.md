# Tasks: Local Package Installation

**Input**: Design documents from `/specs/004-local-package-installation/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup & Foundational (Shared Infrastructure)

**Purpose**: Core paths resolution and utilities required by all installation user stories.

- [ ] T001 Implement `GetInstallPackagesDir()` and `GetInstallBinDir()` in `internal/env/install_paths.go`
- [ ] T002 Implement `ExtractTarGz(src, dest string)` utility using `archive/tar` in `internal/client/extractor.go`
- [ ] T003 Add unit tests for `ExtractTarGz` in `internal/client/extractor_test.go`
- [ ] T004 Implement `CreateSymlink(target, link string)` utility in `internal/client/extractor.go`

**Checkpoint**: Core extraction and linking functions are ready.

---

## Phase 2: User Story 2 - Offline/Local File Installation (Priority: P2, but foundation for US1)

**Goal**: Users can install directly from a local `.tar.gz` file. We do this first because installing from the registry (US1) is just a superset (download + install).

**Independent Test**: Install a local `.tar.gz` and verify binary is linked.

### Implementation for User Story 2

- [ ] T005 [US2] Create `Installer` struct and `InstallFromLocal(pkgPath string) error` in `internal/client/installer.go`
- [ ] T006 [US2] Implement atomic extraction logic (extract to `.tmp-`, then rename) in `InstallFromLocal`
- [ ] T007 [US2] Implement executable discovery and symlinking to `~/.local/bin` in `InstallFromLocal`
- [ ] T008 [US2] Add unit/integration tests for local installation in `internal/client/installer_test.go`
- [ ] T009 [US2] Create `cmd/19.install.go` and register `unistack install <file>` command logic for local file paths

**Checkpoint**: At this point, User Story 2 (Offline install) should be fully functional and testable independently.

---

## Phase 3: User Story 1 - Install a Package by Name (Priority: P1) 🎯 MVP

**Goal**: Users can install a package by simply providing its name, downloading it automatically from the registry.

**Independent Test**: `unistack install <package_name>` works end-to-end.

### Implementation for User Story 1

- [ ] T010 [US1] Extend `cmd/19.install.go` to handle registry lookup when the argument is not a local file
- [ ] T011 [US1] Implement flow: `registry.QueryPackage` -> `client.DownloadPackage` -> `client.InstallFromLocal` in `cmd/19.install.go`
- [ ] T012 [US1] Add integration test for remote package installation in `internal/client/installer_test.go`

**Checkpoint**: `unistack install` now works for both local files and remote packages.

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases and final validation.

- [ ] T013 Add conflict resolution (warn/overwrite) when symlink already exists in `internal/client/installer.go`
- [ ] T014 Run validation steps from `quickstart.md`
- [ ] T015 Ensure no CGO dependencies were introduced

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Must complete first.
- **User Story 2 (Phase 2)**: Depends on Setup. We implement US2 first because it contains the core installation logic.
- **User Story 1 (Phase 3)**: Depends on US2 (it wraps download + local install).
- **Polish (Final Phase)**: Depends on all user stories.

### Parallel Opportunities

- Tests (T003, T008, T012) can be worked on in parallel with their implementations.
- Extraction utility (T002) and path resolution (T001) can be implemented in parallel.

## Implementation Strategy

### MVP First (User Story 1 & 2)

1. Complete Phase 1: Setup
2. Complete Phase 2: Local Install (US2)
3. Validate local installation.
4. Complete Phase 3: Remote Install (US1)
5. Run full E2E testing using `quickstart.md`.
