# Research: Ansible Execution Hook

## Needs Clarification & Technical Decisions

### 1. How to invoke `ansible-playbook` correctly on the local machine?

**Decision**: Use `os/exec` to run `ansible-playbook -i localhost, -c local <path-to-app_loader.yml> -e "app_source_path=<package-path>"`.
**Rationale**: By explicitly forcing `-i localhost, -c local`, we ensure Ansible runs against the current machine without trying to SSH or load complicated global inventories. We pass `app_source_path` as an extra variable (`-e`) so the playbook knows where the package contents reside.
**Alternatives considered**: Using Ansible Go bindings, but they are heavy and often wrap CLI calls anyway. Since we require zero CGO, `os/exec` is the most reliable approach.

### 2. How to handle `stdout` and `stderr` streams?

**Decision**: Attach `cmd.Stdout = os.Stdout` and `cmd.Stderr = os.Stderr`.
**Rationale**: This streams Ansible's output directly to the user's terminal in real-time, which is crucial for long-running installations. Users will see exactly what tasks are being performed or why they failed.
**Alternatives considered**: Buffering the output and printing it at the end. Rejected because it provides a poor UX for long tasks (user might think the CLI is hung).

### 3. How to detect if Ansible is needed?

**Decision**: In `internal/client/installer.go`, after `ExtractTarGz`, check if `app_loader.yml` exists in the extracted package directory using `os.Stat()`. If it exists, execute it. Otherwise, assume it's a pure Go/binary package and finish successfully.
**Rationale**: Simple, zero-configuration detection that relies purely on conventions (convention over configuration).
