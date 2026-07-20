# Feature Specification: refactor-loader

**Feature Branch**: `011-refactor-loader`

**Created**: 2026-07-20

**Status**: Draft

**Input**: User description: "1. Ensure app can fully load a software package, including its tasks, vars, files, templates, defaults, without hacky ways. 2. Split foundation. Extract sub-packages from foundation, and make foundation a metapackage."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Standardized Package Loading Context (Priority: P1)

As a system orchestrator, I want the `app` loader to execute packages using standard Ansible role execution contexts (e.g., `include_role`), so that all files, templates, and defaults within the package are resolved accurately without manual path concatenation or hacky workarounds.

**Why this priority**: Correctly loading templates and files relative to the target package is fundamentally required for the Ansible orchestrator to work reliably across different operating systems.

**Independent Test**: Can be fully tested by creating a dummy package with a relative template file and ensuring it renders correctly when called via the `app` loader.

**Acceptance Scenarios**:

1. **Given** a package with a `templates/config.j2` file, **When** the `app` loader executes this package, **Then** the template is successfully resolved and rendered without throwing a "file not found" error.

---

### User Story 2 - Modular Foundation Sub-packages (Priority: P1)

As a system administrator, I want the monolithic `foundation` package to be split into distinct, independently usable sub-packages (e.g., `openssh`, `sudo`, `repositories`, `mirror`, `user`), so that I can deploy them individually if needed.

**Why this priority**: Monolithic packages limit reusability and complicate maintenance. Modular packages allow granular control over system provisioning.

**Independent Test**: Can be fully tested by attempting to deploy just the `openssh` or `user` sub-package independently of the `foundation` package.

**Acceptance Scenarios**:

1. **Given** a target system, **When** a user applies the standalone `openssh` package via the UniStack `app` loader, **Then** the SSH service is correctly configured and restarted.

---

### User Story 3 - Foundation Metapackage Orchestration (Priority: P2)

As a system administrator, I want the `foundation` package to act as a metapackage, so that I can apply the complete foundational baseline by running a single package which automatically pulls in its constituent sub-packages.

**Why this priority**: Preserves the existing user experience of deploying a unified baseline while benefiting from modularity under the hood.

**Independent Test**: Can be fully tested by running the `foundation.yml` scenario playbook and verifying all 5 sub-packages are executed.

**Acceptance Scenarios**:

1. **Given** a target system, **When** a user applies the `foundation` metapackage, **Then** all underlying sub-packages (repositories, mirror, user, sudo, openssh) are sequentially deployed.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `app_loader.yml` and `native_loader.yml` MUST use `ansible.builtin.include_role` with the `tasks_from` property to execute resolved task files, rather than `ansible.builtin.include_tasks`.
- **FR-002**: The system MUST automatically resolve variables, defaults, files, and templates relative to the target package's root directory.
- **FR-003**: The `foundation` package MUST be split into independent Ansible roles under `ansible/roles/apps/`: `openssh`, `sudo`, `repositories`, `mirror`, and `user`.
- **FR-004**: The `foundation` package MUST act as a metapackage that orchestrates the execution of its sub-packages in the correct dependency order.

### Key Entities

- **Metapackage**: A package (role) that contains no direct implementation logic, but instead lists dependencies or directly calls other modular packages.
- **Sub-package**: A standalone, domain-specific package (e.g., `openssh`) that contains its own defaults, vars, tasks, handlers, and templates.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of dynamically loaded task executions resolve their templates, vars, and files from the target package's directory without manual path manipulation errors.
- **SC-002**: The monolithic `apps/foundation` role is successfully refactored into at least 5 distinct sub-package roles.
- **SC-003**: The end-to-end `scenarios/foundation.yml` playbook executes successfully across target environments, producing identical outcomes to the monolithic version.

## Assumptions

- Ansible version >= 2.14 is used, which fully supports dynamic `tasks_from` resolution in `include_role`.
- Sub-packages are designed to be executed sequentially and are largely stateless or idempotent.
- The `app_name` provided to the loader matches the directory name of the target package.
