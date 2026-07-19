# Feature Specification: Registry Sources CRUD

**Feature Branch**: `008-registry-sources`

**Created**: 2026-07-19

**Status**: Draft

**Input**: User description: "一个镜像源，私有源，公有源的CRUD，新增，更新，删除，查询，列表，你明白吗。没有更新，没有删除吗？unistack source (add/remove/list)。网站改域名，网站不维护了，这很正常。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Add and List Package Sources (Priority: P1)

As a user, I want to add new package sources (such as private mirrors or internal registries) and list them, so that I can download packages from multiple origins.

**Why this priority**: Core capability needed to fetch packages from non-default locations.

**Independent Test**: Can be fully tested by adding a new source url and listing the configured sources to verify it appears correctly.

**Acceptance Scenarios**:

1. **Given** a default setup, **When** the user runs the command to add a new source named "private" with URL "http://private.repo", **Then** the source is successfully saved.
2. **Given** multiple configured sources, **When** the user runs the list command, **Then** a clear list of all sources (name and URL) is presented.

---

### User Story 2 - Update Existing Sources (Priority: P2)

As a user, I want to update the URL of an existing source if the domain changes, so that I do not have to remove and re-add it.

**Why this priority**: Domains frequently change (as stated by the user), and updating in place provides a better user experience than deleting and recreating.

**Independent Test**: Can be fully tested by modifying an existing source's URL and verifying it is updated in the configuration.

**Acceptance Scenarios**:

1. **Given** a source named "mirror" with URL "http://old.com", **When** the user runs the update command for "mirror" with "http://new.com", **Then** the source's URL is updated successfully.
2. **Given** a non-existent source, **When** the user attempts to update it, **Then** a clear error message is shown indicating the source does not exist.

---

### User Story 3 - Remove Deprecated Sources (Priority: P2)

As a user, I want to remove sources that are no longer maintained, so that I do not waste time attempting to fetch from dead links.

**Why this priority**: Essential for maintaining a clean and functional configuration over time.

**Independent Test**: Can be tested by deleting a configured source and confirming it no longer appears in the source list.

**Acceptance Scenarios**:

1. **Given** an existing source named "obsolete", **When** the user runs the remove command, **Then** the source is removed from the configuration.
2. **Given** a source that was removed, **When** a registry update is triggered, **Then** the system does not attempt to contact the removed source's URL.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow users to add new package sources with a unique name and URL.
- **FR-002**: System MUST allow users to update the URL of an existing package source by providing its name.
- **FR-003**: System MUST allow users to remove a package source by its name.
- **FR-004**: System MUST allow users to list all currently configured package sources.
- **FR-005**: System MUST prevent adding a source if a source with the same name already exists.
- **FR-006**: System MUST persist the source configurations locally across terminal sessions.
- **FR-007**: System MUST provide a default source out-of-the-box if no custom sources have been configured.

### Key Entities

- **Source Configuration**: Represents a package registry endpoint containing a unique name (identifier) and a URL.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can successfully manage multiple package sources purely via the CLI without manually editing configuration files.
- **SC-002**: Updating or removing a source takes less than 1 second.
- **SC-003**: 100% of the active configured sources are displayed correctly when the list command is executed.
- **SC-004**: Adding a duplicate source name correctly fails with an informative error message in 100% of cases.

## Assumptions

- Users have write permissions to their local configuration directory.
- The URLs provided by the user are valid endpoints.
- The default registry provided out-of-the-box is "https://registry.unistack.org".
