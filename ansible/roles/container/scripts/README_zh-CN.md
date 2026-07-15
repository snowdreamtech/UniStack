# 镜像 UID/GID 知识库维护工具 - 跨平台脚本说明

## 📁 脚本文件清单

本目录包含三套完整的脚本，每套脚本都有 **Shell、CMD 和 PowerShell** 三种版本：

### 1. 镜像 UID/GID 探测工具

| 平台 | 脚本文件 | 使用方式 |
|------|---------|----------|
| **macOS/Linux/WSL** | `detect_image_uid.sh` | `./detect_image_uid.sh mongo:7.0` |
| **Windows (CMD)** | `detect_image_uid.cmd` | `detect_image_uid.cmd mongo:7.0` |
| **Windows/全平台 (PowerShell)** | `detect_image_uid.ps1` | `.\detect_image_uid.ps1 mongo:7.0` |

### 2. 知识库验证工具

| 平台 | 脚本文件 | 使用方式 |
|------|---------|----------|
| **macOS/Linux/WSL** | `validate_image_database.sh` | `./validate_image_database.sh` |
| **Windows (CMD)** | `validate_image_database.cmd` | `validate_image_database.cmd` |
| **Windows/全平台 (PowerShell)** | `validate_image_database.ps1` | `.\validate_image_database.ps1` |

### 3. 知识库条目生成工具

| 平台 | 脚本文件 | 使用方式 |
|------|---------|----------|
| **macOS/Linux/WSL** | `generate_kb_entry.sh` | `./generate_kb_entry.sh mongo:7.0 999 999` |
| **Windows (CMD)** | `generate_kb_entry.cmd` | `generate_kb_entry.cmd mongo:7.0 999 999` |
| **Windows/全平台 (PowerShell)** | `generate_kb_entry.ps1` | `.\generate_kb_entry.ps1 mongo:7.0 999 999` |

---

## 🚀 快速开始

### macOS / Linux / WSL

```bash
cd /path/to/ansible/roles/container/scripts

# 探测镜像
./detect_image_uid.sh mongo:7.0 postgres:16

# 验证知识库
./validate_image_database.sh

# 生成条目
./generate_kb_entry.sh kafka:latest 1000 1000
```

### Windows (命令提示符)

```cmd
cd C:\path\to\ansible\roles\container\scripts

REM 探测镜像
detect_image_uid.cmd mongo:7.0 postgres:16

REM 验证知识库
validate_image_database.cmd

REM 生成条目
generate_kb_entry.cmd kafka:latest 1000 1000
```

### Windows (PowerShell)

```powershell
cd C:\path\to\ansible\roles\container\scripts

# 探测镜像
.\detect_image_uid.ps1 mongo:7.0 postgres:16

# 验证知识库
.\validate_image_database.ps1

# 生成条目
.\generate_kb_entry.ps1 kafka:latest 1000 1000
```

### PowerShell Core (跨平台)

```powershell
# 在 macOS/Linux/Windows 上使用 PowerShell Core (pwsh)
pwsh detect_image_uid.ps1 mongo:7.0
pwsh validate_image_database.ps1
pwsh generate_kb_entry.ps1 kafka:latest 1000 1000
```

---

## 💡 选择指南

### 如何选择适合你的脚本版本？

| 场景 | 推荐版本 | 原因 |
|------|---------|------|
| **macOS 用户** | `.sh` (Shell) | 原生支持，性能最佳 |
| **Linux 用户** | `.sh` (Shell) | 原生支持，POSIX 兼容 |
| **Windows 用户（传统）** | `.cmd` (批处理) | 无需额外工具，兼容性好 |
| **Windows 用户（现代）** | `.ps1` (PowerShell) | 功能强大，更好的错误处理 |
| **跨平台团队** | `.ps1` (PowerShell Core) | 统一体验，跨平台运行 |
| **CI/CD 环境** | `.sh` (Shell) | 最广泛的支持 |
| **WSL 用户** | `.sh` (Shell) | 原生 Linux 环境 |
| **Git Bash 用户** | `.sh` (Shell) | 完美支持 POSIX sh |

---

## 🔧 技术特性

### Shell 脚本 (`.sh`)

- ✅ **POSIX sh 兼容** - 不依赖 Bash
- ✅ **轻量高效** - 最小依赖
- ✅ **跨平台** - macOS/Linux/WSL/Git Bash
- ✅ **智能颜色** - 自动检测终端支持

### CMD 批处理 (`.cmd`)

- ✅ **Windows 原生** - 无需额外安装
- ✅ **广泛兼容** - Windows XP 到 Windows 11
- ✅ **简单直接** - 批处理语法
- ⚠️ **功能有限** - 错误处理较弱

### PowerShell (`.ps1`)

