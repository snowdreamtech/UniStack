---
description: "Task list for Ansible Execution Hook implementation"
---

# Tasks: Ansible Execution Hook

**Input**: Design documents from `/specs/005-ansible-execution-hook/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Project initialization and basic structure

- [ ] T001 Setup Phase: Ensure nothing is blocking since this feature only modifies existing code.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

- [ ] T002 Update `internal/client/installer.go` to import `"os/exec"`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - 自动拉起 Ansible 底层装配 (Priority: P1) 🎯 MVP

**Goal**: 当包中存在 `app_loader.yml` 配置文件时，自动使用 `os/exec` 拉起本地环境的 `ansible-playbook` 命令执行该剧本。

**Independent Test**: Install a package via `unistack install hello` and observe Ansible logs streaming.

### Implementation for User Story 1

- [ ] T003 [US1] Create a helper function `runAnsiblePlaybook(pkgPath string)` in `internal/client/installer.go` that uses `os/exec` to execute `ansible-playbook -i localhost, -c local app_loader.yml -e app_source_path=<path>` and pipes Output to os.Stdout/os.Stderr.
- [ ] T004 [US1] In `internal/client/installer.go`, modify `InstallFromLocal(pkgPath string)` to detect if `app_loader.yml` exists in `finalDir`.
- [ ] T005 [US1] In `InstallFromLocal`, if `app_loader.yml` exists, call `runAnsiblePlaybook(finalDir)`. Handle errors properly.
- [ ] T006 [US1] Ensure if `app_loader.yml` does *not* exist, the installation proceeds as a pure binary (skip Ansible).
- [ ] T007 [US1] Add a unit test or adjust `internal/client/installer_test.go` to mock or skip Ansible if not present in testing environments.

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T008 Run quickstart.md manual validation steps to verify End-to-End behavior.
- [ ] T009 Refactor any overly complex Go methods.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
