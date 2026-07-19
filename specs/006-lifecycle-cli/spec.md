# Feature Specification: Lifecycle CLI (Uninstall, Upgrade, List)

**Feature Branch**: `[006-lifecycle-cli]`

**Created**: 2026-07-19

**Status**: Draft

**Input**: User description: "补齐生命周期命令 (Lifecycle CLI: Uninstall / Upgrade / List)..."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - List Installed Packages (Priority: P1)

Users need to see what packages they currently have installed on their system and what versions those packages are on.

**Why this priority**: Visibility is fundamental to any package manager. Users must know their system state before they can confidently manage it.

**Independent Test**: Can be independently tested by installing several packages and running `unistack list` to verify it correctly outputs the names and versions of installed software.

**Acceptance Scenarios**:

1. **Given** a system with installed packages, **When** the user runs `unistack list`, **Then** the CLI displays a list of installed packages and their versions, sourced from `~/.local/share/unistack`.
2. **Given** a system with no packages installed, **When** the user runs `unistack list`, **Then** the CLI outputs a message indicating no packages are installed.

---

### User Story 2 - Uninstall Packages (Priority: P1)

Users need to remove software they no longer need to free up resources and avoid clutter, including removing symlinks, cached data, and invoking any native uninstallation logic.

**Why this priority**: Providing a clean removal path is critical to maintaining a healthy host system, reversing what the `install` command did.

**Independent Test**: Can be tested by installing a complex package (e.g., using Ansible), then running `unistack uninstall <pkg>`, and verifying that the symlinks, cache, and system footprints are completely removed.

**Acceptance Scenarios**:

1. **Given** an installed package (pure binary), **When** the user runs `unistack uninstall <pkg>`, **Then** the symlink is removed, and the package directory is deleted.
2. **Given** an installed package with Ansible lifecycle hooks, **When** the user runs `unistack uninstall <pkg>`, **Then** the CLI invokes the corresponding Ansible uninstall logic (passing `-e state=absent` to `tasks/main.yml`), and only upon success does it remove the package files.
3. **Given** a package that is not installed, **When** the user runs `unistack uninstall <pkg>`, **Then** the CLI returns an appropriate error message indicating the package is not found.

---

### User Story 3 - Upgrade Packages (Priority: P2)

Users want to keep their software up to date effortlessly without having to manually uninstall the old version and install the new version.

**Why this priority**: Crucial for security and feature parity, though slightly lower priority than basic addition/removal.

**Independent Test**: Can be tested by intentionally installing an older version of a package, updating the registry, and running `unistack upgrade <pkg>` to verify it transitions cleanly to the latest version.

**Acceptance Scenarios**:

1. **Given** a package installed with an older version and a newer version available in the SQLite registry, **When** the user runs `unistack upgrade <pkg>`, **Then** the system automatically uninstalls the old version (or upgrades in place) and installs the latest version.
2. **Given** an already up-to-date package, **When** the user runs `unistack upgrade <pkg>`, **Then** the system informs the user that the package is already at the latest version and exits successfully.

### Edge Cases

- What happens when a user attempts to uninstall a package while its background service is running? (The Ansible uninstall playbook should handle service stopping).
- How does the system handle upgrading a package if the network connection drops during the download of the newer version?
- What happens if the SQLite database is not updated (`unistack update` hasn't been run) before an upgrade?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a `list` command that enumerates all currently installed packages by inspecting the local directory `~/.local/share/unistack` or local state file.
- **FR-002**: System MUST provide an `uninstall` command that accepts a package name.
- **FR-003**: System MUST remove the symbolic link(s) created during the installation of the specified package during `uninstall`.
- **FR-004**: System MUST invoke the Ansible uninstall logic (by passing `-e state=absent` to the temporary playbook running the role) if the package contains a `tasks/main.yml` during `uninstall`.
- **FR-005**: System MUST delete the local package directory and related cached files upon successful `uninstall`.
- **FR-006**: System MUST provide an `upgrade` command that accepts a package name.
- **FR-007**: System MUST compare the currently installed version against the latest version available in the local SQLite database during `upgrade`.
- **FR-008**: System MUST perform the update operation if a higher version is detected in the registry.

### Key Entities

- **Installed Package State**: Represents what is currently on the disk, consisting of the package directory, version metadata, and active symlinks.
- **Registry Database (SQLite)**: The local cache of available packages and their versions to compare against for upgrades.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `list` command accurately reflects 100% of the successfully installed packages on the system.
- **SC-002**: `uninstall` command completely removes all traces (symlinks, extracted directories) of a pure Go package without leaving orphaned files.
- **SC-003**: `uninstall` command successfully triggers the Ansible reverse logic (state: absent) for complex packages.
- **SC-004**: `upgrade` safely replaces an older package with a newer package without corrupting the local registry or user configurations.

## Assumptions

- Users have sufficient filesystem permissions to delete packages and symlinks in their user directory (or system directory if running as root).
- The Ansible role structure supports an uninstallation path (e.g., standard Ansible practices like setting `state=absent`).
- The local SQLite registry accurately reflects the latest package metadata following a `unistack update`.
