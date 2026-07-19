# Dependency Resolution Engine (依赖关系解析引擎)

## Feature Description

Implement a dependency graph resolution algorithm in the Go client. When installing a package A, if it depends on packages B and C, the system should automatically prioritize downloading and installing B and C first. This leverages the `dependencies` table previously designed in the Registry DB.

## User Scenarios & Testing

1. **Given** a user wants to install package A, **When** they run `unistack install A`, and A depends on B and C, **Then** the client automatically resolves, downloads, and installs B and C before installing A.
2. **Given** a user wants to install package D, **When** they run `unistack install D`, and D has a circular dependency (e.g., D depends on E, which depends on D), **Then** the client detects the circular dependency and gracefully aborts with an error message.
3. **Given** a user wants to uninstall package B, **When** they run `unistack uninstall B`, and package A is installed and depends on B, **Then** the client warns the user or prevents the uninstallation unless forced.

## Functional Requirements

- **FR-001**: The client MUST resolve dependencies by querying the local SQLite Registry DB (`dependencies` table).
- **FR-002**: The client MUST construct a Directed Acyclic Graph (DAG) representing the installation order.
- **FR-003**: The client MUST implement circular dependency detection during DAG construction.
- **FR-004**: The client MUST install dependencies recursively in topological order before installing the target package.
- **FR-005**: The client MUST verify dependency fulfillment; if a dependency is already installed and meets version constraints, it is skipped.

## Success Criteria

- 100% of multi-level dependency installations succeed without manual intervention.
- The system correctly halts on circular dependencies within 1 second.

## Key Entities

- `PackageMetadata`: Including dependency constraints.
- `DependencyGraph`: In-memory DAG representation.

## Assumptions

- The `dependencies` table in the local SQLite Registry is accurate and updated via `unistack update`.
- The user has sufficient disk space to download all dependencies.
- Network stability is assumed for sequential or concurrent downloads.
