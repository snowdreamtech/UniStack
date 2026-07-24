# UniStack

[![CI 流水线](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniStack/ci.yml?branch=main&label=CI%20Pipeline)](https://github.com/snowdreamtech/UniStack/actions/workflows/ci.yml)
[![CD 自动化发布](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniStack/cd.yml?branch=main&label=CD%20Pipeline)](https://github.com/snowdreamtech/UniStack/actions/workflows/cd.yml)
[![文档站点](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniStack/pages.yml?branch=main&label=%E6%96%87%E6%A1%A3&logo=github)](https://github.com/snowdreamtech/UniStack/actions/workflows/pages.yml)
[![CodeQL 审计](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniStack/codeql.yml?branch=main&label=CodeQL&logo=github)](https://github.com/snowdreamtech/UniStack/actions/workflows/codeql.yml)
[![跨平台验证](https://img.shields.io/badge/Verified-Linux%20%7C%20macOS%20%7C%20Windows-blue)](https://github.com/snowdreamtech/UniStack/actions/workflows/ci.yml)
[![安全审计](https://img.shields.io/badge/Security-Zizmor%20%7C%20Trivy%20%7C%20Gitleaks-brightgreen)](https://github.com/snowdreamtech/UniStack/actions/workflows/ci.yml)
[![SBOM 背书](https://img.shields.io/badge/SBOM-Available-success)](https://github.com/snowdreamtech/UniStack/releases/latest)
[![许可证: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/license/MIT)
[![最新发布](https://img.shields.io/github/v/release/snowdreamtech/UniStack?logo=github&sort=semver)](https://github.com/snowdreamtech/UniStack/releases/latest)
[![Dependabot 已启用](https://img.shields.io/badge/Dependabot-Enabled-brightgreen?logo=dependabot)](https://github.com/snowdreamtech/UniStack/blob/main/.github/dependabot.yml)
[![pre-commit 已启用](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit)](https://github.com/pre-commit/pre-commit)
[![GitHub Stars](https://img.shields.io/github/stars/snowdreamtech/UniStack?style=social)](https://github.com/snowdreamtech/UniStack)
[![GitHub Issues](https://img.shields.io/github/issues/snowdreamtech/UniStack)](https://github.com/snowdreamtech/UniStack/issues)
[![代码规模](https://img.shields.io/github/languages/code-size/snowdreamtech/UniStack)](https://github.com/snowdreamtech/UniStack)

[English](README.md) | [简体中文](README_zh-CN.md)

UniStack 是一个快速、灵活且企业级的 Golang CLI 模板，深受 UniRTM 和 helloworld 的启发。它为构建现代 Go 命令行应用程序提供了一个坚实的基础，内置了精美的终端输出、结构化日志、缓存管理机制以及全套的 CI/CD 流水线。

## 🌟 特性

- **现代 CLI 架构**：基于 Go 1.24+ 和 Cobra 构建，提供结构化且易于扩展的命令行框架。
- **精美输出**：使用 `pterm` 实现丰富、多彩且交互友好的终端输出效果。
- **结构化日志**：内置对 Go 标准库 `slog` 的支持，提供高性能的结构化日志记录。
- **丰富的内置工具**：自带 `doctor`、`self-update`、`cache`、`config` 和 `license` 等实用命令。
- **三重保证质量**：通过 IDE 实时检查、Pre-commit 本地拦截和 GitHub Actions 远端全量审计，构建 100% 代码纯净度防线。
- **跨平台就绪**：在 macOS、Linux 和 Windows 上均可无缝运行。

## 🏗️ 架构与设计

### 概览

UniStack 旨在解决每次启动新的 Go CLI 项目时遇到的模板化和重复配置问题。它提供了开箱即用的标准开发环境、架构模式和自动化流水线。

### 核心组件

- **CLI 框架**：使用 `spf13/cobra` 和 `spf13/viper` 进行强大的命令解析和配置管理。
- **UI 与 日志**：集成了 `pterm` 作为 UI 组件，并使用 `slog` 进行结构化的事件记录。
- **数据库与缓存**：由 SQLite 驱动 (通过 `modernc.org/sqlite`)，支持本地缓存和事务性数据存储。
- **任务编排**：内置了 `UniRTM` 配置，用于管理本地开发工作流（lint、test、verify）。

## 📖 使用指南

### 前置条件

- **运行时**: Go (>= 1.24)。
- **Git**: 需要全局安装 git。
- **UniRTM**: 必须安装以执行任务编排。

### 安装

**通过 NPM 安装**:

```sh-session
npm install -g @snowdreamtech/unigo
```

**通过 PyPI 安装**:

```sh-session
pip install snowdreamtech-unigo
```

### 快速开始

1. **安装 UniRTM**：请确保系统已安装 `unirtm`。
2. **初始化**：`unirtm run setup`（引导安装核心工具与钩子）。
3. **安装依赖**：`unirtm run install`（安装项目依赖）。
4. **验证**：`unirtm run verify`（确保所有代码检查通过）。
5. **构建**：`go build -o unistack main.go`

### 可用命令

- `unistack version`: 打印版本号
- `unistack doctor`: 检查系统健康状况并诊断问题
- `unistack self-update`: 更新到最新版本
- `unistack cache`: 管理本地缓存
- `unistack config`: 管理配置
- `unistack license`: 管理源代码文件中的版权许可头

## 🛠️ 运维指南

### 部署前检查清单

1. 运行 `unirtm run verify` 确保所有质量门禁均为绿色（执行格式化、Lint、测试及安全审计）。
2. 确保 `CHANGELOG.md` 已更新。

### 故障排除

- **问题**: `unirtm run verify` 报测试失败。
  - **解决方案**: 确保代码通过所有 Go 单元测试。UniStack 会执行严格的代码覆盖率和竞态条件 (race) 检查。
- **问题**: Pre-commit 钩子报错。
  - **解决方案**: 许多钩子（如代码格式化）会自动修复问题。将修改后的文件重新 `git add` 并再次提交即可。

## 🔒 安全注意事项

### 安全模型

- **审计日志**: 所有的关键操作与安全扫描，均在使用 `unirtm run verify` 时，由 `trivy`、`gitleaks` 和 `govulncheck` 等工具自动执行。
- **依赖管理**: 由 Dependabot 支持，以确保所有项目依赖项始终保持最新且安全的状态。

## 🧑‍💻 开发者指南

### 本地开发设置

```bash
git clone https://github.com/snowdreamtech/UniStack.git
cd UniStack
unirtm run setup
unirtm run install
```

### 🚀 代理使用场景

`GITHUB_PROXY` (默认: `https://gh-proxy.sn0wdr1am.com/`) 针对特定的网络加速场景进行了优化。在不支持的协议（如 Git）上误用它会导致错误。

| 场景                   | 是否支持      | 示例 / 说明                                    |
| :--------------------- | :------------ | :--------------------------------------------- |
| **Release 文件**       | ✅ 支持       | `.../releases/download/v1.0/tool.zip`          |
| **源码归档 (Archive)** | ✅ 支持       | `.../archive/master.zip` 或 `.tar.gz`          |
| **文件直接链接**       | ✅ 支持       | `.../blob/master/filename`                     |
| **Git Clone**          | ❌ **不支持** | **请勿**用于 `git clone` 或 `insteadOf` 配置。 |
| **项目文件夹**         | ❌ **不支持** | 不支持通过代理进行项目文件夹的浏览或克隆。     |

> [!IMPORTANT]
> 为了防止破坏工具链（如 `unirtm`），本模板显式禁用了通过此代理进行的 Git 重定向。请仅在脚本中进行直接 HTTP 下载时使用它。

## 📄 许可证

本项目采用 **MIT 许可证** 授权。
版权所有 (c) 2026-现在 [SnowdreamTech Inc.](https://github.com/snowdreamtech)
详见 [LICENSE](./LICENSE) 文件。

## Star History

[![Star History Chart](https://api.star-history.com/image?repos=snowdreamtech/UniStack&type=date&legend=top-left)](https://www.star-history.com/?repos=snowdreamtech%2FUniStack&type=date&legend=top-left)
