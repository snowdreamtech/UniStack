# 🔬 UniStack 软件包与注册表设计：深度研究报告 & 最终方案

> **文档说明**：本文件保留了最初对全球主流包管理系统的**深度研究与备选方案（前半部分）**，以及经过多轮讨论后**最终敲定的实施方案（后半部分 v3 版）**。两者并存以便未来随时回顾架构设计的背景与决策逻辑。

---

## 📖 第一部分：前置研究与备选方案探索

> **目标**：在 **通用性、安全性、扩展性、速度** 四个维度寻找设计空间。

## 一、全球顶级包管理系统的设计解剖

我研究了 10 个最成功的包管理系统，提取它们在四个维度上的核心设计智慧：

### 📊 多维对比矩阵

| 系统 | 包格式 | 索引策略 | 安全模型 | 私有源支持 | 速度评级 |
|------|--------|---------|---------|-----------|---------|
| **APT** (Debian) | `.deb` (二进制 + 元数据) | 全量索引 `Packages.gz` | GPG 签名 + Release 文件 | `sources.list` 配置 | ⭐⭐⭐ |
| **DNF** (RHEL) | `.rpm` (二进制 + 元数据) | **SQLite 缓存** `primary.sqlite` | GPG + repo metadata 签名 | `.repo` 文件配置 | ⭐⭐⭐⭐ |
| **Homebrew** | Ruby Formula (源码描述) | **Git 仓库** + JSON API | Bottle SHA-256 校验 | **Tap 机制** (任意 Git 仓库) | ⭐⭐⭐ |
| **Scoop** (Windows) | **JSON manifest** | Git 仓库 (bucket) | SHA-256 / SHA-512 | 自定义 bucket (Git 仓库) | ⭐⭐⭐⭐ |
| **Cargo** (Rust) | `Cargo.toml` (源码描述) | **Sparse HTTP Index** (按需获取) | SHA-256 + API token | `[registries]` 配置 | ⭐⭐⭐⭐⭐ |
| **OCI/ORAS** | OCI Manifest (通用) | Registry API (content-addressable) | **Sigstore/cosign** + SHA-256 | 任意 OCI 兼容 Registry | ⭐⭐⭐⭐ |
| **Nix** | `.nix` derivation (函数式描述) | Nixpkgs Git monorepo | 内容寻址 (CA) Hash | Flake inputs | ⭐⭐ |
| **mise/aqua** | **YAML manifest** | Git 仓库 + 编译缓存 | GitHub Release checksum | 自定义 registry URL | ⭐⭐⭐⭐ |
| **Go modules** | `go.mod` | **GOPROXY 协议** (简单 HTTP) | `go.sum` 透明日志 | `GOPRIVATE` + `GONOSUMDB` | ⭐⭐⭐⭐⭐ |
| **npm** | `package.json` | CouchDB REST API | `npm audit` + ECDSA 签名 | `.npmrc` registry URL | ⭐⭐⭐ |

### 🏆 每个维度的冠军

#### 1. 通用性冠军：OCI (Open Container Initiative)

- ✅ 最初为 Docker 容器设计，但 v1.1+ 变成了"通用分发协议"
- ✅ 基础设施遍布全球：Docker Hub, GitHub GHCR, AWS ECR, GCP GCR, Harbor
- ⚠️ 概念复杂 (manifest, layer, config blob)，对于轻量"安装地图"杀鸡用牛刀

#### 2. 安全性冠军：TUF (The Update Framework)

- ✅ 专门为软件分发安全设计，即使 Registry 被入侵 或 签名密钥泄露 也能防御
- ✅ 防御回滚攻击、冻结攻击、混合攻击等
- ⚠️ 实现复杂度高，对小型项目来说过于沉重

#### 3. 扩展性冠军：Homebrew Tap + aqua Registry

- ✅ 任何人都可以创建自己的 "Tap" (Git 仓库)，去中心化分发
- ✅ 格式极简：每个软件一个 YAML/Ruby/JSON 文件
- ⚠️ Git 仓库索引在包数量极大时性能下降

#### 4. 速度冠军：Cargo Sparse Index + Go GOPROXY

- ✅ 不下载全量索引，按需通过极简 HTTP 获取包信息
- ✅ 可层层级联代理，支持 ETag 增量同步
- ⚠️ 首次全量搜索较慢（没有本地完整索引）

---

## 二、核心设计张力 (Design Tensions)

