# Research: Package Namespace Flat-mapping

**Decision**: We will replace `/` with `_` when mapping package names to filesystem paths.

**Rationale**:
- `filepath.Join` converts `/` into physical directory separators, causing `packages/s/snowdreamtech/hello-1.0.0.tar.gz`.
- Local installation into `~/.local/share/unistack/packages/` uses `ListInstalledPackages` which only loops exactly one directory deep. If we use `snowdreamtech/hello`, the system fails to detect it.
- Flattening names (e.g. `snowdreamtech_hello`) makes it 100% compatible with our existing shallow-scanning algorithms, maintaining O(N) linear scan performance for idempotency checks.

**Alternatives considered**:
- Updating `ListInstalledPackages` to use recursive scanning (`filepath.WalkDir`). Rejected because walking arbitrary directories is slower and more prone to errors than flat directory scanning.
