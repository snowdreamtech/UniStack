# Feature Specification: Local Package Installation

**Feature Branch**: `004-local-package-installation`

**Created**: 2026-07-19

**Status**: Draft

**Input**: User description: "Local Package Installation（实现解压并安装下载好的 .tar.gz 文件）"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Install a Package by Name (Priority: P1)

Users need to be able to install a package by simply providing its name, allowing them to easily acquire software without dealing with manual downloading and extraction.

**Why this priority**: Core functionality of any package manager.

**Independent Test**: Can be fully tested by running `unistack install <package_name>` and verifying the software becomes executable in the user's environment.

**Acceptance Scenarios**:

1. **Given** a valid package name in the registry, **When** the user runs `unistack install <package_name>`, **Then** the package is downloaded (if not cached), extracted to the local installation directory, and its executable is linked to the bin path.
2. **Given** a package is already installed, **When** the user runs `unistack install <package_name>`, **Then** the system notifies the user that the package is already installed and skips installation.
3. **Given** an invalid package name, **When** the user runs `unistack install <invalid_name>`, **Then** the system shows an error indicating the package was not found.

---

### User Story 2 - Offline/Local File Installation (Priority: P2)

Users may have already downloaded the `.tar.gz` package (e.g., in a secure offline environment) and want to install it directly from the local file system.

**Why this priority**: Essential for air-gapped environments or manual offline installations, which aligns with Ansible's strengths in diverse network conditions.

**Independent Test**: Can be fully tested by running `unistack install ./path/to/local-package.tar.gz` and verifying successful extraction and linking.

**Acceptance Scenarios**:

1. **Given** a valid local `.tar.gz` package, **When** the user runs `unistack install ./local-package.tar.gz`, **Then** the package is extracted to the installation directory and linked, skipping the download phase.
2. **Given** a corrupted local package file, **When** the user runs `unistack install ./corrupted.tar.gz`, **Then** the system reports an extraction error and aborts cleanly.

---

### Edge Cases

- What happens when the extraction directory runs out of disk space? -> Abort installation, clean up partial extraction, and report disk full error.
- What happens when a package has conflicting executable names with an already installed package? -> Fail safely with a conflict warning, or overwrite if forced.
- How does system handle interrupted extractions (e.g., Ctrl+C)? -> Installation state should be atomic; partial directories must be cleaned up to prevent corrupted installs.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support installing packages directly by name from the registry (via `unistack install <name>`).
- **FR-002**: System MUST support installing packages from a local `.tar.gz` file (via `unistack install <filepath>`).
- **FR-003**: System MUST extract the `.tar.gz` contents into a dedicated, versioned directory (e.g., `~/.local/share/unistack/packages/<name>-<version>`).
- **FR-004**: System MUST create a symlink (or equivalent wrapper script on Windows) from the package's executable to a central bin directory (e.g., `~/.local/bin`).
- **FR-005**: System MUST ensure atomicity of installation (a package is either fully installed and linked, or not at all).

### Key Entities

- **Installed Package**: Represents software that has been extracted and resides in the local installation directory.
- **Symlink/Wrapper**: The executable entry point placed in the user's PATH that points to the actual binary inside the installed package's directory.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can successfully install a typical CLI tool package in under 5 seconds (excluding download time).
- **SC-002**: The installation leaves no residual or temporary extraction files upon success or failure.
- **SC-003**: The installed executable is immediately accessible in the user's terminal environment (assuming the bin directory is in their PATH).

## Assumptions

- Users have write access to the target installation directories (e.g., `~/.local/share` and `~/.local/bin`).
- The package `.tar.gz` format implies a specific internal structure (e.g., a `bin/` directory or a known entry point) that the package manager knows how to link.
- Dependencies between packages are out of scope for this basic installation phase.
