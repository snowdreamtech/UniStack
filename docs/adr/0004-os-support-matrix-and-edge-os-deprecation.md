# ADR 0004: 操作系统支持矩阵与边缘生态剥离 (OS Support Matrix & Edge OS Deprecation)

## 背景 (Context)

在 UniStack 的持续迭代过程中，为了追求极致的覆盖率，我们一度尝试兼容诸如 Mageia (`urpmi`)、VMware Photon OS (`tdnf`)、NixOS (`nix-env`) 以及 OpenWrt (`opkg`) 等边缘操作系统。
然而，这些“未正式支持 / 随缘支持”的系统从根基上不符合标准配置管理范式，或者缺乏活跃的维护生态。这导致了大量晦涩的 Shell 补救措施 (Workarounds) 和条件分支，引发了代码圈复杂度的暴增，并且常常导致 CI/CD 流程大面积飘红，严重污染了主干逻辑的纯粹性。

## 决策 (Decision)

为了保证 UniStack 底座的极简、绝对稳定与架构正确性，我们决定**彻底剥离边缘系统支持**，并遵循 Ansible 官方支持金字塔，建立 UniStack 严格的生态分级策略（OS Support Matrix）。

### 官方背书与分级依据

根据红帽（Ansible 母公司）的官方定义，Ansible 对操作系统的支持被划分为明显的两个层级：

#### 1. 核心级 / 企业级支持 (Core & Enterprise Supported)
这些是 Ansible 官方核心团队投入重金进行完整 CI/CD 自动化测试，甚至提供商业级 SLA 保证的系统。Ansible 的内置模块（如 `ansible.builtin.package`, `service`, `user` 等）在这些系统上的表现可以说是完美无缺。
* **Red Hat 家族**：RHEL (红帽企业版), Fedora, CentOS (以及其继任者 Rocky Linux, AlmaLinux)。
* **Debian 家族**：Ubuntu (LTS 及非 LTS), Debian。
* **SUSE 家族**：SLES (SUSE Linux Enterprise Server), openSUSE (Leap, Tumbleweed)。
* *(特殊非 Linux)*：Windows (主要基于 WinRM 和 PowerShell)。

> 💡 **架构启示**：这也就是为什么我们的脚本在这三大派系上运行得如丝般顺滑，根本不需要加任何 Workaround 补丁。

#### 2. 社区级支持 (Community Supported)
这些系统 Ansible 官方不做商业担保，但是因为用户群体庞大，社区开发者积极维护了大量的对应模块。官方通常会合并这些代码，但如果有 Bug，需要社区自己修。
* **极简/容器系**：Alpine Linux (由于 Docker 的流行，支持度极高，几乎等同于核心支持)。
* **激进/极客系**：Arch Linux (`pacman`)、Gentoo (`portage`/`emerge`)、Void Linux。
* **BSD 家族**：FreeBSD, OpenBSD, macOS。
* **网络设备**：Cisco, Juniper 等（专门的网络模块）。

#### 3. 未正式支持 / 随缘支持 (现已弃用)
它们连社区模块（Community Collections）都很少或者年久失修。
* **Mageia (`urpmi`)**：曾经有模块，但维护者跑路，模块经常报错，连 Python 数组处理都有 Bug。
* **VMware Photon OS (`tdnf`)**：极小众的闭源生态产物，Ansible 在判断其类型时常引发逻辑混乱。
* **NixOS**：纯声明式系统，它从根基上就否定了 Ansible 这种“渐进式修改状态”的运维哲学，两者水火不容。
* **OpenWrt**：嵌入式路由系统，官方不建议用 Ansible 这种“重型” Python 工具去管理几兆内存的路由器，哪怕用也是走 Raw 模块写 Shell。

---

## UniStack 架构分级策略

结合上述依据，UniStack 遵循以下金字塔设计原则：

1. **绝对主攻**：**Ubuntu/Debian、RHEL/Rocky/Fedora、SUSE、Alpine**。这四个是公有云和企业私有云的绝对霸主。让 UniStack 在它们身上做到 100% 零补丁、零报错。
2. **社区兼容**：**Arch、Gentoo、Void、FreeBSD、macOS**。我们提供支持，但仅限于 Ansible 标准模块能做到什么程度我们就做到什么程度，**绝不为了它们去污染主线逻辑写一堆特定的 Shell Bypass**。
3. **果断抛弃**：全面清理 Mageia、Photon OS、NixOS、OpenWrt 这种会引发代码复杂度暴增的非主流包袱，实现底层的高内聚和低耦合。

## 影响 (Consequences)

- Go 启动器：移除 `tdnf`, `nix-env`, `opkg`, `urpmi`，实现对于不支持包管理器的**防御性失败 (Fail Fast)**。
- Ansible：移除了为了适配特殊环境的所有冗长脚本判断，恢复原生的 `ansible.builtin.package` 调用，极大地提高了代码质量和可读性。
- CI/CD：精简了测试矩阵，保证每一次 PR 测试都聚焦于核心受支持的主干系统。
