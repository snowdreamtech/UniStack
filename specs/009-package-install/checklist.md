# Specification Checklist: Package Installation & E2E Local Verification

**Purpose**: Verifies that the implementation for the package install flow meets all requirements from the specification.
**Created**: 2026-07-19
**Feature**: [spec.md](file:///Users/snowdream/Workspace/snowdreamtech/UniStack/specs/009-package-install/spec.md)

## Source Resolution

- [ ] CHK001 The `PackageMetadata` struct in `client_query.go` includes a `Source` string field.
- [ ] CHK002 `QueryPackage` SQL retrieves the `source` column and correctly maps it to the struct.
- [ ] CHK003 Verify that a queried package correctly identifies whether it comes from `core`, `community`, or a local `file://` registry.

## Installer Dynamic Mapping

- [ ] CHK004 `registryURL` parameter is removed from `InstallPackage()` signature.
- [ ] CHK005 `InstallPackage()` uses `config.LoadSources()` to look up the exact URL using `meta.Source`.
- [ ] CHK006 `unistack install` passes `ctx` and `target` to `InstallPackage()` without falling back to localhost.

## End-to-End Download & Execution

- [ ] CHK007 `Downloader.DownloadPackage` successfully parses local `file://` URLs and uses `os.Open` instead of HTTP GET.
- [ ] CHK008 Tarball extraction deposits files into the correct `~/.local/share/unistack/packages/[pkg]` directory.
- [ ] CHK009 If a `tasks/main.yml` Ansible role is present, it is successfully executed via local playbook wrapper.
- [ ] CHK010 Target binary (e.g. `hello`) is symlinked into `~/.local/share/unistack/bin/`.

## Validation

- [ ] CHK011 Complete the full lifecycle test strictly against `/Users/snowdream/Workspace/snowdreamtech/UniStack/ansible/roles/apps/hello`.
