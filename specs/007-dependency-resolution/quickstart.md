# Quickstart Validation: Dependency Resolution

## Prerequisites

- A compiled `unistack` binary with the resolver integrated.
- A local registry DB containing test packages A, B, and C.
  - A depends on B and C.
  - B depends on C.

## Setup Commands

```bash
# Insert mock data into registry
sqlite3 ~/.local/share/unistack/registry/packages.db "INSERT INTO packages (name, version) VALUES ('A', '1.0.0'), ('B', '1.0.0'), ('C', '1.0.0');"
sqlite3 ~/.local/share/unistack/registry/packages.db "INSERT INTO dependencies (package_name, dependency_name, version_constraint) VALUES ('A', 'B', '>=1.0.0'), ('A', 'C', '>=1.0.0'), ('B', 'C', '>=1.0.0');"
```

## Validation Scenario 1: Successful Resolution

```bash
unistack install A
```

**Expected Outcome**:
The system logs indicate:

1. Resolved dependency graph.
2. Installing C...
3. Installing B...
4. Installing A...

## Validation Scenario 2: Circular Dependency Detection

```bash
# Create a cycle: C depends on A
sqlite3 ~/.local/share/unistack/registry/packages.db "INSERT INTO dependencies (package_name, dependency_name, version_constraint) VALUES ('C', 'A', '>=1.0.0');"

unistack install A
```

**Expected Outcome**:
Error: `circular dependency detected in package graph`