- ✅ **现代化** - 强大的对象处理
- ✅ **跨平台** - PowerShell Core 7+
- ✅ **错误处理** - 详细的异常信息
- ✅ **丰富输出** - 彩色格式化
- ⚠️ **执行策略** - 可能需要 `Set-ExecutionPolicy`

---

## ⚙️ 系统要求

### 所有版本通用要求

- Docker 已安装并运行
- 网络连接（用于拉取镜像）

### Shell 脚本 (`.sh`)

- **macOS**: 内置 `/bin/sh`
- **Linux**: POSIX sh (dash/bash/ash)
- **Windows**: Git Bash / WSL / MSYS2

### CMD 批处理 (`.cmd`)

- **Windows**: XP / Vista / 7 / 8 / 10 / 11
- **命令提示符**: `cmd.exe`

### PowerShell (`.ps1`)

- **Windows**: PowerShell 5.1+ (内置于 Windows 10+)
- **跨平台**: PowerShell Core 7+ (需单独安装)
- **macOS/Linux**: `brew install powershell` 或 `snap install powershell`

---

## 🛡️ PowerShell 执行策略

如果在 Windows 上运行 `.ps1` 脚本时遇到权限问题：

```powershell
# 查看当前执行策略
Get-ExecutionPolicy

# 临时允许当前会话运行脚本
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass

# 或者永久允许当前用户运行脚本（推荐）
Set-ExecutionPolicy -Scope CurrentUser -ExecutionPolicy RemoteSigned

# 然后运行脚本
.\detect_image_uid.ps1 mongo:7.0
```

---

## 📊 功能对比

| 功能 | Shell (.sh) | CMD (.cmd) | PowerShell (.ps1) |
|------|-------------|------------|-------------------|
| **跨平台支持** | ✅ 优秀 | ❌ 仅 Windows | ✅ 良好 (需 PS Core) |
| **性能** | ✅ 最快 | ⚠️ 中等 | ⚠️ 稍慢 |
| **错误处理** | ✅ 良好 | ⚠️ 基础 | ✅ 优秀 |
| **彩色输出** | ✅ 支持 | ⚠️ 有限 | ✅ 丰富 |
| **调试能力** | ✅ 良好 | ⚠️ 基础 | ✅ 优秀 |
| **代码可读性** | ✅ 高 | ⚠️ 中 | ✅ 高 |
| **学习曲线** | ⚠️ 中等 | ✅ 简单 | ⚠️ 中等 |

---

## 🧪 测试验证

### 快速测试（所有平台）

```bash
# macOS/Linux/WSL
./detect_image_uid.sh --help
./validate_image_database.sh
./generate_kb_entry.sh alpine:latest 0 0

# Windows CMD
detect_image_uid.cmd
validate_image_database.cmd
generate_kb_entry.cmd alpine:latest 0 0

# Windows PowerShell
.\detect_image_uid.ps1 -?
.\validate_image_database.ps1
.\generate_kb_entry.ps1 alpine:latest 0 0
```

---

## 📝 注意事项

### Windows 路径问题

- **CMD**: 使用反斜杠 `\` (如 `C:\path\to\file`)
- **PowerShell**: 支持正斜杠 `/` 和反斜杠 `\`
- **Shell (Git Bash/WSL)**: 使用正斜杠 `/` (如 `/c/path/to/file`)

### 换行符问题

- **Windows**: CRLF (`\r\n`)
- **macOS/Linux**: LF (`\n`)
- 所有脚本已处理跨平台换行符兼容性

### Docker Desktop for Windows

- CMD 和 PowerShell 脚本需要 Docker Desktop 运行
- WSL 用户可以使用 Shell 脚本访问 Docker

---

## 🎯 最佳实践

1. **优先使用原生脚本**
   - macOS/Linux → `.sh`
   - Windows → `.ps1` 或 `.cmd`

2. **CI/CD 环境**
   - 优先使用 `.sh` (最广泛支持)
   - 或使用 PowerShell Core (统一体验)

3. **团队协作**
   - 统一使用一种脚本格式
   - 在 README 中说明使用方式

4. **版本控制**
   - `.sh` 文件必须有执行权限 (`chmod +x`)
   - `.ps1` 和 `.cmd` 不需要特殊权限

---

## 💬 获取帮助

所有脚本都支持帮助信息：

```bash
# Shell
./detect_image_uid.sh          # 无参数显示帮助

# CMD
detect_image_uid.cmd           # 无参数显示帮助

# PowerShell
.\detect_image_uid.ps1 -?      # -? 参数显示帮助
Get-Help .\detect_image_uid.ps1  # PowerShell 内置帮助
```

---

**维护者**: Infrastructure Team
**最后更新**: 2024-01-24
**版本**: 1.0.0
