# Research: 客户端对软件仓库的集成

## Decisions

1. **Decision**: HTTP Download Logic
   **Rationale**: We need robust download logic with retries and exponential backoff.
   **Alternatives considered**: Calling `curl` via shell, rejected due to zero-CGO and cross-platform principles. Native `net/http` provides all required capabilities.

2. **Decision**: Zstd Decompression on Client
   **Rationale**: We used zstd during the registry builder. The client must use the identical `github.com/klauspost/compress/zstd` to stream-decompress `packages.db.zst` directly to `packages.db`.
   **Alternatives considered**: Gzip, but zstd was chosen in the builder step for speed.

3. **Decision**: SQLite Client Query
   **Rationale**: The client reads `packages.db` using standard SQL `SELECT * FROM packages WHERE name=?`. `modernc.org/sqlite` works well here without requiring write transactions.
