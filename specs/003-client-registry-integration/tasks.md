# Tasks: 客户端对软件仓库的集成

## Phase 1: Setup

- [x] T001 Initialize internal/client package structure
- [x] T002 Update cmd root to configure cache directories (`~/.unistack/cache`)

## Phase 2: Foundational

- [x] T003 Implement `client.Downloader` utility for HTTP GET with retries and exponential backoff

## Phase 3: User Story 1 - 同步软件库 (unistack update)

- [x] T004 [US1] Create `internal/client/updater.go` to download `packages.db.zst`
- [x] T005 [US1] Implement zstd decompression stream from download to `packages.db` in `internal/client/updater.go`
- [x] T006 [US1] Register `cmd/17.update.go` Cobra command for `unistack update`
- [x] T007 [US1] Add integration test for `unistack update` against a mock local HTTP server

## Phase 4: User Story 2 & 3 - 下载离线安装包及安全校验 (unistack download)

- [x] T008 [P] [US2] Extend `internal/registry/client_query.go` to query SQLite `packages.db` for package metadata
- [x] T009 [US3] Create `internal/client/hash.go` for SHA-256 validation stream
- [x] T010 [US2] Implement `internal/client/downloader.go` `DownloadPackage` handling `.tar.gz` download with hash validation (US3)
- [x] T011 [US2] Register `cmd/18.download.go` Cobra command for `unistack download <pkg>`
- [x] T012 [US2] Add integration test for successful download
- [x] T013 [US3] Add integration test for hash mismatch rejection

## Phase 5: Polish & Cross-Cutting

- [x] T014 Handle edge cases (network 404/429, disk full, Ctrl+C cleanup)
- [x] T015 Ensure pure Go constraints are respected (zero CGO)
