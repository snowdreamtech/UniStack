# Feature Specification: Registry Builder 集成

**Feature Branch**: `002-registry-builder`

**Created**: 2026-07-18

**Status**: Draft

**Input**: User description: "Registry Builder 集成 - 把 YAML 种子包解析入库"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 构建本地注册表数据库 (Priority: P1)

作为包维护者，我希望能够使用 Go CLI 工具扫描工作区中的 `package.yml` 文件，提取它们的元数据并存入 SQLite 数据库中，以便后续通过数据库极速查询包的信息，而无需解析缓慢的 YAML 文件。

**Why this priority**: 建立 SQLite 注册表是 UniStack 大一统架构“现代包管理”体系的基石，所有极速查询（0.1 毫秒内解析依赖）都依赖于它。

**Independent Test**: 可以独立测试：给定一个包含有效 `package.yml` 的目录，运行 builder 命令后，生成包含正确元数据的 SQLite 数据库文件 (`packages.db`)。

**Acceptance Scenarios**:

1. **Given** 一个结构标准的 Ansible 角色目录（包含 `package.yml`），**When** 运行注册表构建器，**Then** 生成一个 SQLite 数据库文件，其中包含对应包的名称、版本、依赖等记录。
2. **Given** 多个不同的角色目录，**When** 运行注册表构建器，**Then** 数据库能够包含所有的记录，且无冗余或冲突。

---

### User Story 2 - 严格的语义化版本校验 (Priority: P1)

作为系统设计者，我希望在解析 `package.yml` 存入数据库之前，严格校验版本号是否符合规范，防止非法或混乱的版本号进入包生态系统。

**Why this priority**: 版本解析是依赖管理的核心。垃圾版本数据会导致依赖解析树崩溃，必须在入口处（Builder）严格拦截。

**Independent Test**: 给定包含非法版本（如 `v1.0` 而非 `1.0.0`，或 `latest`）的 `package.yml`，系统应明确报错并拒绝入库。

**Acceptance Scenarios**:

1. **Given** 一个包含 `version: "1.0.0"` 的包，**When** 构建器解析，**Then** 成功入库。
2. **Given** 一个包含 `version: "abc.1"` 的包，**When** 构建器解析，**Then** 报错并指出版本号不符合 SemVer 规范。
3. **Given** 依赖声明中包含 `">= 1.0.0"` 的包，**When** 构建器解析，**Then** 成功识别并解析为版本范围。

---

### User Story 3 - 数据库压缩输出 (Priority: P2)

作为平台构建者，我希望生成的 `packages.db` 能够自动被压缩为 `.zst` (Zstandard) 格式，以便未来的客户端能够以最小的网络带宽消耗下载注册表。

**Why this priority**: 虽然功能重要（关系到分发效率），但它建立在 P1 已经成功构建出 SQLite 文件的基础上。

**Independent Test**: 构建完成后，检查输出目录是否存在有效的 `.zst` 压缩文件，且能被标准工具解压还原为原始的 SQLite 数据库。

**Acceptance Scenarios**:

1. **Given** 成功生成的 `packages.db`，**When** 构建器完成收尾工作，**Then** 自动生成一个 `packages.db.zst` 文件。

### Edge Cases

- 如果目标目录不存在任何 `package.yml` 怎么办？系统应安全退出并提示“未发现有效包”。
- 如果两个不同的目录声明了相同的包名和版本怎么办？系统应报错提示“重复包版本冲突”。
- 如果 `package.yml` 格式损坏或缺失必填字段（如名称、版本）怎么办？解析失败并跳过该包，打印警告日志。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系统 MUST 递归扫描指定目录，查找所有名为 `package.yml` 的文件。
- **FR-002**: 系统 MUST 能够正确解析 YAML 格式并提取 `name`, `version`, `description`, `dependencies` 等核心字段。
- **FR-003**: 系统 MUST 强依赖 `github.com/Masterminds/semver/v3` 对 `version` 字段进行语义化版本校验。
- **FR-004**: 系统 MUST 创建或连接到一个 SQLite 数据库（使用纯 Go 的 `modernc.org/sqlite`）。
- **FR-005**: 系统 MUST 将合法的包数据插入到 SQLite 数据库的表中（如 `packages` 表和 `dependencies` 表）。
- **FR-006**: 系统 MUST 在数据库写入完成后，使用 Zstandard 算法对数据库文件进行压缩，生成 `.zst` 产物。

### Key Entities

- **PackageMetadata**: 表示从 YAML 中提取的包结构，包含名称、版本、作者、描述等。
- **PackageDependency**: 表示某个包对其他包的具体版本范围依赖。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 构建工具能够在 1 秒内扫描并解析超过 100 个 `package.yml` 文件。
- **SC-002**: 输出的 SQLite 数据库必须能用标准的 sqlite3 工具读取，且表结构清晰。
- **SC-003**: 成功拦截 100% 的非 SemVer 标准版本号输入，不产生脏数据。

## Assumptions

- 所有的源数据均由开发者在本地提供，目前不涉及网络远程拉取 YAML。
- 压缩算法默认采用 zstd 的均衡或默认压缩率，无需向用户提供复杂的压缩级别配置。
- `package.yml` 的结构目前已在 Phase 1 (Seed Packages) 中被标准化。