1. **包格式的简单性 vs 表达力**：JSON (极易解析) vs YAML (人类友好) vs Ruby/Nix DSL (任意逻辑)。UniStack 需要声明式，YAML/JSON 足矣。
2. **索引速度 vs 离线可用**：Sparse HTTP (在线最快) vs Git/SQLite 全量 (可离线搜索)。UniStack 需要混合策略：在线托管 + 本地 SQLite 缓存。
3. **安全性 vs 实现成本**：SHA-256 (最简单) vs GPG vs Sigstore vs TUF (最安全)。UniStack 需要从 SHA-256 起步，架构预留升级空间。
4. **通用分发 vs 自有格式**：自定义格式 (完全控制) vs OCI Registry/Nix (生态复用)。UniStack 倾向于轻量级自有架构。

---

## 三、早期研究的 3 个候选方案

### 方案 A：Aqua-Inspired — YAML 蓝图 + SQLite 本地缓存 (综合得分 17)

- **概念**：一个 YAML 文件描述安装地图，Git/HTTP 托管，本地 SQLite 查询。
- **优点**：人类可读，极易扩展，解析快，适合作为轻量级包的载体。
- **缺点**：搜索强依赖本地缓存同步。

### 方案 B：OCI-Native — 复用容器注册表基础设施 (综合得分 17)

- **概念**：将包打包为 OCI Artifact 推送到 Docker Hub / GHCR 等镜像站。
- **优点**：无需自建后端基础设施，原生支持 Sigstore 签名。
- **缺点**：查询慢（无原生搜索），依赖复杂的 ORAS 客户端逻辑。

### 方案 C：Nix-Inspired — 内容寻址 + 函数式不可变 (综合得分 13)

- **概念**：基于哈希的不可变存储，完美实现依赖隔离与去重。
- **优点**：理论最纯粹，完全防篡改。
- **缺点**：学习曲线极其陡峭，基础设施需重头搭建，过于沉重。

---
---

## 🎯 第二部分：最终设计方案 (v3)

> 基于与用户的多轮决策讨论，最终敲定以下基于**压缩包+内置索引**的实用主义方案。

## 一、包格式：免费/付费双轨制

### 1.1 免费版：标准 `.tar.gz`（不发明格式）

```
vim-1.0.0.tar.gz                    # 标准 tar.gz，任何工具都能解压
│
├── package.yml                      # 包元数据
├── defaults/main.yml                # (复用 Ansible Role 结构)
├── vars/
│   ├── debian.yml                   # 系统包管理器 Fallback 映射天然在此
│   ├── alpine.yml
│   └── ...
├── tasks/main.yml                   # 执行入口
├── templates/
└── files/
```

**优势**：零门槛，`tar xzf` 即可查看内容，Ansible 直接执行。完全复用了已经在工作的 `roles/apps/foundation` 目录模式。

---

## 二、包类型：`kind: package` vs `kind: meta`

### 2.1 单软件包 (`kind: package`)

```yaml
# vim-1.0.0/package.yml
apiVersion: "v1"
kind: "package"                     # ⭐ 单软件包
name: "vim"
version: "1.0.0"
description: "Vi IMproved text editor"
delivery_mode: "native"
tier: "free"
platforms:
  supported: [debian, alpine, redhat, darwin]
```

### 2.2 元包 (`kind: meta`)

类似 `build-essential` 这种包含几十个软件的"场景包"（如 foundation）。

```yaml
# foundation-1.0.0/package.yml
apiVersion: "v1"
kind: "meta"                        # ⭐ 元包
name: "foundation"
version: "1.0.0"
description: "System baseline: 70+ essential packages"

# 核心：声明式依赖列表
packages:
  - name: "vim"
  - name: "curl"
  - name: "xz"
    version: ">=5.0"                 # 版本约束
  - name: "nano"
    optional: true                   # 可选包

has_tasks: true                      # 表明自带额外配置逻辑(如 sudo/sshd 配置)
```

**元包安装流程**：

1. **依赖解析**：检查 `packages` 列表，在注册表中查找对应的单软件包。
2. **批量安装**：如果是 native 包则合并为一条指令执行，archive 则并行下载。
3. **自身逻辑**：执行元包自身的 `tasks/main.yml`。

---

## 三、仓库建设：UniStack 自身集成

**同一个 `unistack` 二进制，既是客户端也是仓库构建器。**

### 3.1 仓库初始化与管理

```bash
# 1. 初始化空仓库
unistack repo init /opt/my-registry

# 2. 添加包 (自动校验格式、计算哈希并放入正确目录)
unistack repo add /opt/my-registry vim-1.0.0.tar.gz

# 3. 构建索引 (生成 SQLite 数据库 packages.db)
unistack repo build /opt/my-registry
```

### 3.2 两种托管方式

**方式 A：静态托管** (最低成本)
直接用 Nginx 托管 `/opt/my-registry` 目录，或 `rsync` 到 S3/GitHub Pages。由于有了预先构建的 SQLite 索引，静态服务器也能实现高级搜索。

