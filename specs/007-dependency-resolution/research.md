# Research: Dependency Resolution Engine

## Decisions

### 1. Graph Resolution Algorithm

- **Decision**: Use Kahn's Algorithm for Topological Sorting.
- **Rationale**: Kahn's algorithm is straightforward to implement in Go using channels or standard queues/slices. It effectively identifies circular dependencies (if the sorted list has fewer nodes than the graph, there's a cycle).
- **Alternatives considered**: Depth First Search (DFS) with node coloring. While DFS is elegant, Kahn's algorithm can easily be extended for concurrent installation steps later (installing nodes with in-degree 0 in parallel).

### 2. Dependency Query Mechanism

- **Decision**: Query the local `packages.db` SQLite registry using a recursive function or a single batch query.
- **Rationale**: SQLite is fast enough that querying dependencies on-demand for small trees (<100 nodes) is sub-millisecond. We will implement `registry.GetDependencies(packageName string) ([]string, error)`.
- **Alternatives considered**: Pulling the entire DB into memory. Unnecessary and doesn't scale if the registry grows large.
