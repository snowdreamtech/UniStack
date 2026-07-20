# Implementation Plan: Ansible Execution Hook

**Branch**: `005-ansible-execution-hook` | **Date**: 2026-07-19 | **Spec**: [spec.md](file:///Users/snowdream/Workspace/snowdreamtech/UniStack/specs/005-ansible-execution-hook/spec.md)

**Input**: Feature specification from `/specs/005-ansible-execution-hook/spec.md`

## Summary

实现当包中存在 `app_loader.yml` 配置文件时，`unistack install` 能够在完成文件解压后，自动使用 `os/exec` 拉起本地环境的 `ansible-playbook` 命令执行该剧本，将包解压绝对路径作为参数传递，并将输出流实时显示给用户。

## Technical Context

**Language/Version**: Go 1.26.4

**Primary Dependencies**: None (Standard Library `os/exec` and `os.Stat`)

**Storage**: Local File System (checking for `app_loader.yml`)

**Testing**: Go standard testing (`testing` package), manually testing with CLI.

**Target Platform**: Any platform running UniStack (macOS/Linux) that has `ansible` installed.

**Project Type**: CLI tool

**Performance Goals**: N/A (Ansible takes however long it takes).

**Constraints**: Must strictly rely on the existence of `app_loader.yml` in the extracted directory to trigger Ansible.

**Scale/Scope**: Modifying the `Install` method in `internal/client/installer.go`.

## Constitution Check

*GATE: Passed*
This adheres to our strategy of pushing heavy logic to Ansible while Go remains the fast wrapper/distributor. Zero CGO is maintained by using `os/exec`.

## Project Structure

### Documentation (this feature)

```text
specs/005-ansible-execution-hook/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/
└── client/
    ├── installer.go
    └── installer_test.go
```

**Structure Decision**: Modifying existing `installer.go` inside `internal/client`.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A       | N/A        | N/A                                 |