**方式 B：UniStack 内置服务**

```bash
unistack repo serve /opt/my-registry --port 8080
```

启动自带的轻量 HTTP 服务器，支持 ETag 增量同步、私有库 BasicAuth/Bearer 认证。

---

## 四、仓库目录结构与 URL 规则

```
https://registry.unistack.dev/              # 仓库根 URL
│
├── repodata/                                # 索引目录
│   ├── packages.db                          # DNF-style SQLite 索引
│   ├── packages.db.zst                      # Zstandard 压缩版 (供客户端同步)
│   └── repomd.json                          # 仓库元数据
│
└── packages/                                # 按首字母组织
    ├── a/
    │   └── aria2/aria2-1.0.0.tar.gz
    ├── f/
    │   └── foundation/foundation-1.0.0.tar.gz  # 元包
    └── r/
        └── redis/redis-1.0.0.uspkg             # 付费包
```

**下载链接** = `仓库 URL` + SQLite 中的 `relative_path`

---

## 五、安全升级路径

架构设计保证向后兼容，旧客户端自动忽略无法识别的新安全字段：

1. **v1 (MVP)**：`SHA-256` checksum（SQLite 索引存储 + 客户端下载后比对）。
2. **v2**：`GPG` 签名包与索引。
3. **v3**：`Sigstore cosign` OIDC 无密钥签名与透明日志审计。
4. **v4**：`TUF (The Update Framework)` 防御冻结、回滚等高阶供应链攻击。

---

## 六、内嵌策略 (`go:embed`)

在 Go 代码中：

```go
// internal/registry/embedded.go
// Top 50 核心基础包直接编译进二进制，实现离线、开箱即用的极速安装体验。
// 仅嵌入元数据和轻量配置，增加体积 < 500KB。
//
//go:embed builtin/vim builtin/curl builtin/foundation ...
var BuiltinPackages embed.FS
```

其余海量长尾软件包，则通过 `unistack update` 同步云端索引后在线获取。

---

## 七、架构答疑 (Q&A)

> **Q1：软件包都是 `.tar.gz`，我没办法进行合法性验证。它到底是我们的安装包，还是普通的压缩包？**

**修正决策**：按照你的最新指示，我们将**严格保持免费版与付费版的边界**：

1. **免费版（开源生态）**：坚持使用标准的 `.tar.gz`。**合法性验证无需在文件内部做文章**，而是完全依赖 Registry 索引（`packages.db`）中提供的 `SHA-256` 哈希值进行比对。如果哈希一致，即证明是官方出品的合法包。
2. **付费版（商业生态）**：使用自研发明的加密格式（如 `.uspkg`）。该格式的内部结构**闭源、不公开**，仅供付费用户使用，并且通过特定的闭源逻辑在内存中解密运行。

> **Q2：Ansible 有发布渠道的包（Galaxy）是什么格式的，他们不满足我们的要求吧？**

**回答**：Ansible Galaxy 发布的内容（Roles 和 Collections）本质上也是普通 `.tar.gz`（或者仅仅是个 Git Repo）。它们**完全无法满足 UniStack 的分发要求**：

1. **缺乏 OS 原生路由**：Galaxy Role 往往通过写死大量 `when: ansible_os_family == 'Debian'` 来分支，导致代码冗长。我们设计的 `vars/{{ platform }}.yml` Fallback 机制要优雅得多。
2. **缺乏多交付模式**：Galaxy 只懂跑脚本。我们的包通过 `delivery_mode` (native/archive/container) 能够让 Go 层智能选择是调用系统包管理器、直接下载二进制，还是起容器。
3. **缺乏本地高速索引**：Galaxy 的客户端查询非常慢（严重依赖全量 API 网络请求）。我们的 DNF 模式（SQLite 客户端索引）能实现毫秒级的离线和在线搜索。

## 八、外围生态扩展与自动化出包体系 (Harvester Ecosystem)

为了解决初期包管理生态“冷启动”的窘境，我们构想了独立于 UniStack 核心之外的辅助生态项目：**`unistack-harvester` (自动收割机)**。

### 8.1 “借用规则，沉淀数据”的顶层战略

利用业内最权威的跨平台包元数据项目 **Repology (repology.org)** 作为初始“映射字典（Rosetta Stone）”。

1. **防“卡脖子”设计**：绝不将 Repology API 设为运行时依赖。我们采用周期性下载其最终结果数据（Data Dumps）或直接解析官方源码（如 Debian `Packages`, Alpine `APKINDEX`）的方式，在本地独立生成属于 UniStack 的 `synonyms.yml` 字典。
2. **法务与开源合规 (Clean Room)**：Repology 的源码规则遵循 GPL。我们通过“仅提取运算后的事实数据（事实不受版权限制）”并以 MIT / CC-BY-SA 协议开源我们独立编写的解析器脚本，完美规避了 GPL 传染问题。这确保了作为闭源商业引擎的 `UniStack CLI` 永远处于 100% 的安全和自由地带。

