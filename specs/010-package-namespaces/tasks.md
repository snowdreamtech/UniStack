# Tasks: Package Namespace Support

**Input**: Design documents from `/specs/010-package-namespaces/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Initialize package namespace feature branch

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T002 Implement safe string transformation helper for namespace flattening (if needed as a shared util, otherwise inline)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Publish a namespaced package (Priority: P1) 🎯 MVP

**Goal**: As a package author publishing to the community repository, I want to prefix my package name with my namespace so that it does not collide with official core packages or packages from other authors.

**Independent Test**: Can be fully tested by creating a `package.yml` with `name: snowdreamtech/hello`, running `registry build`, and verifying it is successfully archived without creating unintended deep nested directories.

### Implementation for User Story 1

- [ ] T003 [US1] Modify `scanAndArrangePackages` in `internal/registry/builder.go` to apply `strings.ReplaceAll(name, "/", "_")` when creating output archives.

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently.

---

## Phase 4: User Story 2 - Install and Upgrade a namespaced package (Priority: P1)

**Goal**: As a user, I want to install and upgrade packages that have a namespace prefix, and have the system correctly track its installation status.

**Independent Test**: Can be fully tested by installing a namespaced package twice. The second time should skip installation. Upgrading should successfully remove the old version.

### Implementation for User Story 2

- [ ] T004 [US2] Modify package extraction logic in `internal/client/installer.go` to use flattened safe names for local folder paths.
- [ ] T005 [US2] Ensure `pkgID` assignment in installer logic correctly handles the flat name format for idempotency checks.

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T006 Run quickstart.md validation locally to verify E2E workflow

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - Integrates closely with US1; relies on US1 to build a valid registry archive to test with.

### Implementation Strategy

### MVP First

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1 & Phase 4: User Story 2 (They go hand in hand for full lifecycle)
4. **STOP and VALIDATE**: Test both stories using `quickstart.md`.
5. Deploy/merge if ready.
