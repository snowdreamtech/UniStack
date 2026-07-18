# Feature Specification: Seed Packages

**Feature Branch**: `001-seed-packages`

**Created**: 2026-07-18

**Status**: Draft

**Input**: User description: "Phase 1: 种子包准备 (Seed Packages) - 对应你的第四点
没有真实的数据，就写不出强壮的代码。这是目前最紧迫的第一步。

 1.1 在 ansible/roles/apps/ 下创建 template 目录，作为标准的展示型模版，包含规范的 package.yml、tasks/main.yml、defaults/main.yml 和 vars/。
 1.2 在 ansible/roles/apps/ 下创建 hello 目录，参考 Repology 编写一个真实可用的 GNU Hello 打包配置，作为我们后续所有构建和下载的测试用例。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create Seed Package Templates (Priority: P1)

Developers and packaging authors need a standard reference template when contributing new software packages to UniStack.

**Why this priority**: Without a reference template, the structure of package definitions will fragment, breaking automated registry parsing.

**Independent Test**: Can be tested by manually verifying that `ansible/roles/apps/template` contains a valid `package.yml` and essential Ansible folders.

**Acceptance Scenarios**:

1. **Given** a new developer wants to package software, **When** they look into `ansible/roles/apps/template`, **Then** they see a fully commented, valid `package.yml` and corresponding tasks/defaults structure.

---

### User Story 2 - Implement the GNU Hello Package (Priority: P1)

System maintainers need a real, functioning software package to test the entire lifecycle (build, sync, install) of the UniStack registry.

**Why this priority**: GNU Hello is the universal standard for "hello world" packages. Having a real package enables testing the Go registry builder and client.

**Independent Test**: Can be tested by verifying the `hello` package can be parsed and hypothetically installed via Ansible.

**Acceptance Scenarios**:

1. **Given** the UniStack repository, **When** developers examine `ansible/roles/apps/hello`, **Then** they find a realistic configuration using Repology data for GNU Hello.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a `template` directory under `ansible/roles/apps/` containing a fully standard-compliant `package.yml`, `tasks/main.yml`, `defaults/main.yml`, and `vars/`.
- **FR-002**: System MUST provide a `hello` directory under `ansible/roles/apps/` containing a real-world configuration for GNU Hello.
- **FR-003**: The `hello` package MUST include realistic `appVersion` data referencing actual versions tracked on Repology (e.g., versions for Debian, Alpine).
- **FR-004**: The `package.yml` in both packages MUST be strictly valid YAML and strictly conform to Semantic Versioning 2.0.0 for `version` and `appVersion`.

### Key Entities

- **Package Template (`template`)**: A structural blueprint containing boilerplate metadata and directory layouts.
- **Test Package (`hello`)**: A concrete instance of a package used as a live test fixture for registry building.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `ansible/roles/apps/template/package.yml` successfully passes YAML linting.
- **SC-002**: `ansible/roles/apps/hello/package.yml` successfully passes YAML linting and accurately models GNU Hello metadata.
- **SC-003**: Both directories contain the minimal required subdirectories (`tasks`, `defaults`) to be recognized as valid Ansible roles.

## Assumptions

- We assume GNU Hello is representative enough of a standard package for initial testing.
- We assume `Masterminds/semver/v3` will be used to parse the versions defined in these seed packages.
