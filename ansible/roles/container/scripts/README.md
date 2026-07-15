# Image UID/GID Knowledge Database Maintenance Tools - Cross-Platform Scripts Guide

## 📁 Script File List

This directory contains three complete toolsets, each with **Shell, CMD, and PowerShell** versions:

### 1. Image UID/GID Detector

| Platform | Script File | Usage |
|----------|-------------|-------|
| **macOS/Linux/WSL** | `detect_image_uid.sh` | `./detect_image_uid.sh mongo:7.0` |
| **Windows (CMD)** | `detect_image_uid.cmd` | `detect_image_uid.cmd mongo:7.0` |
| **Windows/Cross-platform (PowerShell)** | `detect_image_uid.ps1` | `.\detect_image_uid.ps1 mongo:7.0` |

### 2. Knowledge Database Validator

| Platform | Script File | Usage |
|----------|-------------|-------|
| **macOS/Linux/WSL** | `validate_image_database.sh` | `./validate_image_database.sh` |
| **Windows (CMD)** | `validate_image_database.cmd` | `validate_image_database.cmd` |
| **Windows/Cross-platform (PowerShell)** | `validate_image_database.ps1` | `.\validate_image_database.ps1` |

### 3. Knowledge Database Entry Generator

| Platform | Script File | Usage |
|----------|-------------|-------|
| **macOS/Linux/WSL** | `generate_kb_entry.sh` | `./generate_kb_entry.sh mongo:7.0 999 999` |
| **Windows (CMD)** | `generate_kb_entry.cmd` | `generate_kb_entry.cmd mongo:7.0 999 999` |
| **Windows/Cross-platform (PowerShell)** | `generate_kb_entry.ps1` | `.\generate_kb_entry.ps1 mongo:7.0 999 999` |

---

## 🚀 Quick Start

### macOS / Linux / WSL

```bash
cd /path/to/ansible/roles/container/scripts

# Detect images
./detect_image_uid.sh mongo:7.0 postgres:16

# Validate knowledge database
./validate_image_database.sh

# Generate entry
./generate_kb_entry.sh kafka:latest 1000 1000
```

### Windows (Command Prompt)

```cmd
cd C:\path\to\ansible\roles\container\scripts

REM Detect images
detect_image_uid.cmd mongo:7.0 postgres:16

REM Validate knowledge database
validate_image_database.cmd

REM Generate entry
generate_kb_entry.cmd kafka:latest 1000 1000
```

### Windows (PowerShell)

```powershell
cd C:\path\to\ansible\roles\container\scripts

# Detect images
.\detect_image_uid.ps1 mongo:7.0 postgres:16

# Validate knowledge database
.\validate_image_database.ps1

# Generate entry
.\generate_kb_entry.ps1 kafka:latest 1000 1000
```

### PowerShell Core (Cross-platform)

```powershell
# Use PowerShell Core (pwsh) on macOS/Linux/Windows
pwsh detect_image_uid.ps1 mongo:7.0
pwsh validate_image_database.ps1
pwsh generate_kb_entry.ps1 kafka:latest 1000 1000
```

---

## 💡 Selection Guide

### How to choose the right script version?

| Scenario | Recommended Version | Reason |
|----------|---------------------|--------|
| **macOS Users** | `.sh` (Shell) | Native support, best performance |
| **Linux Users** | `.sh` (Shell) | Native support, POSIX compatible |
| **Windows Users (Traditional)** | `.cmd` (Batch) | No extra tools needed, good compatibility |
| **Windows Users (Modern)** | `.ps1` (PowerShell) | Powerful features, better error handling |
| **Cross-platform Teams** | `.ps1` (PowerShell Core) | Unified experience, runs everywhere |
| **CI/CD Environments** | `.sh` (Shell) | Widest support |
| **WSL Users** | `.sh` (Shell) | Native Linux environment |
| **Git Bash Users** | `.sh` (Shell) | Perfect POSIX sh support |

---

## 🔧 Technical Features

### Shell Scripts (`.sh`)

- ✅ **POSIX sh Compatible** - No Bash dependency
- ✅ **Lightweight & Efficient** - Minimal dependencies
- ✅ **Cross-platform** - macOS/Linux/WSL/Git Bash
- ✅ **Smart Colors** - Auto-detect terminal support

### CMD Batch (`.cmd`)

