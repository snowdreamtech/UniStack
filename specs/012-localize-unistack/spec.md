# Feature Specification: Localize UniGo to UniStack

**Feature Branch**: `012-localize-unistack`

**Created**: 2026-07-24

**Status**: Draft

**Input**: User description: "审查一下刚才合并进来的代码，有哪些需要本地化修改的。我举一个例子，pypi，需要更改unigo为unstack，其他还有那一些？ /speckit.specify 制定计划，"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Publish PyPI Package for UniStack (Priority: P1)

As a maintainer, I want the project to publish the Python package to PyPI under the name "snowdreamtech-unistack" so that users can install the UniStack CLI via pip correctly, rather than installing the upstream "snowdreamtech-unigo" package.

**Why this priority**: Correct package naming is critical for distribution. Releasing under the wrong name will cause conflicts and confuse users.

**Independent Test**: Can be fully tested by verifying the `setup.py` and `pyproject.toml` configurations, and ensuring that `build.sh` produces wheel files with the correct `unistack` naming convention.

**Acceptance Scenarios**:

1. **Given** the PyPI configuration files, **When** they are built, **Then** the generated package name is `snowdreamtech-unistack` and it creates `unistack` binaries.
2. **Given** the documentation, **When** a user reads it, **Then** the installation command instructs them to use `pip install snowdreamtech-unistack`.

---

### User Story 2 - Maintain NPM Package Integrity (Priority: P1)

As a user reading the documentation, I want to see the correct NPM installation instructions for UniStack so I can install the software globally without ambiguity.

**Why this priority**: Documentation must be accurate to ensure a smooth user onboarding experience.

**Independent Test**: Read the `README.md` files and ensure `npm install -g @snowdreamtech/unistack` is referenced instead of `unigo`.

**Acceptance Scenarios**:

1. **Given** the `README.md` and `README_zh-CN.md`, **When** I check the installation instructions, **Then** I see the command `npm install -g @snowdreamtech/unistack`.

---

### User Story 3 - Validate GoReleaser Workflows (Priority: P2)

As a developer releasing a new version, I want the CI/CD pipeline to properly warm up the Go Proxy using the `unistack` repository so that the release process doesn't fail due to caching issues on `proxy.golang.org`.

**Why this priority**: Ensures reliable automated releases via GitHub Actions.

**Independent Test**: Check the GitHub Actions workflow files to ensure the `curl` command points to the `unistack` repository instead of `unigo`.

**Acceptance Scenarios**:

1. **Given** the GoReleaser GitHub Action workflow, **When** a new tag is pushed, **Then** the proxy warmup step pings `https://proxy.golang.org/github.com/snowdreamtech/unistack/...`.

### Edge Cases

- What happens when a user previously installed the `unigo` package? The new package will be completely independent; `unigo` will not be automatically uninstalled, but running `unistack` will work.
- How does the system handle older tags being pushed? The release workflow applies to any new tag, and it should correctly reference `unistack`.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST rename the Python package structure from `snowdreamtech_unigo` to `snowdreamtech_unistack`.
- **FR-002**: System MUST configure Python package metadata to use the name `snowdreamtech-unistack` and update descriptions accordingly.
- **FR-003**: System MUST update PyPI build scripts to generate and package binaries named `unistack` (e.g., `unistack.exe`, `unistack_linux_amd64`) instead of `unigo`.
- **FR-004**: System MUST update the CLI entry point mapping in Python to map `unistack` to `snowdreamtech_unistack.cli:main`.
- **FR-005**: System MUST update the README documentation (both English and Chinese) to use the correct installation commands for both NPM and PyPI.
- **FR-006**: System MUST update the `goreleaser.yml` workflow to use the correct URL for the Go proxy warmup ping (`github.com/snowdreamtech/unistack`).

### Key Entities 

- **PyPI Package Configuration**: The metadata that defines how the Python wrapper is distributed.
- **CI/CD Workflow**: The GitHub Actions setup that builds and releases the software.
- **Documentation**: The `README` files that guide end users on installation.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of references to `unigo` introduced in the latest upstream merge are replaced with `unistack` in the Python, Docs, and Workflow configurations.
- **SC-002**: The `pypi` directory structure correctly reflects the new package name, enabling a successful local build without errors.
- **SC-003**: Automated checks and Linters pass successfully after renaming.

## Assumptions

- No upstream code behavior is fundamentally broken by this renaming.
- The `unigo` references in the project's historical `CHANGELOG.md` do not need to be updated.
- The PyPI publishing credentials (tokens) are already configured correctly in the repository for `unistack`.