### 8.2 降低海量包自动化测试的 CI/CD 成本

1. **对于原生包 (Native, apt/apk)**：采取**“延迟绑定与架构甩锅”**策略。CI/CD 只需在最廉价的 `amd64` 环境下验证，不做多架构（ARM64/RISC-V）测试。架构检查完全交由终端用户的原生包管理器（如 apt）在运行时自行判断和阻断，以此节省 80% 的云端测试成本。
2. **对于静态下载包 (Archive)**：使用极低成本的 `HTTP HEAD` 探测（秒级）来替代高昂的物理容器模拟测试，从而精准得出 `supported_archs` 结论。

### 8.3 生态演进三步走路线

* **Phase 1（当前最优解）**：写极简 Go 脚本，纯解析外部 JSON 数据 Dump 并组装 `package.yml`。
* **Phase 2（Server-side Isolation）**：若需自行运行复杂的 Python/数据库解析引擎，将其强制隔离在 CI/CD (Docker 容器) 中执行，产出干净数据后销毁环境。彻底阻绝用户端 Go 项目引入 Python 环境依赖。
* **Phase 3（原生规则演进）**：若项目做大做强，再考虑自研基于纯 Go 原生的正则表达式和 YAML 规则引擎（难度高且目前优先级最低）。

## 九、多源架构与生态安全设计 (Multi-Registry & Security Architecture)

基于对大规模协作生态的研究，我们最终确立了**“Git 协作源码，HTTP 分发产物”**的双轨制隔离架构，并引入了以下企业级安全与生态设计：

### 9.1 强制命名空间 (Strict Namespacing) 与 防混淆攻击 (Dependency Confusion)

为了彻底粉碎同名高版本覆盖的“依赖混淆攻击”，我们制定了极其严苛的命名红线：

1. **Core 核心包**：独占全局顶层命名空间，**绝对不允许**带有任何前缀（例如 `vim`，`nginx`）。
2. **Community / 第三方包**：**必须**带有作者或组织的命名空间前缀（例如 `community/vim`，`snowdreamtech/nginx`）。
在打包构建阶段（`unistack repo build`），如果发现第三方仓库试图提交无前缀的核心包，构建程序将直接抛出致命错误拒绝打包。

### 9.2 优先级锁 (Priority Pinning)

在客户端配置文件（如 `~/.unistack/config.yaml`）中，每个源都必须分配 `priority`：

```yaml
registries:
  - name: "core"
    url: "https://core.unistack.org"
    priority: 10          # 数字越小，优先级越高，绝不被覆盖
  - name: "community"
    url: "https://community.unistack.org"
    priority: 20
```

客户端查询时严格按照优先级顺序进行。哪怕社区库存在一个版本号极其离谱的恶意包（如 `v99.9.9`），只要在最高优先级的库中找到了同名（或同前缀）的包，搜索立即终止，从而保证官方防线的绝对安全。

### 9.3 纯客户端的多源镜像与物理隔离

* **镜像天然支持**：由于我们抛弃了后端 API，转而采用纯静态的分发模式（DNF 模式）。CDN 厂商或开源镜像站只需要原封不动地同步（`rsync`）静态文件即可。用户通过修改配置文件中的 `url`，即可享受千兆极速下载。
* **本地沙盒化索引**：为了防止多个源的 `packages.db` 在用户本地互相覆盖打架，客户端在下载时会利用配置的 `name` 字段作为本地文件名（如保存为 `core.db`, `community.db`）。本地的所有 SQLite 查询将在这些被物理隔离的副本上联合执行，彻底避免了文件冲突。

### 9.4 克制的协议与分支管理

* **废弃 Edge 分支**：与 Alpine 和 Debian 不同，UniStack 聚焦于提供“绝对稳定的跨平台环境”。为了避免系统极度臃肿，官方 Core 源**仅包含主干稳定版 (Stable) 软件**。所有追求激进更新（Edge / Testing）的需求都被转移至社区源由极客自行管理。
* **精简网络协议**：系统坚决只支持 `http(s)://` 和 `file://` 协议，主动砍掉了 `ftp://` 和 `git://` 下载。依靠强大的操作系统原生挂载能力，用户通过配置 `file:///mnt/nfs_share/` 即可实现全离线断网和企业级 NFS/SMB 私有化部署，不仅规避了 SQLite 在网络驱动器上的文件锁崩溃灾难，更将 Go 核心代码的依赖和体积缩减到了极致。