- ✅ **Windows Native** - No additional installation
- ✅ **Wide Compatibility** - Windows XP to Windows 11
- ✅ **Simple & Direct** - Batch scripting syntax
- ⚠️ **Limited Features** - Weak error handling

### PowerShell (`.ps1`)

- ✅ **Modern** - Powerful object handling
- ✅ **Cross-platform** - PowerShell Core 7+
- ✅ **Error Handling** - Detailed exception information
- ✅ **Rich Output** - Colorful formatting
- ⚠️ **Execution Policy** - May require `Set-ExecutionPolicy`

---

## ⚙️ System Requirements

### Common Requirements for All Versions

- Docker installed and running
- Network connection (for pulling images)

### Shell Scripts (`.sh`)

- **macOS**: Built-in `/bin/sh`
- **Linux**: POSIX sh (dash/bash/ash)
- **Windows**: Git Bash / WSL / MSYS2

### CMD Batch (`.cmd`)

- **Windows**: XP / Vista / 7 / 8 / 10 / 11
- **Command Prompt**: `cmd.exe`

### PowerShell (`.ps1`)

- **Windows**: PowerShell 5.1+ (built-in on Windows 10+)
- **Cross-platform**: PowerShell Core 7+ (requires separate installation)
- **macOS/Linux**: `brew install powershell` or `snap install powershell`

---

## 🛡️ PowerShell Execution Policy

If you encounter permission issues when running `.ps1` scripts on Windows:

```powershell
# Check current execution policy
Get-ExecutionPolicy

# Temporarily allow scripts for current session
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass

# Or permanently allow for current user (recommended)
Set-ExecutionPolicy -Scope CurrentUser -ExecutionPolicy RemoteSigned

# Then run the script
.\detect_image_uid.ps1 mongo:7.0
```

---

## 📊 Feature Comparison

| Feature | Shell (.sh) | CMD (.cmd) | PowerShell (.ps1) |
|---------|-------------|------------|-------------------|
| **Cross-platform Support** | ✅ Excellent | ❌ Windows Only | ✅ Good (requires PS Core) |
| **Performance** | ✅ Fastest | ⚠️ Medium | ⚠️ Slightly Slower |
| **Error Handling** | ✅ Good | ⚠️ Basic | ✅ Excellent |
| **Colored Output** | ✅ Supported | ⚠️ Limited | ✅ Rich |
| **Debugging Capability** | ✅ Good | ⚠️ Basic | ✅ Excellent |
| **Code Readability** | ✅ High | ⚠️ Medium | ✅ High |
| **Learning Curve** | ⚠️ Medium | ✅ Easy | ⚠️ Medium |

---

## 🧪 Testing and Validation

### Quick Test (All Platforms)

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

## 📝 Important Notes

### Windows Path Issues

- **CMD**: Use backslashes `\` (e.g., `C:\path\to\file`)
- **PowerShell**: Supports both forward `/` and backslashes `\`
- **Shell (Git Bash/WSL)**: Use forward slashes `/` (e.g., `/c/path/to/file`)

### Line Ending Issues

- **Windows**: CRLF (`\r\n`)
- **macOS/Linux**: LF (`\n`)
- All scripts handle cross-platform line ending compatibility

### Docker Desktop for Windows

- CMD and PowerShell scripts require Docker Desktop running
- WSL users can use Shell scripts to access Docker

---

## 🎯 Best Practices

1. **Use Native Scripts First**
   - macOS/Linux → `.sh`
   - Windows → `.ps1` or `.cmd`

2. **CI/CD Environments**
   - Prefer `.sh` (widest support)
   - Or use PowerShell Core (unified experience)

3. **Team Collaboration**
   - Standardize on one script format
   - Document usage in README

4. **Version Control**
   - `.sh` files must have executable permissions (`chmod +x`)
   - `.ps1` and `.cmd` don't need special permissions

---

## 💬 Getting Help

All scripts support help information:

```bash
# Shell
./detect_image_uid.sh          # No arguments shows help

# CMD
detect_image_uid.cmd           # No arguments shows help

# PowerShell
.\detect_image_uid.ps1 -?      # -? parameter shows help
Get-Help .\detect_image_uid.ps1  # PowerShell built-in help
```

---

**Maintainer**: Infrastructure Team
**Last Updated**: 2024-01-24
**Version**: 1.0.0
