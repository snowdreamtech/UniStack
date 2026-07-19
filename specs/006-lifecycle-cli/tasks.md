# Implementation Tasks: Lifecycle CLI (Uninstall, Upgrade, List)

**Feature**: [006-lifecycle-cli]
**Status**: Ready

## Phase 1: Setup

- [x] T001 Initialize tasks definition in specs/006-lifecycle-cli/tasks.md

## Phase 2: User Story 1 (List Command)

- [x] T002 [US1] Implement ListInstalledPackages in internal/client/installer.go
- [x] T003 [US1] Add unistack list command in cmd/14.list.go

## Phase 3: User Story 2 (Uninstall Command)

- [x] T004 [US2] Implement Uninstall(pkgName) in internal/client/installer.go
- [x] T005 [US2] Add unistack uninstall command in cmd/15.uninstall.go

## Phase 4: User Story 3 (Upgrade Command)

- [x] T006 [US3] Implement registry.QueryPackage logic integration in cmd/16.upgrade.go
- [x] T007 [US3] Add unistack upgrade command in cmd/16.upgrade.go

## Phase 5: Polish & Cross-Cutting Concerns

- [x] T008 Verify test pass with go test ./... in root
