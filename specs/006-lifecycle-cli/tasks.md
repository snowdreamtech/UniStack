# Implementation Tasks: Lifecycle CLI (Uninstall, Upgrade, List)

**Feature**: [006-lifecycle-cli]
**Status**: Ready

## Phase 1: List Command

- [x] T001 Implement `ListInstalledPackages` in `internal/client/installer.go` (or `client.go`) that scans `~/.local/share/unistack/packages` and extracts names and versions.
- [x] T002 Add `cmd/unistack/14.list.go` to provide the `unistack list` command, invoking the client logic and printing a table.

## Phase 2: Uninstall Command

- [x] T003 Implement `Uninstall(pkgName)` in `internal/client/installer.go`. It should locate the active version of the package.
- [x] T004 In `Uninstall`, check for `app_loader.yml`. If present, use `os/exec` to run `ansible-playbook -i localhost, -c local app_loader.yml -e app_source_path=<path> -e state=absent`.
- [x] T005 In `Uninstall`, remove the binary symlink from `~/.local/share/unistack/bin`.
- [x] T006 In `Uninstall`, delete the package directory from `~/.local/share/unistack/packages`.
- [x] T007 Add `cmd/unistack/15.uninstall.go` to expose the `uninstall` command.

## Phase 3: Upgrade Command

- [x] T008 Implement `GetHighestVersion(pkgName)` querying the SQLite database in `internal/repository`. (used registry.QueryPackage)
- [x] T009 Implement `Upgrade(pkgName)` in `internal/client/installer.go`. It should compare the current version with the highest version. (implemented in 16.upgrade.go)
- [x] T010 If `highest > current`, invoke `Uninstall(pkgName)` then `Install(pkgName, highest_version)`.
- [x] T011 Add `cmd/unistack/16.upgrade.go` to expose the `upgrade` command.

## Verification

- [x] T012 Run End-to-End manual testing of `list`, `uninstall`, and `upgrade`.
- [x] T013 Verify tests pass (`go test ./...`).
