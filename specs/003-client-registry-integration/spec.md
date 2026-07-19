# Feature Specification: 客户端对软件仓库的集成 (Client Registry Integration)

**Feature Branch**: `003-client-registry-integration`

**Created**: 2026-07-19

**Status**: Draft

**Input**: User description: "任务 2：客户端对软件仓库的集成 - 实现云端 .tar.gz 离线包下载逻辑 (unistack download) - 实现 Registry DB 的更新与同步逻辑 (unistack update) - 实现包哈希与签名校验"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 同步软件库 (unistack update) (Priority: P1)

作为 UniStack 用户，我希望能够通过简单的命令更新本地的软件源数据表，以便了解最新的可用包。

**Why this priority**: 这是整个客户端集成的基础，没有 Registry DB，就无法知道有哪些包可供下载。

**Independent Test**: Can be fully tested by running `unistack update` and verifying that the local SQLite database (`packages.db.zst` -> `packages.db`) is fetched, decompressed, and available for querying.

**Acceptance Scenarios**:

1. **Given** 没有任何本地缓存，**When** 用户执行 `unistack update`，**Then** 系统从远程仓库下载 `packages.db.zst`，解压为 `packages.db`，并提示更新成功。
2. **Given** 网络不佳或断网，**When** 用户执行 `unistack update`，**Then** 系统给出明确的重试和失败提示，本地已有的旧库保持不变。

---

### User Story 2 - 下载离线安装包 (unistack download) (Priority: P1)

作为 UniStack 用户，我希望能够指定包名称并下载对应的离线 `.tar.gz` 文件，以便于无网环境下安装。

**Why this priority**: 核心功能，让 Ansible 可以通过离线包路径进行自动化部署。

**Independent Test**: Can be fully tested by running `unistack download <package-name>` and verifying that the `.tar.gz` is present in the cache folder.

**Acceptance Scenarios**:

1. **Given** 软件库中有 `hello` 1.0.0 版本，**When** 用户执行 `unistack download hello`，**Then** 客户端下载 `hello-1.0.0.tar.gz` 到本地缓存目录，并提示下载成功。
2. **Given** 用户请求了一个不存在的包，**When** 用户执行 `unistack download invalid-pkg`，**Then** 系统查询 SQLite 后直接报错提示“未找到包”。

---

### User Story 3 - 包安全校验 (哈希与签名) (Priority: P2)

作为关注安全的用户，我希望下载下来的安装包经过严格的哈希校验或签名验证，防止遭遇中间人篡改（GPL 投毒）。

**Why this priority**: 安全底线，确保客户端安装的包和 Registry 中心发布的包完全一致。

**Independent Test**: Can be fully tested by modifying a downloaded package and verifying that the download/install process rejects it with a security error.

**Acceptance Scenarios**:

1. **Given** 一个正常的包已下载，**When** 客户端进行哈希比对，**Then** 校验通过并允许解压。
2. **Given** 包在传输过程中被篡改（哈希不匹配），**When** 客户端下载完毕后进行校验，**Then** 报错并删除被篡改的文件。

### Edge Cases

- 远程服务器 404 或由于限流导致的 429 响应时，系统是否进行重试？
- 解压 `.zst` 或 `.tar.gz` 过程中磁盘空间不足如何提示？
- 用户强行 `Ctrl+C` 中断下载，如何处理未下载完的临时文件？

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST 提供 `unistack update` 命令拉取远端 `packages.db.zst`，并使用纯 Go (klauspost/compress/zstd) 解压成 SQLite 数据库。
- **FR-002**: System MUST 允许通过配置文件或环境变量指定远程 Registry 的 URL。
- **FR-003**: System MUST 提供 `unistack download [pkgName]` 命令。
- **FR-004**: System MUST 在执行下载时，先从本地 SQLite 查询该包的下载地址、版本号以及期望哈希值。
- **FR-005**: System MUST 在下载 `.tar.gz` 完毕后，立刻通过 SHA-256（或其他哈希算法）与数据库中记录的哈希进行校验。
- **FR-006**: System MUST 保证在校验失败时，删除残留的不安全包并抛出异常。
- **FR-007**: System MUST 处理下载过程中的网络重试机制（默认至少3次）。

### Key Entities *(include if feature involves data)*

- **Registry Config**: 记录上游仓库地址、本地缓存路径的对象。
- **Package Manifest**: 本地 SQLite 缓存数据，用以检索目标包是否存在、URL 以及 SHA-256。
- **Download Request**: 抽象的下载任务，包含目标 URL、本地保存路径、重试次数、校验哈希。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 网络环境良好的情况下，`unistack update` 命令执行成功且整个流程控制在 2 秒内。
- **SC-002**: 用户能在无外部依赖的情况下（无需系统级 `tar` 或 `curl` 命令），完成 100% 纯 Go 实现的下载、校验和 Zstd 解压。
- **SC-003**: 篡改过的 `.tar.gz` 被拦截率达到 100%。

## Assumptions

- 假设远程服务器提供 `packages.db.zst` 作为全量软件目录快照。
- 假设包下载链接为公开资源，无需复杂的 OAuth 等鉴权，直接 HTTP GET 获取。
- 假设网络异常能够通过标准的回退重试机制 (Exponential Backoff) 处理。
