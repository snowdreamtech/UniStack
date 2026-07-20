# Tasks: refactor-loader

**Input**: Design documents from `/specs/011-refactor-loader/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

## Phase 1: Setup

**Purpose**: Project initialization and basic structure

- [ ] T001 Verify Ansible version and core dependencies

---

## Phase 2: User Story 1 - Standardized Package Loading Context (Priority: P1)

**Goal**: Update the `app` loader to execute packages using standard Ansible role execution contexts (`include_role`), solving file/template path resolution issues.

**Independent Test**: Can be tested by creating a dummy role with templates and running it via the loader.

### Implementation for User Story 1

- [ ] T002 [US1] Update `ansible/roles/app/tasks/app_loader.yml` to replace `include_tasks` with `include_role` (using `tasks_from`)
- [ ] T003 [US1] Update `ansible/roles/native/tasks/native_loader.yml` to replace `include_tasks` with `include_role`

**Checkpoint**: At this point, the app loader correctly uses `include_role` for execution.

---

## Phase 3: User Story 2 - Modular Foundation Sub-packages (Priority: P1)

**Goal**: Split the monolithic `foundation` role into distinct, independently usable sub-packages.

**Independent Test**: Can be verified by executing a sub-package (e.g. `openssh`) independently using the updated `app` loader.

### Implementation for User Story 2

- [ ] T004 [P] [US2] Extract `openssh` tasks, handlers, templates, and defaults from `foundation` into `ansible/roles/apps/openssh`
- [ ] T005 [P] [US2] Extract `sudo` tasks, templates, and defaults from `foundation` into `ansible/roles/apps/sudo`
- [ ] T006 [P] [US2] Extract `repositories` tasks, templates, and defaults from `foundation` into `ansible/roles/apps/repositories`
- [ ] T007 [P] [US2] Extract `mirror` tasks, templates, and defaults from `foundation` into `ansible/roles/apps/mirror`
- [ ] T008 [P] [US2] Extract `user` tasks, templates, and defaults from `foundation` into `ansible/roles/apps/user`

**Checkpoint**: All 5 new modular packages exist and are structurally correct.

---

## Phase 4: User Story 3 - Foundation Metapackage Orchestration (Priority: P2)

**Goal**: Refactor the `foundation` package to act as a metapackage that invokes its sub-packages.

**Independent Test**: Can be tested by running the `scenarios/foundation.yml` playbook.

### Implementation for User Story 3

- [ ] T009 [US3] Refactor `ansible/roles/apps/foundation/tasks/main.yml` to sequentially execute `mirror`, `repositories`, `sudo`, `user`, and `openssh` using `include_role`
- [ ] T010 [US3] Clean up obsolete files (tasks, handlers, templates, vars) from `ansible/roles/apps/foundation/` since they are now delegated

**Checkpoint**: All user stories are independently functional. The foundation baseline works identically to before.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Validation and cleanup

- [ ] T011 Run `quickstart.md` Scenario 1 validation in Docker
- [ ] T012 Run `quickstart.md` Scenario 2 validation in Docker

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **User Story 1 (Phase 2)**: Depends on Setup.
- **User Story 2 (Phase 3)**: Depends on User Story 1 (Loader must be fixed before we can effectively test the sub-packages individually).
- **User Story 3 (Phase 4)**: Depends on User Story 2 (Sub-packages must exist before metapackage can call them).
- **Polish (Phase 5)**: Depends on all user stories.

### Parallel Opportunities

- Within User Story 2 (Phase 3), the extraction of the 5 roles (T004-T008) can all be done in parallel because they touch completely different files.
