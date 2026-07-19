# Tasks: Dependency Resolution Engine

**Input**: Design documents from `/specs/007-dependency-resolution/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Create `internal/client/resolver.go` and `internal/client/resolver_test.go` per implementation plan

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T002 Implement `GetDependencies(ctx context.Context, db *sql.DB, pkgName string) ([]string, error)` in `internal/registry/query.go` to fetch direct dependencies from the SQLite `dependencies` table.
- [ ] T003 Implement `GetReverseDependencies(ctx context.Context, db *sql.DB, pkgName string) ([]string, error)` in `internal/registry/query.go` to fetch packages that depend on the given package.

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Recursive Installation (Priority: P1) 🎯 MVP

**Goal**: When a user installs package A (which depends on B and C), the client automatically resolves, downloads, and installs B and C before A.

**Independent Test**: `unistack install A` successfully installs C, then B, then A (based on dependency tree).

### Implementation for User Story 1

- [ ] T004 [US1] Create `DependencyGraph` and `Node` structs in `internal/client/resolver.go` based on data-model.md
- [ ] T005 [US1] Implement `BuildGraph(targetPkg string)` in `resolver.go` using `registry.GetDependencies` to recursively build the graph
- [ ] T006 [US1] Implement `TopologicalSort()` in `resolver.go` using Kahn's algorithm to return the ordered installation list
- [ ] T007 [P] [US1] Add unit tests for successful graph building and topological sort in `internal/client/resolver_test.go`
- [ ] T008 [US1] Update `InstallFromRemote` and `InstallFromLocal` in `internal/client/installer.go` to call the resolver and loop through the ordered list to download/install dependencies first.
- [ ] T009 [US1] Add logic to skip already installed dependencies (by querying local state/manifest)

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Circular Dependency Detection (Priority: P2)

**Goal**: When a user installs a package with a circular dependency, the client gracefully aborts with an error message within 1 second.

**Independent Test**: `unistack install D` (where D depends on E, which depends on D) aborts with `circular dependency detected in package graph`.

### Implementation for User Story 2

- [ ] T010 [P] [US2] Ensure `TopologicalSort()` in `resolver.go` returns `ErrCircularDependency` if the sorted slice length does not equal the total number of graph nodes.
- [ ] T011 [P] [US2] Add unit tests for circular dependency scenarios in `internal/client/resolver_test.go`
- [ ] T012 [US2] Update `installer.go` to catch `ErrCircularDependency` and gracefully abort the installation process, printing a user-friendly error.

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Safe Uninstallation (Priority: P3)

**Goal**: When a user uninstalls package B (which package A depends on), the client warns the user or prevents uninstallation.

**Independent Test**: `unistack uninstall B` (when A is installed) aborts with a warning about breaking dependencies.

### Implementation for User Story 3

- [ ] T013 [US3] Update `Uninstall` in `internal/client/installer.go` to query `registry.GetReverseDependencies` for the target package.
- [ ] T014 [US3] Filter reverse dependencies to check if any of them are currently installed locally (via `ListInstalledPackages` or local state).
- [ ] T015 [US3] Return an error or warning if active installed dependents exist, aborting the uninstallation.
- [ ] T016 [P] [US3] Add unit tests covering safe uninstallation logic in `installer_test.go`.

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T017 [P] Run quickstart.md validation locally to ensure E2E scenarios work (Scenario 1 & 2).
- [ ] T018 Code cleanup and verify error messages are consistent.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - US1, US2, US3 can technically be developed in parallel if stubs are used, but US2 and US3 rely on the resolver foundation from US1.
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### Parallel Opportunities

- Foundation tasks (T002, T003) can run in parallel.
- Tests (T007, T011, T016) can be written in parallel.
- `resolver.go` logic (T004-T006) and `installer.go` modifications (T008-T009) can be worked on concurrently if interfaces are agreed upon.
