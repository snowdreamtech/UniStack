# Feature Specification: Package Namespace Support

**Feature Branch**: `010-package-namespaces`

**Created**: 2026-07-20

**Status**: Draft

**Input**: User description: "Add namespace prefix support to community packages to prevent naming collisions"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Publish a namespaced package (Priority: P1)

As a package author publishing to the community repository, I want to prefix my package name with my namespace (e.g., `snowdreamtech/hello`) so that it does not collide with official core packages or packages from other authors.

**Why this priority**: Without namespaces, a single global namespace in a community repository will immediately lead to name squatting and collisions.

**Independent Test**: Can be fully tested by creating a `package.yml` with `name: snowdreamtech/hello`, running `registry build`, and verifying it is successfully archived without creating unintended deep nested directories.

**Acceptance Scenarios**:

1. **Given** a package with name `snowdreamtech/hello`, **When** `registry build` is executed, **Then** the package is compressed to a flat filename (e.g., `snowdreamtech_hello-1.0.0.tar.gz`) instead of nested directories (`s/snowdreamtech/hello...`).

---

### User Story 2 - Install and Upgrade a namespaced package (Priority: P1)

As a user, I want to install and upgrade packages that have a namespace prefix (e.g. `unistack install snowdreamtech/hello`), and have the system correctly track its installation status.

**Why this priority**: If the local package extraction folder contains slashes, the local state scanner (`ListInstalledPackages`) fails to detect it because it only scans one level deep. This breaks idempotency and upgrades.

**Independent Test**: Can be fully tested by installing a namespaced package twice. The second time should skip installation. Upgrading should successfully remove the old version.

**Acceptance Scenarios**:

1. **Given** a namespaced package is installed, **When** the system lists installed packages, **Then** it correctly detects the package and its version.
2. **Given** a namespaced package is installed, **When** the user installs it again, **Then** the system skips it (idempotent).
3. **Given** an updated namespaced package is available, **When** the user upgrades it, **Then** the old version is successfully uninstalled and the new one is installed.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support `/` in package names (e.g., `namespace/pkgname`) at the schema and CLI level.
- **FR-002**: System MUST serialize namespaced packages into a flat URL-safe string (e.g. converting `/` to `_` or `-`) when writing tarballs in `registry build`.
- **FR-003**: System MUST serialize namespaced packages into a flat directory name when extracting packages in `~/.local/share/unistack/packages/` during installation.
- **FR-004**: System MUST successfully resolve, install, upgrade, and list packages that contain namespaces.

### Key Entities

- **PackageName**: A string that can now safely contain a namespace prefix (e.g., `author/package`).
- **PackageID/SafeName**: The URL-safe and File-System-safe physical representation of the package name (e.g., `author_package`).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of operations (install, upgrade, remove, list) work correctly for packages with namespaces.
- **SC-002**: The `packages/` directory inside a built registry has a max depth of 2 (i.e. `packages/<initial>/<safename-version>.tar.gz`), regardless of how many slashes are in the package name.
- **SC-003**: The `~/.local/share/unistack/packages/` directory has a max depth of 1 for package folders.

## Assumptions

- We assume a single `/` character is the standard delimiter for namespaces, similar to Docker images (`user/image`) or GitHub repos (`user/repo`).
- We assume replacing `/` with `_` is safe and will not cause secondary collisions (e.g., assuming users won't intentionally create `foo_bar` and `foo/bar` to cause conflicts).
