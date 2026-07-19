# Feature Specification: Package Installation & Local E2E Verification

**Feature Branch**: `009-package-install`

**Created**: 2026-07-19

**Status**: Draft

**Input**: User description: "以roles /Users/snowdream/Workspace/snowdreamtech/UniStack/ansible/roles/apps/hello 为我们的测试包，建立本地仓库，然后依据整个本地仓库，跑通整个仓库以及软件的下载安装流程。你能明白吗？使用speckit工具创建需求。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Local Registry E2E Installation (Priority: P1)

As a package consumer, I want to install a package (`hello`) from a local UniStack registry so that I can verify the end-to-end package distribution flow works seamlessly without relying on network components.

**Why this priority**: Validating the core download and installation phases using a local registry is the most controlled and resilient method for ensuring the dependency graph, registry sourcing, and local unpack logic all function correctly before rolling out to online endpoints.

**Independent Test**: Can be fully tested by generating a local registry from the `hello` package, adding it as a source, and running `unistack install hello`.

**Acceptance Scenarios**:

1. **Given** a locally built registry containing the `hello` package and added as a UniStack source, **When** I execute `unistack install hello`, **Then** the package is correctly downloaded via the `file://` scheme, extracted to `~/.local/share/unistack/packages/hello-1.0.0`, the Ansible tasks execute, and the binary is symlinked.

### User Story 2 - Source Resolution during Query (Priority: P1)

As a system process, the Installer needs to know exactly which registry source a package comes from during dependency resolution, so that it can dynamically map the package to the correct download URL.

**Why this priority**: Without this, the installer is hardcoded to a single endpoint (`http://localhost:8080`), defeating the purpose of the multi-source registry system (008).

**Independent Test**: Can be verified by executing a search/query in the registry code and asserting the returned metadata contains the `Source` property matching the configured source name.

**Acceptance Scenarios**:

1. **Given** a virtual SQLite view of multiple package registries, **When** the Installer queries for a package, **Then** the package metadata returned includes the `source` field, allowing the system to look up the correct `url`.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The registry query layer MUST extract and return the `source` field alongside package metadata.
- **FR-002**: The `unistack install` command MUST remove any hardcoded `registryURL` fallbacks.
- **FR-003**: The installer MUST dynamically map a package's `source` to the `URL` defined in `sources.json`.
- **FR-004**: The system MUST successfully parse local file URIs (`file://`) during the download phase (already partially implemented by Downloader, needs verification).
- **FR-005**: The `unistack install` operation MUST correctly execute the Ansible role (`tasks/main.yml`) and symlink binaries post-extraction.

### Key Entities

- **PackageMetadata**: Must now include `Source` string to track origin.
- **Source Config**: The `sources.json` configuration, acting as the lookup table for the final download URL.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: E2E Installation of the `hello` package from a local `file://` registry completes with zero errors.
- **SC-002**: The installed `hello` binary is successfully executed from `~/.local/share/unistack/bin/hello` and outputs the expected result.
- **SC-003**: Hardcoded `http://localhost:8080` references are fully eradicated from the installation code path.

## Assumptions

- The `hello` package is a well-formed UniStack/Ansible role package located in `/Users/snowdream/Workspace/snowdreamtech/UniStack/ansible/roles/apps/hello`.
- `unistack repo build` is capable of properly indexing this package.
- Local repository paths will be normalized correctly for SQLite attachment and file downloads.
