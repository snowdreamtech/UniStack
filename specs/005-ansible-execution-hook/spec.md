# Feature Specification: Ansible Execution Hook

**Feature Branch**: `005-ansible-execution-hook`

**Created**: 2026-07-19

**Status**: Draft

**Input**: User description: "选项 1：🎯 深度 Ansible 执行链路融合 (Ansible Execution Hook)..."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 自动拉起 Ansible 底层装配 (Priority: P1)

作为 UniStack 用户，我希望在使用 `unistack install <复杂的包>` 时，如果该包包含系统级配置要求（Ansible playbook），系统能自动帮我运行底层 Ansible 装配，而不仅仅是简单的文件解压。

**Why this priority**: 这是整个 UniStack "Go处理业务+Ansible处理执行" 核心哲学落地的最关键功能。

**Independent Test**: 可以创建一个携带 `app_loader.yml` 并且执行了某个副作用（比如 touch 一个文件或打印一条信息）的测试包。当执行 `unistack install` 时，能观察到 Ansible playbook 执行的控制台输出，且副作用生效。

**Acceptance Scenarios**:

1. **Given** 目标包存在 `package.yml` 和依赖的 `app_loader.yml` 剧本，**When** 用户触发安装该包并解压完毕后，**Then** 系统自动调用 `ansible-playbook` 传入对应参数，并完成最终装配。
2. **Given** 目标包仅仅是一个普通二进制包（无 playbook），**When** 用户触发安装时，**Then** 系统像以前一样正常跳过 Ansible 环节并成功完成安装。

---

### Edge Cases

- What happens when Ansible is not installed on the user's local machine?
- How does the system handle an Ansible execution failure (non-zero exit code)? Does it rollback or abort?
- How does the user see the Ansible execution logs (streaming stdout/stderr)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST be able to detect the presence of Ansible playbooks/configurations for an extracted package.
- **FR-002**: System MUST use Go's `os/exec` to invoke `ansible-playbook` locally.
- **FR-003**: System MUST pass the correct context parameters (e.g., extracted absolute path) to the `app_loader.yml`.
- **FR-004**: System MUST stream or capture the `stdout`/`stderr` from Ansible so the user knows what is happening during long installations.
- **FR-005**: System MUST properly handle execution errors from Ansible, reporting them back to the user.

### Key Entities

- **Installer Execution Context**: Contains the absolute path of the downloaded package and necessary variables passed to Ansible.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of packages defining an Ansible playbook trigger the Ansible execution correctly.
- **SC-002**: Pure Go packages (without Ansible requirements) experience 0 regression and install just as fast as before.
- **SC-003**: Ansible execution logs are visible to the user during the process.

## Assumptions

- Ansible (e.g., `ansible-playbook` command) is already installed and available in the `$PATH` of the host operating system.
- The user has appropriate permissions (sudo/root if necessary) configured via Ansible if the playbook requires it.
