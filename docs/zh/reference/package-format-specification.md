# 📦 UniStack Package Format Specification

本文档详细定义了 UniStack 生态系统中的软件包格式规范。为了在“极致的开源生态扩展性”与“顶级的商业代码保护”之间取得平衡，UniStack 采用了**免费/付费双轨制格式**。

---

## 1. 免费包格式规范 (Free Package)

免费包主要面向开源社区分发，主打**高兼容性、高透明度与零学习成本**。

### 1.1 物理封装
* **扩展名**：**强制统一为 `.tar.gz`**。
  * *架构考量*：绝不使用 `.zip`。因为 `.tar` 能够完美原生保留 Linux/Unix 的 POSIX 文件权限（如可执行权限 `0755`）以及软链接（Symlinks），这对基础设施和代码部署是致命且必需的。
* **压缩算法**：标准的 gzip。
* **合法性校验与本地脱机安装**：
  * **Registry 模式**：通过 `unistack install <name>` 安装时，合法性由 Registry 索引库（`packages.db`）下发的 `SHA-256` 散列值进行严格比对。
  * **Local 本地模式**：为了支持用户自定义包的安装，当用户通过本地路径安装（如 `unistack install ./my-tool.tar.gz`）时，引擎将自动跳过 Registry 校验，仅在终端打印警告提示（类似 `Warning: Installing untrusted local package`），确保系统保持极致的开放性。

### 1.2 内部目录结构 (Ansible Role 原生兼容)
解压后的根目录即为该包的作用域，其内部结构**严格兼容 Ansible Role 标准目录树**。这使得任何熟悉 Ansible 的开发者都能零门槛制作 UniStack 免费包。

```text
/ (Archive Root)
├── package.yml          # [必选] UniStack 专属元数据清单
├── tasks/
│   └── main.yml         # [必选] 安装与执行的主入口
├── defaults/
│   └── main.yml         # [可选] 默认变量 (低优先级)
├── vars/
│   ├── main.yml         # [可选] 内部变量 (高优先级)
│   ├── debian.yml       # [可选] OS 平台特定的降级映射字典
│   └── alpine.yml
├── files/               # [可选] 需要下发的静态二进制或配置文件
└── templates/           # [可选] 需要动态渲染的 Jinja2 模板 (.j2)
```

### 1.3 元数据规范 (`package.yml`)
为了保证系统未来的极强扩展性、向下兼容性以及健壮处理，`package.yml` 摒弃了扁平化的简单结构，引入了类似 Kubernetes/Helm 的强类型和分层规范。

```yaml
# Schema 规范版本，为未来可能的大幅结构重构预留兼容性接口
apiVersion: "v1alpha1"

# 包类型：
#  - package: 标准的独立功能包（如 vim, nginx）
#  - meta: 元包，作为虚拟依赖聚合器（如 foundation）
kind: "package" 

# 核心元数据
metadata:
  name: "nginx"
  
  # 【配方版本号】(必须是 SemVer)：
  # 仅用于 UniStack 客户端判断“安装脚本”本身是否需要更新升级（例如修复了某个 Ansible bug）。
  version: "1.0.1"
  
  # 【真实软件版本号 / App Version】：
  # 为了彻底消除用户的懵逼感，这里单独剥离出真正的软件版本。
  # 针对 Linux 各大发行版自带版本严重碎片化（如 Debian 是 1.18，Alpine 是 1.24）的现状，
  # 这个字段不仅可以是单字符串，还原生支持数组或 SemVer 范围语法！
  appVersion: 
    - "1.18.x"
    - "1.24.x"
    # 或者写成单一字符串区间：">= 1.18.0, < 1.25.0"
    
  description: "High performance web server"
  authors: ["SnowdreamTech <snowdreamtech@qq.com>"]
  homepage: "https://nginx.org"
  license: "MIT"
  tags: ["web", "infrastructure"]

# 规范化环境支持声明 (精准防错机制)
compatibility:
  # 采用对象数组，精准表达 N×M 的支持关系。
  # 规范：os 和 arch 必须全部强制为【小写】！
  - os: "debian"
    # [可选] 约束底层 OS 的最低大版本号（解决 Debian 12 没有，Debian 13 才有包的问题）
    min: "13"
    # 如果不写 archs，则代表支持该 OS 下的所有架构
    archs: ["amd64", "arm64"]
  - os: "alpine"
    # 假设 Alpine 只有 amd64 的编译包
    archs: ["amd64"]
  - os: "redhat"
    min: "9"

# 依赖声明 (字典结构，保留未来扩展版本号约束的能力)
# ！！！极其重要：这里声明的依赖，全部是指 UniStack 自身生态内的包名，而不是 Linux 底层包！
# 例如依赖了 foundation，引擎会先去 UniStack Registry 下载 foundation 的 .tar.gz 并执行它的 Ansible 任务。
dependencies:
  # 当前版本仅写包名，未来可无缝平滑扩展为 `foundation: { version: ">= 1.0.0" }`
  foundation: {}
  curl: {}
```

### 1.4 元包规范 (Meta Package)
元包（`kind: meta`）是生态系统中用于“一键安装套餐”的特殊规范（例如一键安装 `foundation`，它会自动装上 vim, git, curl 等）。

**元包的特殊约束：**
1. **执行旁路 (Execution Bypass)**：当 UniStack 识别到 `kind: meta` 时，**将不会执行**该包里的 `tasks/main.yml`。它的存在仅仅是为了声明依赖。
2. **依赖展开**：引擎会将元包的 `dependencies` 字段压入安装队列，通过拓扑排序（Topology Sort）先解析并安装所有子包。
3. **极简目录**：元包的 `.tar.gz` 内部通常只有一个干瘪的 `package.yml`，没有庞大的 Ansible Task 和 Files 目录，极其轻量。

```yaml
# foundation/package.yml (元包示例)
apiVersion: "v1alpha1"
kind: "meta"

# 元包与标准包共享完全一致的 metadata 结构！(即使你不填，Schema 的约束依然存在)
metadata:
  name: "foundation"
  version: "1.0.0"
  description: "The core foundation toolchain for all Linux nodes"
  authors: ["SnowdreamTech <snowdreamtech@qq.com>"]
  homepage: "https://github.com/snowdreamtech/unistack"
  license: "MIT"

# 这里的依赖全部是指向 UniStack 生态内的软件包
dependencies:
  vim: {}
  curl: {}
  git: {}
  jq: {}
```

