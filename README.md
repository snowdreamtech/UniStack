# UniGo

[![CI Pipeline](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniGo/ci.yml?branch=main&label=CI%20Pipeline)](https://github.com/snowdreamtech/UniGo/actions/workflows/ci.yml)
[![CD Pipeline](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniGo/cd.yml?branch=main&label=CD%20Pipeline)](https://github.com/snowdreamtech/UniGo/actions/workflows/cd.yml)
[![GitHub Pages](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniGo/pages.yml?branch=main&label=Docs&logo=github)](https://github.com/snowdreamtech/UniGo/actions/workflows/pages.yml)
[![CodeQL](https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniGo/codeql.yml?branch=main&label=CodeQL&logo=github)](https://github.com/snowdreamtech/UniGo/actions/workflows/codeql.yml)
[![Multi-OS Verified](https://img.shields.io/badge/Verified-Linux%20%7C%20macOS%20%7C%20Windows-blue)](https://github.com/snowdreamtech/UniGo/actions/workflows/ci.yml)
[![Security Audit](https://img.shields.io/badge/Security-Zizmor%20%7C%20Trivy%20%7C%20Gitleaks-brightgreen)](https://github.com/snowdreamtech/UniGo/actions/workflows/ci.yml)
[![SBOM Available](https://img.shields.io/badge/SBOM-Available-success)](https://github.com/snowdreamtech/UniGo/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/license/MIT)
[![Release](https://img.shields.io/github/v/release/snowdreamtech/UniGo?logo=github&sort=semver)](https://github.com/snowdreamtech/UniGo/releases/latest)
[![Dependabot Enabled](https://img.shields.io/badge/Dependabot-Enabled-brightgreen?logo=dependabot)](https://github.com/snowdreamtech/UniGo/blob/main/.github/dependabot.yml)
[![pre-commit](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit)](https://github.com/pre-commit/pre-commit)
[![GitHub Stars](https://img.shields.io/github/stars/snowdreamtech/UniGo?style=social)](https://github.com/snowdreamtech/UniGo)
[![GitHub Issues](https://img.shields.io/github/issues/snowdreamtech/UniGo)](https://github.com/snowdreamtech/UniGo/issues)
[![Code Size](https://img.shields.io/github/languages/code-size/snowdreamtech/UniGo)](https://github.com/snowdreamtech/UniGo)

[English](README.md) | [简体中文](README_zh-CN.md)

UniGo is a fast, flexible, and enterprise-grade Golang CLI template inspired by UniRTM and helloworld. It provides a robust foundation for building modern command-line applications in Go, featuring beautiful terminal output, structured logging, built-in caching, and a comprehensive CI/CD pipeline.

## 🌟 Features

- **Modern CLI Architecture**: Built with Go 1.24+ and Cobra, providing a structured and extensible command-line application framework.
- **Beautiful Output**: Utilizes `pterm` for rich, colorful, and interactive terminal output.
- **Structured Logging**: Built-in support for Go's standard `slog` package for high-performance structured logging.
- **Robust Tooling**: Includes built-in commands like `doctor`, `self-update`, `cache`, `config`, and `license`.
- **Triple Guarantee Quality**: 100% code purity enforced through IDE checks, Pre-commit hooks, and GitHub Actions integrated quality gates.
- **Cross-Platform Ready**: Runs seamlessly on macOS, Linux, and Windows.

## 🏗️ Architecture & Design

### Overview

UniGo is engineered to solve the boilerplate problem when starting new Go CLI projects. It standardizes the development environment, architectural patterns, and automation pipelines out of the box.

### Core Components

- **CLI Framework**: Utilizes `spf13/cobra` and `spf13/viper` for powerful command parsing and configuration management.
- **UI & Logging**: Integrated `pterm` for UI components and `slog` for structured event logging.
- **Database & Caching**: SQLite-backed (via `modernc.org/sqlite`) local caching and transactional data storage.
- **Task Orchestration**: Configured with `UniRTM` for managing local development workflows (lint, test, verify).

## 📖 Usage Guide

### Prerequisites

- **Runtime**: Go (>= 1.24).
- **Git**: Global git installation required.
- **UniRTM**: Required for task execution and orchestration.

### Quick Start

1. **Install UniRTM**: Ensure `unirtm` is installed on your system.
2. **Initialize**: `unirtm run setup` (bootstraps core dependencies and hooks).
3. **Install**: `unirtm run install` (installs project dependencies).
4. **Verify**: `unirtm run verify` (ensures everything is green).
5. **Build**: `go build -o unigo main.go`

### Available Commands

- `unigo version`: Print the version number
- `unigo doctor`: Check system health and diagnose issues
- `unigo self-update`: Update to the latest version
- `unigo cache`: Manage local cache
- `unigo config`: Manage configuration
- `unigo license`: Manage copyright license headers

## 🛠️ Operations Guide

### Pre-deployment Checklist

1. Run `unirtm run verify` to ensure all quality gates are green (runs formatters, linters, tests, and security audits).
2. Ensure `CHANGELOG.md` is updated.

### Troubleshooting

- **Problem**: `unirtm run verify` fails with test errors.
  - **Solution**: Ensure your code passes all Go unit tests. UniGo enforces strict coverage and race condition checks.
- **Problem**: Pre-commit hooks fail.
  - **Solution**: The hooks often auto-fix issues (like formatting). Stage the modified files and commit again.

## 🔒 Security Considerations

### Security Model

- **Audit Logging**: All critical operations and security scans are executed during the `verify` task using tools like `trivy`, `gitleaks`, and `govulncheck`.
- **Dependency Management**: Powered by Dependabot to ensure all dependencies are kept up-to-date and secure.

## 🧑‍💻 Development Guide

### Local Development Setup

```bash
git clone https://github.com/snowdreamtech/UniGo.git
cd UniGo
unirtm run setup
unirtm run install
```

### 🚀 Proxy Usage Scenarios

The `GITHUB_PROXY` (default: `https://gh-proxy.sn0wdr1am.com/`) is optimized for specific network acceleration scenarios. Misusing it for unsupported protocols (like Git) will result in errors.

| Scenario              | Supported? | Example / Note                                         |
| :-------------------- | :--------- | :----------------------------------------------------- |
| **Release Files**     | ✅ Yes     | `.../releases/download/v1.0/tool.zip`                  |
| **Source Archives**   | ✅ Yes     | `.../archive/master.zip` or `.tar.gz`                  |
| **Direct File Links** | ✅ Yes     | `.../blob/master/filename`                             |
| **Git Clone**         | ❌ **No**  | Do **not** use for `git clone` or `insteadOf` configs. |
| **Project Folders**   | ❌ **No**  | Browsing/cloning via proxy is not supported.           |

> [!IMPORTANT]
> To prevent breaking toolchains (like `unirtm`), this template explicitly disables Git redirection via this proxy. Use it only for direct HTTP downloads in scripts.

## 📄 License

This project is licensed under the **MIT License**.
Copyright (c) 2026-present [SnowdreamTech Inc.](https://github.com/snowdreamtech)
See the [LICENSE](./LICENSE) file for the full license text.

## Star History

[![Star History Chart](https://api.star-history.com/image?repos=snowdreamtech/UniGo&type=date&legend=top-left)](https://www.star-history.com/?repos=snowdreamtech%2FUniGo&type=date&legend=top-left)
