# Tasks: Localize UniGo to UniStack

**Input**: Design documents from `/specs/012-localize-unistack/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, quickstart.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Rename directory `pypi/snowdreamtech_unigo` to `pypi/snowdreamtech_unistack`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

*No foundational blocking infrastructure for this localization chore.*

---

## Phase 3: User Story 1 - Replace Package References (Priority: P1) 🎯 MVP

**Goal**: Ensure all PyPI and CLI Python code correctly uses `unistack`.

**Independent Test**: Build the PyPI package locally and check metadata (quickstart.md).

### Implementation for User Story 1

- [ ] T012 [P] [US1] Replace `unigo` with `unistack` in `pypi/pyproject.toml`
- [ ] T013 [P] [US1] Replace `unigo` with `unistack` in `pypi/setup.py`
- [ ] T014 [P] [US1] Replace `unigo` with `unistack` in `pypi/scripts/build.sh`
- [ ] T015 [P] [US1] Replace `unigo` with `unistack` in `pypi/snowdreamtech_unistack/__init__.py` (if any references exist)
- [ ] T016 [P] [US1] Replace `unigo` with `unistack` in `pypi/snowdreamtech_unistack/cli.py`

**Checkpoint**: At this point, PyPI package references are updated.

---

## Phase 4: User Story 2 - Update CI Workflows and Docs (Priority: P2)

**Goal**: Ensure automated workflows and documentation correctly reflect the localized name.

**Independent Test**: Review generated Markdown documentation and workflow proxy paths.

### Implementation for User Story 2

- [ ] T020 [P] [US2] Replace `unigo` with `unistack` in `.github/workflows/goreleaser.yml` (proxy warmup URL)
- [ ] T021 [P] [US2] Replace `unigo` with `unistack` in `README.md`
- [ ] T022 [P] [US2] Replace `unigo` with `unistack` in `README_zh-CN.md`

**Checkpoint**: At this point, all CI and documentation references are updated.

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T099 Run quickstart.md validation locally to verify package name changes

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Can start immediately (Directory renaming)
- **Foundational (Phase 2)**: Skipped
- **User Stories (Phase 3+)**: Depend on Setup completion
- **Polish (Final Phase)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Setup (Phase 1)
- **User Story 2 (P2)**: Can start after Setup (Phase 1). Independent of US1.

### Parallel Opportunities

- All tasks within US1 can be run in parallel since they touch different files.
- All tasks within US2 can be run in parallel since they touch different files.
- US1 and US2 can be executed in parallel.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 3: User Story 1
3. **STOP and VALIDATE**: Verify PyPI configuration

### Incremental Delivery

1. Rename directories
2. Localize Python scripts and configs
3. Localize GitHub workflows
4. Localize Documentation
