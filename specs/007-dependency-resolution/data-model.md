# Data Model: Dependency Resolution Engine

## Entities

### `DependencyGraph`

In-memory representation of the package dependencies.

- **Fields**:
  - `Nodes`: Map of string (package name) to `*Node`.

### `Node`

A single package in the graph.

- **Fields**:
  - `Name`: String.
  - `Dependencies`: List of string (package names this node depends on).
  - `InDegree`: Integer (number of packages that depend on this node).

## State Transitions

1. `BuildGraph(targetPkg)`: Recursively fetches dependencies from SQLite and populates the `DependencyGraph`.
2. `TopologicalSort()`: Iteratively finds nodes with `InDegree == 0`, adds them to the resolved list, and decrements `InDegree` for their dependencies. Returns the ordered slice of package names. If resolved slice length != total nodes, throws `ErrCircularDependency`.
