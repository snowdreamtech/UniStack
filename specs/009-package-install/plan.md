# Implementation Plan: Package Installation & Dynamic Download Sourcing

**Feature**: [spec.md](file:///Users/snowdream/Workspace/snowdreamtech/UniStack/specs/009-package-install/spec.md)
**Status**: Ready for implementation

## Technical Context

- **Environment**: macOS/Linux/Windows Go runtime
- **Dependencies**: `modernc.org/sqlite`
- **Key Constraints**: Ensure all legacy `localhost:8080` references are removed. E2E test must pass strictly on local `file://` repository without any network fetches.

## Constitution Check

- **Idempotency**: Downloader must safely overwrite `.tmp` files. Installation should error gracefully or skip if already installed.
- **Portability**: Path normalizations for `file://` must work securely across Unix/Windows (handled by `Downloader.Download`).

## Implementation Phases

### Phase 1: Query Enhancement & Data Model

- Update `PackageMetadata` in `internal/registry/client_query.go` to include `Source string`.
- Update the SQLite query string in `QueryPackage` to select the `source` column.

### Phase 2: Installer Transformation

- Modify `InstallPackage` signature in `internal/client/installer.go` to drop `registryURL`.
- Within `InstallPackage`, use `config.LoadSources()` to map the `meta.Source` string back to the actual URL of the source.
- Update `cmd/19.install.go` to match the new signature of `InstallPackage`.

### Phase 3: Validation (E2E)

- Build a local registry from `/Users/snowdream/Workspace/snowdreamtech/UniStack/ansible/roles/apps/hello`.
- Add it via `unistack source add local ./local_repo`.
- Run `unistack install hello` and verify successful symlink generation and Ansible playbook execution.
