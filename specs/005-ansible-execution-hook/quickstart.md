# Quickstart Validation: Ansible Execution Hook

This guide outlines how to manually validate that `unistack install` correctly triggers Ansible playbooks when they are present.

## Prerequisites

1. `ansible` must be installed on your local machine (`ansible-playbook` command available).
2. UniStack CLI binary must be compiled and available (`go run ./cmd/unistack/...`).
3. A local test registry must be built (which we already have via `test-hello.yml`).

## Validation Scenarios

### Scenario 1: Installing a Package with an Ansible Playbook

1. Build the local registry containing the `hello` package (which uses Ansible):

   ```bash
   go run ./cmd/unistack/... registry build -d .ansible -o packages.db
   # Compress if required
   zstd --force packages.db -o packages.db.zst
   ```

2. Start the local HTTP registry server:

   ```bash
   python3 -m http.server 8080
   ```

3. Update the client database to point to the local server (in another terminal):

   ```bash
   go run ./cmd/unistack/... update
   ```

4. Install the package:

   ```bash
   go run ./cmd/unistack/... install hello
   ```

5. **Expected Outcome**: You should see the standard extraction logs, immediately followed by the colorful output of `ansible-playbook` running the tasks defined in `hello`'s role (e.g., creating a file, printing a debug message). The command must return exit code 0.

### Scenario 2: Failing Ansible Playbook

1. Create a dummy package with a broken `app_loader.yml` (e.g., invalid syntax or an explicit `fail` module).
2. Attempt to install it.
3. **Expected Outcome**: The `ansible-playbook` output should stream to the terminal, show the failure in red, and the `unistack install` command should gracefully return a non-zero exit code with an error message indicating Ansible execution failed.
